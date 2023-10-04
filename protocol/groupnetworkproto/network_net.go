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
		for _, peerID := range group.PeerIDs {
			if _, exists := n.groupPeers[group.GroupID]; !exists {
				n.groupPeers[group.GroupID] = GroupPeer{
					PeerIDs:     make(map[peer.ID]struct{}),
					AcptPeerIDs: make(map[peer.ID]struct{}),
				}
			}
			n.groupPeers[group.GroupID].PeerIDs[peerID] = struct{}{}
		}

		for _, peerID := range group.AcptPeerIDs {
			if _, exists := n.groupPeers[group.GroupID]; !exists {
				n.groupPeers[group.GroupID] = GroupPeer{
					PeerIDs:     make(map[peer.ID]struct{}),
					AcptPeerIDs: make(map[peer.ID]struct{}),
				}
			}
			n.groupPeers[group.GroupID].AcptPeerIDs[peerID] = struct{}{}
		}
	}
	n.groupPeersMutex.Unlock()

	ctx := context.Background()
	for _, group := range groups {
		log.Debugln("init group: ", group.GroupID)

		rendezvous := "/dchat/group/" + group.GroupID
		dutil.Advertise(ctx, n.discv, rendezvous)

		peerChan, err := n.discv.FindPeers(ctx, rendezvous, discovery.Limit(10))
		if err != nil {
			log.Errorf("discv find peers error: %w", err)
			continue
		}

		go n.handleFindPeers(ctx, group.GroupID, peerChan)
	}

	return nil
}

func (n *NetworkProto) addNetwork(groups []myevent.Groups) error {

	n.groupPeersMutex.Lock()
	for _, group := range groups {
		n.groupPeers[group.GroupID] = GroupPeer{
			PeerIDs:     make(map[peer.ID]struct{}),
			AcptPeerIDs: make(map[peer.ID]struct{}),
		}

		for _, peerID := range group.PeerIDs {
			n.groupPeers[group.GroupID].PeerIDs[peerID] = struct{}{}
		}

		for _, peerID := range group.AcptPeerIDs {
			n.groupPeers[group.GroupID].AcptPeerIDs[peerID] = struct{}{}
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

		peerChan, err := n.discv.FindPeers(ctx, rendezvous, discovery.Limit(10))
		if err != nil {
			return err
		}

		go n.handleFindPeers(ctx, group.GroupID, peerChan)
	}

	return nil
}

func (n *NetworkProto) deleteNetwork(groupIDs []GroupID) error {

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

func (n *NetworkProto) updateNetwork(groupID GroupID, peerIDs []peer.ID, acptPeerIDs []peer.ID) {

	peerIDsMap := make(map[peer.ID]struct{}, len(peerIDs))
	for _, peerID := range peerIDs {
		peerIDsMap[peerID] = struct{}{}
	}

	acptPeerIDsMap := make(map[peer.ID]struct{}, len(acptPeerIDs))
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
		PeerIDs:     peerIDsMap,
		AcptPeerIDs: acptPeerIDsMap,
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

func (n *NetworkProto) handleFindPeers(ctx context.Context, groupID GroupID, peerCh <-chan peer.AddrInfo) {
	var foundPeerIDs []peer.ID

	log.Infoln("handleFindPeers: ", groupID)

	hostID := n.host.ID()
	timeout := time.After(5 * time.Second)
Loop:
	for i := 0; i < 5; i++ {
		select {
		case peer := <-peerCh:
			if peer.ID == hostID {
				continue
			}

			if peer.ID.String() == "" {
				continue
			}

			fmt.Println("find peer: ", peer.ID.String())
			foundPeerIDs = append(foundPeerIDs, peer.ID)

			// 交换路由表信息
			err := n.switchRoutingTable(groupID, peer.ID)
			if err != nil {
				log.Errorf("switch routing table error: %v", err)
			}

		case <-timeout:
			break Loop
		}
	}

	if len(foundPeerIDs) == 0 {
		// 没有找到则返回
		log.Warnln("not found group peer", groupID)
		return
	}

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
	for peerID := range onlinesMap {
		if len(n.groupPeers[groupID].PeerIDs) == 0 { // 如果无主动连接Peer，则所有都可以连
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
