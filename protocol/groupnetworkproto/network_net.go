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
func (n *NetworkProto) initNetwork(groups []myevent.Groups) error {

	n.groupPeersMutex.Lock()
	n.groupPeers = make(GroupPeers)
	for _, group := range groups {
		n.groupPeers[group.GroupID] = GroupPeer{
			PeerIDs:       make(map[peer.ID]struct{}),
			AcptPeerIDs:   make(map[peer.ID]struct{}),
			RefusePeerIDs: make(map[peer.ID]string),
		}

		for _, peerID := range group.PeerIDs {
			n.groupPeers[group.GroupID].PeerIDs[peerID] = struct{}{}
		}

		for _, peerID := range group.AcptPeerIDs {
			n.groupPeers[group.GroupID].AcptPeerIDs[peerID] = struct{}{}
		}

		for peerID, logID := range group.RefusePeerIDs {
			n.groupPeers[group.GroupID].RefusePeerIDs[peerID] = logID
		}
	}
	n.groupPeersMutex.Unlock()

	ctx := context.Background()
	for _, group := range groups {
		log.Debugln("init group: ", group.GroupID)

		rendezvous := "/dchat/group/" + group.GroupID
		fmt.Println("1, advertise rendezvous: ", rendezvous)
		dutil.Advertise(ctx, n.discv, rendezvous)

		go n.startNetworking(ctx, group.GroupID, rendezvous)
	}

	return nil
}

func (n *NetworkProto) addNetwork(groups []myevent.Groups) error {
	fmt.Println("addNetwork...")

	n.groupPeersMutex.Lock()
	for _, group := range groups {
		n.groupPeers[group.GroupID] = GroupPeer{
			PeerIDs:       make(map[peer.ID]struct{}),
			AcptPeerIDs:   make(map[peer.ID]struct{}),
			RefusePeerIDs: make(map[peer.ID]string),
		}

		for _, peerID := range group.PeerIDs {
			n.groupPeers[group.GroupID].PeerIDs[peerID] = struct{}{}
		}

		for _, peerID := range group.AcptPeerIDs {
			n.groupPeers[group.GroupID].AcptPeerIDs[peerID] = struct{}{}
		}

		for peerID, logID := range group.RefusePeerIDs {
			n.groupPeers[group.GroupID].RefusePeerIDs[peerID] = logID
		}
	}
	n.groupPeersMutex.Unlock()

	ctx := context.Background()
	for _, group := range groups {
		if _, exists := n.network[group.GroupID]; exists {
			continue
		}

		rendezvous := "/dchat/group/" + group.GroupID
		dutil.Advertise(ctx, n.discv, rendezvous)

		go n.startNetworking(ctx, group.GroupID, rendezvous)
	}

	return nil
}

func (n *NetworkProto) startNetworking(ctx context.Context, groupID string, rendezvous string, foundPeerIDs ...peer.ID) {

	foundPeerIDs2, err := n.findOnlinePeers(ctx, rendezvous, 5)
	if err != nil {
		log.Errorf("findOnlinePeers error: %w", err)
		return
	}

	foundPeerIDs = append(foundPeerIDs, foundPeerIDs2...)

	for _, peerID := range foundPeerIDs {
		err := n.switchRoutingTable(groupID, peerID)
		if err != nil {
			log.Errorf("switch routing table error: %w", err)
		}
	}

	if len(foundPeerIDs) > 0 {
		n.doNetworking(groupID, foundPeerIDs)
	}
}

func (n *NetworkProto) updateNetwork(groupID GroupID, peerIDs []peer.ID, acptPeerIDs []peer.ID, refusePeerIDs map[peer.ID]string) {
	fmt.Println("update network...")

	peerIDsMap := make(map[peer.ID]struct{})
	for _, peerID := range peerIDs {
		peerIDsMap[peerID] = struct{}{}
	}

	acptPeerIDsMap := make(map[peer.ID]struct{})
	for _, peerID := range acptPeerIDs {
		acptPeerIDsMap[peerID] = struct{}{}
	}

	n.groupPeersMutex.Lock()

	var removePeerIDs []peer.ID
	if _, exists := n.groupPeers[groupID]; exists {
		for peerID := range n.groupPeers[groupID].AcptPeerIDs {
			if _, exists := acptPeerIDsMap[peerID]; !exists {
				removePeerIDs = append(removePeerIDs, peerID)
			}
		}
	}

	n.groupPeers[groupID] = GroupPeer{
		PeerIDs:       peerIDsMap,
		AcptPeerIDs:   acptPeerIDsMap,
		RefusePeerIDs: refusePeerIDs,
	}
	n.groupPeersMutex.Unlock()

	if len(removePeerIDs) > 0 {
		for _, peerID := range removePeerIDs {
			if _, exists := n.network[groupID][peerID]; !exists {
				continue
			}

			n.disconnect(groupID, peerID)
		}
	}
}

func (n *NetworkProto) deleteNetwork(groupIDs []GroupID) error {
	fmt.Println("delete network...")

	n.groupPeersMutex.Lock()
	for _, groupID := range groupIDs {
		delete(n.groupPeers, groupID)
	}
	n.groupPeersMutex.Unlock()

	for _, groupID := range groupIDs {
		if _, exists := n.network[groupID]; exists {
			for peerID := range n.network[groupID] {
				n.disconnect(groupID, peerID)
			}

			n.networkMutex.Lock()
			delete(n.network, groupID)
			n.networkMutex.Unlock()
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

			fmt.Println("find peer: ", peer.ID.String())
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

func (n *NetworkProto) doNetworking(groupID string, foundPeerIDs []peer.ID) {

	hostID := n.host.ID()

	if len(foundPeerIDs) == 0 {
		// 没有找到则返回
		log.Warnln("not found group peer", groupID)
		return
	}
	log.Debugln("found peers: ", len(foundPeerIDs))

	// 更新完路由后，需要组网了
	onlinesMap := make(map[peer.ID]struct{})
	for _, peerID := range foundPeerIDs {
		onlinesMap[peerID] = struct{}{}
	}

	for _, conn := range n.getRoutingTable(groupID) {
		if conn.State == StateConnected {
			onlinesMap[conn.PeerID0] = struct{}{}
			onlinesMap[conn.PeerID1] = struct{}{}
		}
	}

	fmt.Println("onlinesMap", onlinesMap)

	n.groupPeersMutex.RLock()
	var onlinePeerIDs []peer.ID
	var isOnlySelf bool
	if len(n.groupPeers[groupID].PeerIDs) == 1 {
		if _, exists := n.groupPeers[groupID].PeerIDs[hostID]; exists {
			isOnlySelf = true
		}
	}

	fmt.Println(n.groupPeers[groupID].PeerIDs)
	fmt.Println(isOnlySelf)

	for peerID := range onlinesMap {
		if isOnlySelf { // 如果无主动连接Peer，则所有都可以连
			onlinePeerIDs = append(onlinePeerIDs, peerID)

		} else if _, exists := n.groupPeers[groupID].PeerIDs[peerID]; exists { // 有主动连接Peer，则检查是否在其中
			onlinePeerIDs = append(onlinePeerIDs, peerID)
		}
	}
	n.groupPeersMutex.RUnlock()

	fmt.Println("onlinePeerIDs", onlinePeerIDs)

	if len(onlinePeerIDs) == 0 {
		log.Warnln("not found valid peer")
		return
	}

	// 获取组网节点
	connAlgo := NewConnAlgo(hostID, onlinePeerIDs)
	for _, peerID := range connAlgo.GetClosestPeers() {
		if err := n.connect(groupID, peerID); err != nil {
			log.Errorln("connect error: ", err.Error())
			return
		}

		log.Infoln("connected peer: ", peerID.String())
	}

	log.Infoln("------- group network success")
	if err := n.emitters.evtGroupNetworkSuccess.Emit(myevent.EvtGroupNetworkSuccess{
		GroupID: groupID,
		PeerIDs: onlinePeerIDs,
	}); err != nil {
		log.Errorln("emit group network success error: %w", err)
	}
}
