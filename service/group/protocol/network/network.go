package network

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/service/group/datastore"
	networkpb "github.com/jianbo-zh/dchat/service/group/protocol/network/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-msgio/pbio"
)

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/connect.proto=./pb pb/connect.proto
//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/alive.proto=./pb pb/alive.proto

var log = logging.Logger("network")

var StreamTimeout = 1 * time.Minute
var KeepAliveTimeout = 5 * time.Second

const (
	ID         = "/dchat/group/connect/1.0.0"
	ROUTING_ID = "/dchat/group/routing/1.0.0"

	ServiceName    = "group.connect"
	maxMsgSize     = 4 * 1024 // 4K
	KeepAliveWords = "alive"
)

type GroupID = string

type ConnState struct {
	State      networkpb.GroupConnect_State
	Lamportime uint64
}

type GroupNetworkService struct {
	host host.Host

	datastore datastore.GroupIface

	discv *drouting.RoutingDiscovery

	emitters struct {
		evtPeerConnectChange event.Emitter
	}

	networkMutex sync.Mutex
	network      map[GroupID]map[peer.ID]context.CancelFunc
}

func NewGroupNetworkService(h host.Host, rdiscvry *drouting.RoutingDiscovery, ds datastore.GroupIface, eventBus event.Bus) *GroupNetworkService {
	networksvc := &GroupNetworkService{
		host:      h,
		datastore: ds,
		discv:     rdiscvry,
	}

	h.SetStreamHandler(ID, networksvc.Handler)              // 建立路由连接
	h.SetStreamHandler(ROUTING_ID, networksvc.RouteHandler) // 同步路由表信息

	var err error
	if networksvc.emitters.evtPeerConnectChange, err = eventBus.Emitter(&gevent.EvtGroupPeerConnectChange{}); err != nil {
		log.Errorf("set group network emitter error: %v", err)
	}

	return networksvc
}

// Start 启动后台服务，维护网络和连接
func (networksvc *GroupNetworkService) Start() error {
	ctx := context.Background()
	// 获取组列表
	groups, err := networksvc.datastore.ListGroups(ctx)
	if err != nil {
		return err
	}

	for _, group := range groups {
		rendezvous := "/dchat/group/" + group.ID
		dutil.Advertise(ctx, networksvc.discv, rendezvous)

		peerChan, err := networksvc.discv.FindPeers(ctx, rendezvous, discovery.Limit(10))
		if err != nil {
			return err
		}

		go networksvc.findPeers(ctx, group.ID, peerChan)
	}

	return nil
}

func (networksvc *GroupNetworkService) Handler(s network.Stream) {
	go networksvc.acceptConnect(s)
}

func (networksvc *GroupNetworkService) RouteHandler(s network.Stream) {

	fmt.Println("handler....")
	if err := s.Scope().SetService(ServiceName); err != nil {
		log.Errorf("failed to attaching stream to identify service: %v", err)
		s.Reset()
		return
	}
	defer s.Close()

	rd := pbio.NewDelimitedReader(s, maxMsgSize)
	defer rd.Close()

	s.SetReadDeadline(time.Now().Add(StreamTimeout))

	var routing networkpb.GroupRouting
	if err := rd.ReadMsg(&routing); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		s.Reset()
		return
	}
	s.SetReadDeadline(time.Time{})

	otherConnMaps := make(map[string]ConnState, len(routing.Connects))
	for _, oConn := range routing.Connects {
		key := oConn.PeerIdA + "_" + oConn.PeerIdB
		otherConnMaps[key] = ConnState{
			State:      oConn.State,
			Lamportime: oConn.Lamportime,
		}
	}

	connects, err := networksvc.datastore.GetGroupConnects(context.Background(), datastore.GroupID(routing.GroupId))
	if err != nil {
		log.Errorf("get group routing error: %v", err)
		s.Reset()
		return
	}

	for _, conn := range connects {
		key := conn.PeerIdA + "_" + conn.PeerIdB
		state, exists := otherConnMaps[key]
		if !exists || (conn.Lamportime < state.Lamportime) {
			err = networksvc.datastore.CachePeerConnect(context.Background(), datastore.GroupID(conn.GroupId), &networkpb.GroupConnect{
				GroupId:    conn.GroupId,
				PeerIdA:    conn.PeerIdA,
				PeerIdB:    conn.PeerIdB,
				State:      state.State,
				Lamportime: state.Lamportime,
			})
			if err != nil {
				log.Errorf("cache group routing error: %v", err)
				return
			}
		}
	}

	wd := pbio.NewDelimitedWriter(s)
	defer wd.Close()

	s.SetWriteDeadline(time.Now().Add(StreamTimeout))

	err = wd.WriteMsg(&networkpb.GroupRouting{
		GroupId:  routing.GroupId,
		Connects: connects,
	})
	if err != nil {
		log.Errorf("response group routing error: %v", err)
		return
	}
}

func (networksvc *GroupNetworkService) SwitchRouting(groupID string, peerID peer.ID) error {
	stream, err := networksvc.host.NewStream(network.WithUseTransient(context.Background(), ""), peerID, ROUTING_ID)
	if err != nil {
		return err
	}

	connects, err := networksvc.datastore.GetGroupConnects(context.Background(), datastore.GroupID(groupID))
	if err != nil {
		return err
	}

	routing := networkpb.GroupRouting{
		GroupId:  groupID,
		Connects: connects,
	}

	wd := pbio.NewDelimitedWriter(stream)
	defer wd.Close()

	stream.SetWriteDeadline(time.Now().Add(StreamTimeout))
	if err := wd.WriteMsg(&routing); err != nil {
		return err
	}
	stream.SetWriteDeadline(time.Time{})

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rd.Close()

	stream.SetReadDeadline(time.Now().Add(StreamTimeout))

	var otherRouting networkpb.GroupRouting
	if err := rd.ReadMsg(&otherRouting); err != nil {
		return err
	}
	stream.SetReadDeadline(time.Time{})

	// 合并路由
	connMaps := make(map[string]ConnState, len(routing.Connects))
	for _, conn := range routing.Connects {
		key := conn.PeerIdA + "_" + conn.PeerIdB
		connMaps[key] = ConnState{
			State:      conn.State,
			Lamportime: conn.Lamportime,
		}
	}

	if len(otherRouting.Connects) > 0 {
		for _, oConn := range otherRouting.Connects {
			key := oConn.PeerIdA + "_" + oConn.PeerIdB
			state, exists := connMaps[key]
			if !exists || (oConn.Lamportime > state.Lamportime) {
				err = networksvc.datastore.CachePeerConnect(context.Background(), datastore.GroupID(oConn.GroupId), &networkpb.GroupConnect{
					GroupId:    oConn.GroupId,
					PeerIdA:    oConn.PeerIdA,
					PeerIdB:    oConn.PeerIdB,
					State:      oConn.State,
					Lamportime: oConn.Lamportime,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (networksvc *GroupNetworkService) Disconnect(groupID string, peerID peer.ID) error {
	return networksvc.trigerDisconnect(groupID, peerID)
}

// 主动同群节点建立连接
func (networksvc *GroupNetworkService) Connect(groupID string, peerID peer.ID) error {

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	localPeerID := networksvc.host.ID().String()
	remotePeerID := peerID.String()

	stream, err := networksvc.host.NewStream(ctx, peerID, ID)
	if err != nil {
		return err
	}

	lamportime, err := networksvc.datastore.GetNetworkLamportTime(ctx, datastore.GroupID(groupID))
	if err != nil {
		return err
	}

	// 发送组网请求，同步Lamport时钟
	pbmsg := networkpb.GroupAlive{
		GroupId:    groupID,
		Lamportime: lamportime,
	}

	wb := pbio.NewDelimitedWriter(stream)
	stream.SetWriteDeadline(time.Now().Add(StreamTimeout))
	if err := wb.WriteMsg(&pbmsg); err != nil {
		return err
	}
	stream.SetWriteDeadline(time.Time{})

	var pbmsg2 networkpb.GroupAlive
	stream.SetReadDeadline(time.Now().Add(StreamTimeout))

	rb := pbio.NewDelimitedReader(stream, maxMsgSize)
	if err := rb.ReadMsg(&pbmsg2); err != nil {
		return err
	}

	if pbmsg2.Lamportime > lamportime {
		lamportime = pbmsg2.Lamportime
		err = networksvc.datastore.SetNetworkLamportTime(ctx, datastore.GroupID(groupID), lamportime+1)
		if err != nil {
			return err
		}
	}

	// connect
	// todo peerIdA peerIdB 需要排序
	go networksvc.connectChange(&networkpb.GroupConnect{
		GroupId:    groupID,
		PeerIdA:    localPeerID,
		PeerIdB:    remotePeerID,
		State:      networkpb.GroupConnect_CONNECT,
		Lamportime: lamportime + 1,
	})

	// 保存连接取消的句柄，后面需要主动断开连接提供一个抓手
	if err = networksvc.trigerConnect(groupID, peerID, cancelFn); err != nil {
		return err
	}

	// defer disconnect
	defer func() {
		// unconnected
		lamportime, err := networksvc.datastore.TickNetworkLamportTime(context.Background(), datastore.GroupID(groupID))
		if err != nil {
			log.Errorf("tick network lamportime error: %v", err)
			return
		}

		go networksvc.connectChange(&networkpb.GroupConnect{
			GroupId:    groupID,
			PeerIdA:    localPeerID,
			PeerIdB:    remotePeerID,
			State:      networkpb.GroupConnect_UNCONNECT,
			Lamportime: lamportime,
		})

		networksvc.trigerDeleteConnect(groupID, peerID)
	}()

	// keep alive
	for {
		// 如果连接关闭会发生什么？会panic？
		if err = wb.WriteMsg(&pbmsg); err != nil {
			return err
		}

		time.Sleep(5 * time.Second)
	}
}

func (networksvc *GroupNetworkService) acceptConnect(s network.Stream) {

	localPeerID := networksvc.host.ID().String()
	remotePeerID := s.Conn().RemotePeer().String()

	rd := pbio.NewDelimitedReader(s, maxMsgSize)
	defer rd.Close()

	s.SetReadDeadline(time.Now().Add(KeepAliveTimeout))

	var msg networkpb.GroupAlive
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		s.Reset()
		return
	}

	s.SetReadDeadline(time.Time{})

	lamportime, err := networksvc.datastore.GetNetworkLamportTime(context.Background(), datastore.GroupID(msg.GroupId))
	if err != nil {
		log.Errorf("get network lamportime error: %v", err)
		return
	}

	if msg.Lamportime > lamportime {
		lamportime = msg.Lamportime
	}

	nextLamportime := lamportime + 1
	err = networksvc.datastore.SetNetworkLamportTime(context.Background(), datastore.GroupID(msg.GroupId), nextLamportime)
	if err != nil {
		log.Errorf("set network lamportime error: %v", err)
		return
	}

	go networksvc.connectChange(&networkpb.GroupConnect{
		GroupId:    msg.GroupId,
		PeerIdA:    localPeerID,
		PeerIdB:    remotePeerID,
		State:      networkpb.GroupConnect_CONNECT,
		Lamportime: nextLamportime,
	})

	defer func() {
		// unconnected
		lamportime, err := networksvc.datastore.TickNetworkLamportTime(context.Background(), datastore.GroupID(msg.GroupId))
		if err != nil {
			log.Errorf("tick network lamportime error: %v", err)
			return
		}

		go networksvc.connectChange(&networkpb.GroupConnect{
			GroupId:    msg.GroupId,
			PeerIdA:    localPeerID,
			PeerIdB:    remotePeerID,
			State:      networkpb.GroupConnect_UNCONNECT,
			Lamportime: lamportime,
		})
	}()

	for {
		s.SetReadDeadline(time.Now().Add(KeepAliveTimeout))

		var msg networkpb.GroupAlive
		if err := rd.ReadMsg(&msg); err != nil {
			log.Errorf("failed to read CONNECT message from remote peer: %w", err)
			s.Reset()
			return
		}
	}
}

// 已建立连接，则更新路由表，同时判断组网平衡（嫁接或裁剪）
func (networksvc *GroupNetworkService) connectChange(msgpb *networkpb.GroupConnect) {

	err := networksvc.datastore.CachePeerConnect(context.Background(), datastore.GroupID(msgpb.GroupId), msgpb)
	if err != nil {
		log.Errorf("cache peer connect error: %v", err)
		return
	}

	if err = networksvc.emitters.evtPeerConnectChange.Emit(msgpb); err != nil {
		log.Errorf("emit peer connect error: %v", err)
		return
	}

	connPeerIDs, err := networksvc.getConnectPeers(msgpb.GroupId)
	if err != nil {
		log.Errorf("get connect peers error: %v", err)
		return
	}

	switch msgpb.State {
	case networkpb.GroupConnect_CONNECT:
		if len(connPeerIDs) > 5 {
			// 需要裁剪了
			closestPeerIDs, err := networksvc.getClosestPeers(msgpb.GroupId)
			if err != nil {
				log.Errorf("get closest peers error: %v", err)
				return
			}
			for _, connPeerID := range connPeerIDs {
				isFound := false
				for _, closestPeerID := range closestPeerIDs {
					if connPeerID == closestPeerID {
						isFound = true
						break
					}
				}
				if !isFound {
					// 断开不需要的连接
					if err := networksvc.Disconnect(msgpb.GroupId, connPeerID); err != nil {
						log.Errorf("disconnect peer error: %v", err)
						return
					}
					break
				}
			}
		}

	case networkpb.GroupConnect_UNCONNECT:
		if len(connPeerIDs) <= 1 {
			// 需要增加连接
			closestPeerIDs, err := networksvc.getClosestPeers(msgpb.GroupId)
			if err != nil {
				log.Errorf("get closest peers error: %v", err)
				return
			}

			for _, closestPeerID := range closestPeerIDs {
				isFound := false
				for _, connPeerID := range connPeerIDs {
					if connPeerID == closestPeerID {
						isFound = true
						break
					}
				}
				if !isFound {
					// 主动建立连接
					if err = networksvc.Connect(msgpb.GroupId, closestPeerID); err != nil {
						log.Errorf("connect network error: %v", err)
						return
					}
				}
			}
		}
	default:
		// 其他，范围内不管
	}
}

func (networksvc *GroupNetworkService) getConnectPeers(groupID string) ([]peer.ID, error) {
	peerID := networksvc.host.ID().String()
	connects, err := networksvc.datastore.GetGroupConnects(context.Background(), datastore.GroupID(groupID))
	if err != nil {
		return nil, err
	}
	onlinesMap := make(map[string]struct{})
	for _, connect := range connects {
		if connect.State == networkpb.GroupConnect_CONNECT {
			if connect.PeerIdA == peerID {
				onlinesMap[connect.PeerIdB] = struct{}{}

			} else if connect.PeerIdB == peerID {
				onlinesMap[connect.PeerIdA] = struct{}{}
			}
		}
	}

	var peerIDs []peer.ID
	for pid := range onlinesMap {
		peerID, _ := peer.Decode(pid)
		peerIDs = append(peerIDs, peerID)
	}

	return peerIDs, nil
}

func (networksvc *GroupNetworkService) getClosestPeers(groupID string) ([]peer.ID, error) {

	peerID := networksvc.host.ID()
	connects, err := networksvc.datastore.GetGroupConnects(context.Background(), datastore.GroupID(groupID))
	if err != nil {
		return nil, err
	}
	onlinesMap := make(map[string]struct{})
	for _, connect := range connects {
		if connect.State == networkpb.GroupConnect_CONNECT {
			onlinesMap[connect.PeerIdA] = struct{}{}
			onlinesMap[connect.PeerIdB] = struct{}{}
		}
	}

	var peerIDs []peer.ID
	for pid := range onlinesMap {
		peerID, _ := peer.Decode(pid)
		peerIDs = append(peerIDs, peerID)
	}

	circular, err := NewCircularQueue(peerID, peerIDs)
	if err != nil {
		return nil, err
	}

	return circular.GetClosestPeers(), nil
}

func (networksvc *GroupNetworkService) findPeers(ctx context.Context, groupID string, peerCh <-chan peer.AddrInfo) {

	peerID := networksvc.host.ID()
	succNum := 0
	for peerA := range peerCh {
		if peerA.ID == peerID {
			continue
		}

		// 交换路由表信息
		err := networksvc.SwitchRouting(groupID, peerA.ID)
		if err != nil {
			continue
		}

		succNum += 1
		if succNum >= 3 {
			break
		}
	}

	// 更新完路由后，需要组网了
	connects, err := networksvc.datastore.GetGroupConnects(ctx, datastore.GroupID(groupID))
	if err != nil {
		log.Errorf("get group connects error: %w", err)
		return
	}
	onlinesMap := make(map[string]struct{})
	for _, connect := range connects {
		if connect.State == networkpb.GroupConnect_CONNECT {
			onlinesMap[connect.PeerIdA] = struct{}{}
			onlinesMap[connect.PeerIdB] = struct{}{}
		}
	}

	var onlinePeerIDs []peer.ID
	for pid := range onlinesMap {
		peerID, _ := peer.Decode(pid)
		onlinePeerIDs = append(onlinePeerIDs, peerID)
	}

	circular, err := NewCircularQueue(peerID, onlinePeerIDs)
	if err != nil {
		log.Errorf("new circularqueue error: %v", err)
		return
	}

	// 获取组网节点
	for _, peerID := range circular.GetClosestPeers() {
		go networksvc.Connect(groupID, peerID)
	}
}

type CircularQueue struct {
	peerIndex  int
	orderQueue []peer.ID
}

func NewCircularQueue(peerID peer.ID, otherPeerIDs []peer.ID) (*CircularQueue, error) {
	unimap := make(map[uint64]peer.ID, len(otherPeerIDs)+1)

	hash := sha256.Sum256([]byte(peerID))
	peerBint := big.NewInt(0).SetBytes(hash[:]).Uint64()
	unimap[peerBint] = peerID

	for _, pid := range otherPeerIDs {
		hash := sha256.Sum256([]byte(pid))
		bint := big.NewInt(0).SetBytes(hash[:]).Uint64()

		if _, exists := unimap[bint]; !exists {
			unimap[bint] = pid
		}
	}

	orderUints := make([]uint64, len(unimap))
	for val := range unimap {
		orderUints = append(orderUints, val)
	}

	sort.Slice(orderUints, func(i, j int) bool {
		return orderUints[i] < orderUints[j]
	})

	var peerIndex int
	orderQueue := make([]peer.ID, len(orderUints))
	for i, bint := range orderUints {
		if bint == peerBint {
			peerIndex = i
		}
		orderQueue = append(orderQueue, unimap[bint])
	}

	return &CircularQueue{
		peerIndex:  peerIndex,
		orderQueue: orderQueue,
	}, nil
}

func (cq *CircularQueue) GetClosestPeers() (peers []peer.ID) {
	size := len(cq.orderQueue)
	if size <= 1 {
		return
	} else if size <= 3 {
		for i, peerID := range cq.orderQueue {
			if cq.peerIndex != i {
				peers = append(peers, peerID)
			}
		}
	} else if size <= 100 {
		// 3条线：前后2条，90度一条
		fi := (cq.peerIndex + 1) % size
		bi := (cq.peerIndex - 1 + size) % size
		i90 := cq.peerIndex + (size / 2)
		peers = []peer.ID{
			cq.orderQueue[fi],
			cq.orderQueue[bi],
			cq.orderQueue[i90],
		}

	} else {
		// 4条线：前后2条，60度2条
		fi := (cq.peerIndex + 1) % size
		bi := (cq.peerIndex - 1 + size) % size
		fi60 := (cq.peerIndex + (size / 3)) % size
		bi60 := (cq.peerIndex - (size / 3) + size) % size
		peers = []peer.ID{
			cq.orderQueue[fi],
			cq.orderQueue[bi],
			cq.orderQueue[fi60],
			cq.orderQueue[bi60],
		}
	}

	return peers
}

func (networksvc *GroupNetworkService) trigerConnect(groupID GroupID, peerID peer.ID, cancelFn context.CancelFunc) error {
	networksvc.networkMutex.Lock()
	defer networksvc.networkMutex.Unlock()

	if _, exists := networksvc.network[groupID]; !exists {
		networksvc.network[groupID] = make(map[peer.ID]context.CancelFunc)
	} else {
		return fmt.Errorf("cancel func is exists")
	}

	networksvc.network[groupID][peerID] = cancelFn

	return nil
}

func (networksvc *GroupNetworkService) trigerDisconnect(groupID GroupID, peerID peer.ID) error {
	networksvc.networkMutex.Lock()
	defer networksvc.networkMutex.Unlock()

	if _, exists := networksvc.network[groupID]; !exists {
		return fmt.Errorf("cancel func is exists")

	} else if cancelFn, exists := networksvc.network[groupID][peerID]; !exists {
		return fmt.Errorf("cancel func is exists")

	} else {
		cancelFn()
	}

	return nil
}

func (networksvc *GroupNetworkService) trigerDeleteConnect(groupID GroupID, peerID peer.ID) error {
	networksvc.networkMutex.Lock()
	defer networksvc.networkMutex.Unlock()

	if _, exists := networksvc.network[groupID]; !exists {
		return fmt.Errorf("cancel func is exists")

	} else if _, exists := networksvc.network[groupID][peerID]; !exists {
		return fmt.Errorf("cancel func is exists")

	} else {
		delete(networksvc.network[groupID], peerID)
	}

	return nil
}
