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
			return
		}
		n.network[groupID].State = DoingNetworkState
	}
	n.networkMutex.Unlock()

	findPeerIDs, err := n.findOnlinePeers(ctx, groupRendezvous(groupID), 6)
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

	var okFindPeerIDs []peer.ID
	pullLogTimes := 0
	for _, peerID := range findPeerIDs {
		// 拉取最新日志（找2个）
		if pullLogTimes < 2 {
			if err := n.pullGroupLogs(groupID, peerID); err != nil {
				log.Errorf("pull group logs error: %v", err)
			} else {
				pullLogTimes++
			}
		}

		// 更新路由表
		if err := n.switchRoutingTable(groupID, peerID); err != nil {
			log.Errorf("switch routing table error: %v", err)
			continue
		}

		okFindPeerIDs = append(okFindPeerIDs, peerID)
	}

	if len(okFindPeerIDs) == 0 {
		log.Errorf("not found valid online peer")
		return
	}

	// 开始组网
	if err := n.doNetworking(groupID, okFindPeerIDs); err != nil {
		log.Errorf("do networking error: %w", err)
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
		return fmt.Errorf("emit sync group log error: %w", err)
	}

	if err := <-resultCh; err != nil {
		return fmt.Errorf("sync group log error: %w", err)
	}

	return nil
}

func (n *NetworkProto) updateGroupNetwork(groupID GroupID) {

	acptPeerIDs, err := n.logdata.GetAgreePeerIDs(context.Background(), groupID)
	if err != nil {
		log.Errorf("data.GetAgreePeerIDs error: %w", err)
		return
	}

	acptPeerIDsMap := make(map[peer.ID]struct{})
	for _, peerID := range acptPeerIDs {
		acptPeerIDsMap[peerID] = struct{}{}
	}

	var disconnectPeerIDs []peer.ID

	n.networkMutex.Lock()
	if _, exists := n.network[groupID]; exists {
		for peerID := range n.network[groupID].Conns {
			if _, exists = acptPeerIDsMap[peerID]; !exists {
				disconnectPeerIDs = append(disconnectPeerIDs, peerID)
			}
		}
	}
	n.networkMutex.Unlock()

	for _, peerID := range disconnectPeerIDs {
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
	var connectPeerIDs []peer.ID
	connAlgo := NewConnAlgo(hostID, onlinePeerIDs)
	for _, peerID := range connAlgo.GetClosestPeers() {
		if err := n.connect(groupID, peerID); err != nil {
			log.Errorln("connect error: ", err.Error())
			continue
		}

		connectPeerIDs = append(connectPeerIDs, peerID)
		log.Infoln("connected peer: ", peerID.String())
	}

	if len(connectPeerIDs) == 0 {
		log.Warnln("connect all closest peer failed, try connect found peer")

		for _, peerID := range findPeerIDs {
			if err := n.connect(groupID, peerID); err != nil {
				log.Errorln("connect error: ", err.Error())
				continue
			}

			connectPeerIDs = append(connectPeerIDs, peerID)
			log.Infoln("connected peer: ", peerID.String())
		}
	}

	if len(connectPeerIDs) == 0 {
		return fmt.Errorf("group network failed")
	}

	log.Infoln("------- group network success")
	if err := n.emitters.evtGroupNetworkSuccess.Emit(myevent.EvtGroupNetworkSuccess{
		GroupID: groupID,
		PeerIDs: connectPeerIDs,
	}); err != nil {
		log.Errorln("emit group network success error: %w", err)
	}

	return nil
}
