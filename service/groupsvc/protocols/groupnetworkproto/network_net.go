package groupnetworkproto

import (
	"context"
	"fmt"
	"time"

	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/peer"

	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

// initNetwork 初始化网络
func (n *NetworkProto) initNetwork() error {

	log.Debugln("initNetwork: ")

	ctx := context.Background()

	groupIDs, err := n.GetNormalGroupIDs(ctx)
	if err != nil {
		return fmt.Errorf("data.GetNormalGroupIDs error: %w", err)
	}

	for _, groupID := range groupIDs {
		log.Debugln("init group: ", groupID)

		dutil.Advertise(ctx, n.discv, groupRendezvous(groupID))

		go n.startNetworking(ctx, groupID)
	}

	// 30s后启动监听网络，保证群组网络一直正常
	time.AfterFunc(30*time.Second, n.listenGroupNetworkAlive)

	return nil
}

func (n *NetworkProto) addGroupNetwork(groupID string) error {

	log.Debugln("addGroupNetwork: ", groupID)

	if _, exists := n.network[groupID]; exists {
		log.Warnln("network exists")
		return nil
	}

	ctx := context.Background()
	dutil.Advertise(ctx, n.discv, groupRendezvous(groupID))

	go n.startNetworking(ctx, groupID)

	return nil
}

func (n *NetworkProto) startNetworking(ctx context.Context, groupID string) {

	log.Debugln("startNetworking: ")

	n.networkMutex.Lock()
	if _, exists := n.network[groupID]; !exists {
		n.network[groupID] = &GroupNetwork{
			Conns: make(map[peer.ID]*Connect),
			State: DoingNetworkState,
		}
	} else {
		if n.network[groupID].State == DoingNetworkState {
			// 如果正在启动，则不需要再启动组网了
			n.networkMutex.Unlock()
			log.Debugln("group network is starting, is not need redo")
			return
		}
		n.network[groupID].State = DoingNetworkState
	}
	n.networkMutex.Unlock()

	findPeerIDs, err := n.findOnlinePeers(ctx, groupRendezvous(groupID), 10)
	if err != nil {
		log.Errorf("findOnlinePeers error: %w", err)
		n.networkMutex.Lock()
		n.network[groupID].State = DoneNetworkState
		n.networkMutex.Unlock()
		return

	} else if len(findPeerIDs) == 0 {
		log.Infof("not found peer")
		n.networkMutex.Lock()
		n.network[groupID].State = DoneNetworkState
		n.networkMutex.Unlock()
		return
	}

	// 更新群组日志
	pullLogTimes := 0
	for _, peerID := range findPeerIDs {
		if pullLogTimes < 2 {
			if err := n.pullGroupLogs(groupID, peerID); err != nil {
				log.Errorf("pull group logs error: %v", err)
			} else {
				pullLogTimes++
			}
		}
	}

	// 获取最新的成员IDs
	memberIDs, err := n.logdata.GetMemberIDs(ctx, groupID)
	if err != nil {
		log.Errorln("data.GetMemberIDs error: %w", err)
		return
	}

	if len(memberIDs) <= 1 {
		log.Debugln("no other member, group no need network")
		return
	}

	// 找出发现的会员
	var findMemberIDs []peer.ID
	for _, findPeerID := range findPeerIDs {
		for _, memberID := range memberIDs {
			if findPeerID == memberID {
				findMemberIDs = append(findMemberIDs, findPeerID)
			}
		}
	}

	// 交换路由信息
	var onlineMemberIDs []peer.ID
	for _, peerID := range findMemberIDs {
		if err := n.switchRoutingTable(groupID, peerID); err != nil {
			log.Errorf("switch routing table error: %v", err)
			continue
		}

		onlineMemberIDs = append(onlineMemberIDs, peerID)
	}

	if len(onlineMemberIDs) == 0 {
		log.Errorf("not found valid online member")
		return
	}

	// 开始组网
	if err := n.doNetworking(groupID, onlineMemberIDs); err != nil {
		log.Errorf("do networking error: %v", err)
		return
	}
}

// syncGroupLogs 同步群组日志
func (n *NetworkProto) pullGroupLogs(groupID string, peerID peer.ID) error {
	var resultCh = make(chan error, 1)
	if err := n.emitters.evtPullGroupLog.Emit(myevent.EvtPullGroupLog{
		GroupID: groupID,
		PeerID:  peerID,
		Result:  resultCh,
	}); err != nil {
		return fmt.Errorf("emit pull group log error: %w", err)
	}

	if err := <-resultCh; err != nil {
		return fmt.Errorf("pull group log error: %w", err)
	}

	return nil
}

func (n *NetworkProto) updateGroupNetwork(groupID GroupID) {

	log.Debugln("updateGroupNetwork: ")

	acptPeerIDs, err := n.logdata.GetAgreePeerIDs(context.Background(), groupID)
	if err != nil {
		log.Errorf("data.GetAgreePeerIDs error: %w", err)
		return
	}

	log.Debugln("acptPeerIDs: ")

	acptPeerIDsMap := make(map[peer.ID]struct{})
	for _, peerID := range acptPeerIDs {
		log.Debugln("acptPeerID: ", peerID.String())
		acptPeerIDsMap[peerID] = struct{}{}
	}

	var disconnectPeerIDs []peer.ID

	n.networkMutex.Lock()
	if _, exists := n.network[groupID]; exists {
		for peerID := range n.network[groupID].Conns {
			log.Debugln("exists connected peerID: ", peerID.String())

			if _, exists = acptPeerIDsMap[peerID]; !exists {
				disconnectPeerIDs = append(disconnectPeerIDs, peerID)
			}
		}
	}
	n.networkMutex.Unlock()

	for _, peerID := range disconnectPeerIDs {
		log.Debugln("disconnectPeerID: ", peerID.String())
		n.disconnect(groupID, peerID)
	}
}

func (n *NetworkProto) deleteGroupNetwork(groupID string) error {

	if _, exists := n.network[groupID]; exists {
		for peerID := range n.network[groupID].Conns {
			n.disconnect(groupID, peerID)
		}
	}

	return nil
}

func (n *NetworkProto) findOnlinePeers(ctx context.Context, rendezvous string, limit int) ([]peer.ID, error) {

	var peerIDs []peer.ID
	for i := 0; i < 3; i++ {
		// 尝试3次
		peerChan, err := n.discv.FindPeers(ctx, rendezvous, discovery.Limit(limit), discovery.TTL(time.Minute))
		if err != nil {
			return nil, fmt.Errorf("discv find peers error: %w", err)
		}

		peerIDs = n.handleFindPeers(ctx, peerChan)
		if len(peerIDs) > 0 {
			break
		}
	}

	return peerIDs, nil
}

func (n *NetworkProto) handleFindPeers(ctx context.Context, peerCh <-chan peer.AddrInfo) []peer.ID {
	var peerIDs []peer.ID

	hostID := n.host.ID()
	timeout := time.After(5 * time.Second)
	timeoutTimes := 0
Loop:
	for len(peerIDs) < 5 {
		select {
		case peer := <-peerCh:
			if peer.ID == hostID {
				log.Debugln("found self host")
				continue
			}

			if peer.ID.String() == "" {
				log.Debugln("found empty peer!!!")
				return peerIDs
			}

			peerIDs = append(peerIDs, peer.ID)

		case <-timeout:
			if len(peerIDs) > 0 || timeoutTimes >= 12 {
				log.Debugln("foundPeerIDs or timeout 60s, break loop")
				// 找到或超时60s
				break Loop
			}

			timeout = time.After(5 * time.Second)
			timeoutTimes += 1
			log.Debugln("timeout +1")
		}
	}

	return peerIDs
}

func (n *NetworkProto) doNetworking(groupID string, findPeerIDs []peer.ID) error {

	// 结束后更新组网状态
	defer func() {
		n.networkMutex.Lock()
		n.network[groupID].State = DoneNetworkState
		n.networkMutex.Unlock()
	}()

	var onlinePeerIDs []peer.ID
	hostID := n.host.ID()

	// 合并找到的
	findPeerIDsMap := make(map[peer.ID]struct{})
	for _, peerID := range findPeerIDs {
		onlinePeerIDs = append(onlinePeerIDs, peerID)
		findPeerIDsMap[peerID] = struct{}{}
	}

	// 查找路由在线的
	routingPeerIDs, err := n.GetGroupOnlinePeers(groupID)
	if err != nil {
		return fmt.Errorf("GetGroupOnlinePeers error: %w", err)
	}

	for _, peerID := range routingPeerIDs {
		if peerID != hostID {
			if _, exists := findPeerIDsMap[peerID]; !exists {
				onlinePeerIDs = append(onlinePeerIDs, peerID)
			}
		}
	}

	if len(onlinePeerIDs) == 0 {
		return fmt.Errorf("not found online peer")
	}

	// 获取组网节点
	var connectedPeerIDs []peer.ID
	connAlgo := NewConnAlgo(hostID, onlinePeerIDs)
	for _, peerID := range connAlgo.GetClosestPeers() {
		if err := n.connect(groupID, peerID); err != nil {
			log.Errorln("connect error: ", err.Error())
			continue
		}

		connectedPeerIDs = append(connectedPeerIDs, peerID)
		log.Infoln("connected peer: ", peerID.String())
	}

	if len(connectedPeerIDs) == 0 {
		log.Warnln("connect all closest peer failed, try connect found peer")

		for _, peerID := range findPeerIDs {
			if err := n.connect(groupID, peerID); err != nil {
				log.Errorln("connect error: ", err.Error())
				continue
			}

			connectedPeerIDs = append(connectedPeerIDs, peerID)
			log.Infoln("connected peer: ", peerID.String())
		}
	}

	if len(connectedPeerIDs) == 0 {
		return fmt.Errorf("connect peer failed")
	}

	log.Infoln("------- group network success")
	if err := n.emitters.evtGroupNetworkSuccess.Emit(myevent.EvtGroupNetworkSuccess{
		GroupID: groupID,
		PeerIDs: connectedPeerIDs,
	}); err != nil {
		log.Errorln("emit group network success error: %w", err)
	}

	return nil
}
