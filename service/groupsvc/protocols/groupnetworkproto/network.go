package groupnetworkproto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/myprotocol"
	"github.com/jianbo-zh/dchat/internal/mytype"
	logds "github.com/jianbo-zh/dchat/service/groupsvc/datastore/ds/groupadminds"
	ds "github.com/jianbo-zh/dchat/service/groupsvc/datastore/ds/groupnetworkds"
	pb "github.com/jianbo-zh/dchat/service/groupsvc/protobuf/pb/grouppb"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"

	ipfsds "github.com/ipfs/go-datastore"
	logging "github.com/jianbo-zh/go-log"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var log = logging.Logger("cuckoo/groupnetworkproto")

const (
	CONN_ID    = myprotocol.GroupConnID_v100
	ROUTING_ID = myprotocol.GroupRoutingID_100

	ServiceName = "group.connect"

	HeartbeatInterval = 5 * time.Second
	HeartbeatTimeout  = 11 * time.Second
)

type NetworkProto struct {
	host    myhost.Host
	data    ds.NetworkIface
	logdata logds.AdminIface

	discv *drouting.RoutingDiscovery

	routingTable      RoutingTable
	routingTableMutex sync.RWMutex

	network      Network
	networkMutex sync.RWMutex

	bootTs    uint64 // boot timestamp
	connTimes uint64 // connect change(connect & disconnect) count

	emitters struct {
		evtGroupConnectChange       event.Emitter
		evtGroupNetworkSuccess      event.Emitter
		evtGroupRoutingTableChanged event.Emitter
		evtPullGroupLog             event.Emitter
	}
}

func NewNetworkProto(lhost myhost.Host, rdiscvry *drouting.RoutingDiscovery, ids ipfsds.Batching, ebus event.Bus) (*NetworkProto, error) {

	var err error

	networksvc := &NetworkProto{
		host:         lhost,
		data:         ds.NetworkWrap(ids),
		logdata:      logds.AdminWrap(ids),
		discv:        rdiscvry,
		routingTable: make(RoutingTable),
		network:      make(Network),
		bootTs:       uint64(time.Now().Unix()),
		connTimes:    0,
	}

	lhost.SetStreamHandler(CONN_ID, networksvc.connectHandler)    // 建立路由连接
	lhost.SetStreamHandler(ROUTING_ID, networksvc.routingHandler) // 同步路由表信息

	networksvc.emitters.evtGroupConnectChange, err = ebus.Emitter(&myevent.EvtGroupConnectChange{})
	if err != nil {
		return nil, fmt.Errorf("ebus.Emitter: %s", err.Error())
	}

	networksvc.emitters.evtGroupNetworkSuccess, err = ebus.Emitter(&myevent.EvtGroupNetworkSuccess{})
	if err != nil {
		return nil, fmt.Errorf("ebus.Emitter: %s", err.Error())
	}

	networksvc.emitters.evtGroupRoutingTableChanged, err = ebus.Emitter(&myevent.EvtGroupRoutingTableChanged{})
	if err != nil {
		return nil, fmt.Errorf("ebus.Emitter: %s", err.Error())
	}

	networksvc.emitters.evtPullGroupLog, err = ebus.Emitter(&myevent.EvtPullGroupLog{})
	if err != nil {
		return nil, fmt.Errorf("ebus.Emitter: %s", err.Error())
	}

	// EvtGroupsInit 第一步获取组信息
	sub, err := ebus.Subscribe([]any{new(myevent.EvtGroupInitNetwork), new(myevent.EvtGroupsChange), new(myevent.EvtGroupMemberChange)})
	if err != nil {
		return nil, fmt.Errorf("subscribe event error: %w", err)
	}

	go networksvc.subscribeHandler(context.Background(), sub)

	return networksvc, nil
}

func (n *NetworkProto) GetGroupOnlinePeers(groupID string) ([]peer.ID, error) {
	rTable := n.getRoutingTable(groupID)

	peerIDsMap := make(map[peer.ID]bool)
	for _, connPair := range rTable {
		if connPair.State == StateConnected {
			peerIDsMap[connPair.PeerID0] = true
			peerIDsMap[connPair.PeerID1] = true
		}
	}

	var peerIDs []peer.ID
	for peerID := range peerIDsMap {
		peerIDs = append(peerIDs, peerID)
	}

	return peerIDs, nil
}

func (n *NetworkProto) GetNormalGroupIDs(ctx context.Context) ([]string, error) {

	groupIDs, err := n.logdata.GetGroupIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("data.GetGroupIDs error: %w", err)
	}

	var normalGroupIDs []string
	for _, groupID := range groupIDs {
		state, err := n.logdata.GetState(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("data.GetState error: %w", err)
		}

		switch state {
		case mytype.GroupStateDisband:
			creator, err := n.logdata.GetCreator(ctx, groupID)
			if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
				log.Errorf("data.GetCreate error: %w", err)
			}
			if n.host.ID() == creator {
				normalGroupIDs = append(normalGroupIDs, groupID)
			}
		case mytype.GroupStateNormal:
			normalGroupIDs = append(normalGroupIDs, groupID)
		}
	}

	return normalGroupIDs, nil
}

// checkGroupNetworkAlive 保持群组网络
func (n *NetworkProto) listenGroupNetworkAlive() {
	ctx := context.Background()

	for {
		time.Sleep(10 * time.Second)

		groupIDs, err := n.GetNormalGroupIDs(ctx)
		if err != nil {
			log.Errorf("data.GetGroupIDs error: %v", err)
		}

		log.Debugln("alive groupIDs ", groupIDs)

		for _, groupID := range groupIDs {
			isNeedStart := false
			n.networkMutex.Lock()
			if _, exists := n.network[groupID]; !exists {
				isNeedStart = true
			} else if len(n.network[groupID].Conns) == 0 {
				if n.network[groupID].State != DoingNetworkState {
					isNeedStart = true
				}
			}
			n.networkMutex.Unlock()

			if isNeedStart {
				log.Debugln("listen group network alive, is neet start network")
				n.startNetworking(ctx, groupID)
			}
		}
	}
}

// connectHandler 接收组网请求
func (n *NetworkProto) connectHandler(stream network.Stream) {
	log.Debugln("network.connectHandler: ")

	var connInit pb.GroupConnectInit
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	if err := rd.ReadMsg(&connInit); err != nil {
		log.Errorf("read msg error: %v", err)
		stream.Reset()
		return
	}

	groupID := connInit.GroupId
	remotePeerID := stream.Conn().RemotePeer()

	// 判断对方是否在群里面
	acptPeerIDs, err := n.logdata.GetAgreePeerIDs(context.Background(), groupID)
	if err != nil {
		log.Errorf("data.GetAgreePeerIDs error: %w", err)
		stream.Reset()
		return
	}

	isFoundSelf := false
	isFoundPeer := false
	for _, peerID := range acptPeerIDs {
		if n.host.ID() == peerID {
			isFoundSelf = true
			if isFoundPeer {
				break
			}
		}

		if remotePeerID == peerID {
			isFoundPeer = true
			if isFoundSelf {
				break
			}
		}
	}

	if !isFoundPeer || !isFoundSelf {
		log.Errorf("not agree peer or self exited group")
		stream.Reset()
		return
	}

	hostBootTs := n.bootTs
	hostConnTimes := n.incrConnTimes()

	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&pb.GroupConnectInit{GroupId: groupID, BootTs: hostBootTs, ConnTimes: hostConnTimes}); err != nil {
		log.Errorf("write msg error: %v", err)
		stream.Reset()
		return
	}

	peerConn := Connect{
		PeerID: remotePeerID,
		sendCh: make(chan ConnectPair),
		doneCh: make(chan struct{}),
		reader: rd,
		writer: wt,
	}

	// 更新网络状态
	n.networkMutex.Lock()
	if _, exists := n.network[groupID]; !exists {
		n.network[groupID] = &GroupNetwork{
			Conns: make(map[peer.ID]*Connect),
			State: DoneNetworkState, // 这里是接收的连接，所以默认就完成了
		}
	}

	if _, exists := n.network[groupID].Conns[remotePeerID]; exists {
		log.Errorf("conn is exists, delete old")
		close(n.network[groupID].Conns[remotePeerID].doneCh)
		delete(n.network[groupID].Conns, remotePeerID)
	}
	n.network[groupID].Conns[remotePeerID] = &peerConn
	n.networkMutex.Unlock()

	// 保持网络连接
	peerBootTs := connInit.BootTs
	peerConnTimes := connInit.ConnTimes

	go func() {
		var wg sync.WaitGroup

		wg.Add(2)
		go n.readStream(&wg, stream, groupID, remotePeerID, peerConn.reader, peerConn.doneCh)
		go n.writeStream(&wg, stream, groupID, remotePeerID, peerConn.writer, peerConn.sendCh, peerConn.doneCh)
		wg.Wait()

		stream.Close()
		n.triggerDisconnected(groupID, remotePeerID, peerBootTs, peerConnTimes)
	}()

	log.Debugln("emit: EvtGroupConnectChange")

	// 触发连接改变事件
	if err := n.emitters.evtGroupConnectChange.Emit(myevent.EvtGroupConnectChange{
		GroupID:     groupID,
		PeerID:      remotePeerID,
		IsConnected: true,
	}); err != nil {
		log.Errorf("emit group connect change error: %w", err)
	}

	// 转发连接改变事件
	n.updateRoutingAndForward(groupID, remotePeerID,
		n.connectPair(groupID, n.host.ID(), n.bootTs, hostConnTimes, remotePeerID, peerBootTs, peerConnTimes, StateConnected))
}

// routingHandler 接收更新路由表请求
func (n *NetworkProto) routingHandler(stream network.Stream) {
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	defer rd.Close()

	// 接收对方路由表
	var msg pb.GroupRoutingTable
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("read msg error: %v", err)
		return
	}
	groupID := msg.GroupId

	var remoteGRT GroupRoutingTable
	if err := json.Unmarshal(msg.Payload, &remoteGRT); err != nil {
		log.Errorf("json unmarshal error: %v", err)
		return
	}

	// 获取本地路由表
	localGRT := n.getRoutingTable(groupID)
	bs, err := json.Marshal(localGRT)
	if err != nil {
		log.Errorf("json marshal error: %v", err)
		return
	}

	// 发送本地路由表
	wt := pbio.NewDelimitedWriter(stream)
	if err = wt.WriteMsg(&pb.GroupRoutingTable{GroupId: groupID, Payload: bs}); err != nil {
		log.Errorf("write msg error: %v", err)
		return
	}

	// 合并更新到本地路由
	if n.mergeRoutingTable(groupID, remoteGRT) {
		if err := n.emitters.evtGroupRoutingTableChanged.Emit(myevent.EvtGroupRoutingTableChanged{GroupID: groupID}); err != nil {
			log.Errorf("emit group routing table changed event error: %w", err)
			return
		}
	}
}

func (n *NetworkProto) subscribeHandler(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}
			switch evt := e.(type) {
			case myevent.EvtGroupInitNetwork:
				n.initNetwork()

			case myevent.EvtGroupsChange:

				if evt.AddGroupID != "" {
					n.addGroupNetwork(evt.AddGroupID)
				}

				if evt.DeleteGroupID != "" {
					n.deleteGroupNetwork(evt.DeleteGroupID)
				}

			case myevent.EvtGroupMemberChange:
				n.updateGroupNetwork(evt.GroupID)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (n *NetworkProto) incrConnTimes() uint64 {
	n.connTimes = n.connTimes + 1
	return n.connTimes
}
