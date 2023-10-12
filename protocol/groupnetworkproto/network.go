package groupnetworkproto

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	ds "github.com/jianbo-zh/dchat/datastore/ds/groupnetworkds"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/myprotocol"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/grouppb"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"

	ipfsds "github.com/ipfs/go-datastore"
	logging "github.com/jianbo-zh/go-log"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var log = logging.Logger("group-network")

const (
	CONN_ID    = myprotocol.GroupConnID_v100
	ROUTING_ID = myprotocol.GroupRoutingID_100

	ServiceName = "group.connect"
	maxMsgSize  = 4 * 1024 // 4K

	HeartbeatInterval = 5 * time.Second
	HeartbeatTimeout  = 11 * time.Second
)

type NetworkProto struct {
	host myhost.Host
	data ds.NetworkIface

	discv *drouting.RoutingDiscovery

	routingTable      RoutingTable
	routingTableMutex sync.RWMutex

	network      Network
	networkMutex sync.RWMutex

	groupPeers      GroupPeers
	groupPeersMutex sync.RWMutex

	bootTs    uint64 // boot timestamp
	connTimes uint64 // connect change(connect & disconnect) count

	emitters struct {
		evtGroupConnectChange  event.Emitter
		evtGroupNetworkSuccess event.Emitter
	}
}

func NewNetworkProto(lhost myhost.Host, rdiscvry *drouting.RoutingDiscovery, ids ipfsds.Batching, ebus event.Bus) (*NetworkProto, error) {

	var err error

	networksvc := &NetworkProto{
		host:         lhost,
		data:         ds.NetworkWrap(ids),
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

	// EvtGroupsInit 第一步获取组信息
	sub, err := ebus.Subscribe([]any{new(myevent.EvtGroupsInit), new(myevent.EvtGroupsChange), new(myevent.EvtGroupMemberChange)})
	if err != nil {
		return nil, fmt.Errorf("subscribe event error: %w", err)

	} else {
		go networksvc.subscribeHandler(context.Background(), sub)
	}

	return networksvc, nil
}

func (n *NetworkProto) GetGroupOnlinePeers(groupID string) ([]peer.ID, error) {
	rTable := n.getRoutingTable(groupID)

	peerIDsMap := make(map[peer.ID]bool, 0)
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

// connectHandler 接收组网请求
func (n *NetworkProto) connectHandler(stream network.Stream) {

	var connInit pb.GroupConnectInit
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	if err := rd.ReadMsg(&connInit); err != nil {
		log.Errorf("read msg error: %v", err)
		stream.Reset()
		return
	}

	groupID := connInit.GroupId
	peerID := stream.Conn().RemotePeer()

	// 判断对方是否在群里面
	n.groupPeersMutex.RLock()
	if _, exists := n.groupPeers[groupID]; exists {
		if _, exists := n.groupPeers[groupID].AcptPeerIDs[peerID]; !exists {
			log.Errorf("not group peer, refuse conn")
			stream.Reset()
			return
		}
	}
	n.groupPeersMutex.RUnlock()

	hostBootTs := n.bootTs
	hostConnTimes := n.incrConnTimes()

	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&pb.GroupConnectInit{GroupId: groupID, BootTs: hostBootTs, ConnTimes: hostConnTimes}); err != nil {
		log.Errorf("write msg error: %v", err)
		stream.Reset()
		return
	}

	peerConn := Connect{
		PeerID: peerID,
		sendCh: make(chan ConnectPair),
		doneCh: make(chan struct{}),
		reader: rd,
		writer: wt,
	}

	// 更新网络状态
	n.networkMutex.Lock()
	if _, exists := n.network[groupID]; !exists {
		n.network[groupID] = make(map[peer.ID]*Connect)
	}

	if _, exists := n.network[groupID][peerID]; exists {
		log.Errorf("conn is exists, delete old")
		close(n.network[groupID][peerID].doneCh)
		delete(n.network[groupID], peerID)
	}
	n.network[groupID][peerID] = &peerConn
	n.networkMutex.Unlock()

	// 保持网络连接
	peerBootTs := connInit.BootTs
	peerConnTimes := connInit.ConnTimes

	go func() {
		var wg sync.WaitGroup

		wg.Add(2)
		go n.readStream(&wg, stream, groupID, peerID, peerConn.reader, peerConn.doneCh)
		go n.writeStream(&wg, stream, groupID, peerID, peerConn.writer, peerConn.sendCh, peerConn.doneCh)
		wg.Wait()

		stream.Close()
		n.triggerDisconnected(groupID, peerID, peerBootTs, peerConnTimes)
	}()

	// 触发连接改变事件
	if err := n.emitters.evtGroupConnectChange.Emit(myevent.EvtGroupConnectChange{
		GroupID:     groupID,
		PeerID:      peerID,
		IsConnected: true,
	}); err != nil {
		log.Errorf("emit group connect change error: %w", err)
	}

	// 转发连接改变事件
	n.updateRoutingAndForward(groupID, peerID,
		n.connectPair(groupID, n.host.ID(), n.bootTs, hostConnTimes, peerID, peerBootTs, peerConnTimes, StateConnected))
}

// routingHandler 接收更新路由表请求
func (n *NetworkProto) routingHandler(stream network.Stream) {
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
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
	n.mergeRoutingTable(groupID, remoteGRT)
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
			case myevent.EvtGroupsInit:
				n.initNetwork(evt.Groups)

			case myevent.EvtGroupsChange:
				if len(evt.AddGroups) > 0 {
					n.addNetwork(evt.AddGroups)
				}

				if len(evt.DeleteGroups) > 0 {
					n.deleteNetwork(evt.DeleteGroups)
				}

			case myevent.EvtGroupMemberChange:
				n.updateNetwork(evt.GroupID, evt.PeerIDs, evt.AcptPeerIDs)
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
