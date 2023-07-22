package network

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/peer"

	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

// initNetwork 初始化网络
func (n *NetworkService) initNetwork(groups []struct {
	GroupID string
	PeerIDs []peer.ID
}) error {

	n.groupPeersMutex.Lock()
	defer n.groupPeersMutex.Unlock()

	n.groupPeers = make(GroupPeers)
	for _, group := range groups {
		for _, peerID := range group.PeerIDs {
			if _, exists := n.groupPeers[group.GroupID]; !exists {
				n.groupPeers[group.GroupID] = make(map[peer.ID]struct{})
			}
			n.groupPeers[group.GroupID][peerID] = struct{}{}
		}
	}

	ctx := context.Background()
	for _, group := range groups {
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

func (n *NetworkService) addNetwork(groups []struct {
	GroupID string
	PeerIDs []peer.ID
}) error {

	n.groupPeersMutex.Lock()
	defer n.groupPeersMutex.Unlock()

	for _, group := range groups {
		n.groupPeers[group.GroupID] = make(map[peer.ID]struct{})
		for _, peerID := range group.PeerIDs {
			n.groupPeers[group.GroupID][peerID] = struct{}{}
		}
	}

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

func (n *NetworkService) deleteNetwork(groupIDs []GroupID) error {
	n.groupPeersMutex.Lock()
	defer n.groupPeersMutex.Unlock()

	for _, groupID := range groupIDs {
		delete(n.groupPeers, groupID)
	}

	for _, groupID := range groupIDs {
		if _, exists := n.network[groupID]; exists {
			for peerID := range n.network[groupID] {
				n.disconnect(groupID, peerID)
			}

			delete(n.network, groupID)
		}
	}

	return nil
}

func (n *NetworkService) updateNetwork(groupID GroupID, peerIDs []peer.ID) {
	n.groupPeersMutex.Lock()
	defer n.groupPeersMutex.Unlock()

	peerIDsMap := make(map[peer.ID]struct{}, len(peerIDs))
	for _, peerID := range peerIDs {
		peerIDsMap[peerID] = struct{}{}
	}

	var removePeerIDs []peer.ID
	for peerID := range n.groupPeers[groupID] {
		if _, exists := peerIDsMap[peerID]; !exists {
			removePeerIDs = append(removePeerIDs, peerID)
		}
	}

	n.groupPeers[groupID] = peerIDsMap

	if len(removePeerIDs) > 0 {
		for _, peerID := range removePeerIDs {
			if _, exists := n.network[groupID][peerID]; !exists {
				continue
			}

			n.disconnect(groupID, peerID)
		}
	}
}

func (n *NetworkService) handleFindPeers(ctx context.Context, groupID GroupID, peerCh <-chan peer.AddrInfo) error {
	var foundPeerIDs []peer.ID

	hostID := n.host.ID()
	timeout := time.After(5 * time.Second)
Loop:
	for i := 0; i < 5; i++ {
		select {
		case peer := <-peerCh:
			if peer.ID != hostID {
				continue
			}

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
		return fmt.Errorf("not found peer")
	}

	// 更新完路由后，需要组网了
	onlinesMap := make(map[peer.ID]struct{})
	for _, peerID := range foundPeerIDs {
		onlinesMap[peerID] = struct{}{}
	}

	rtable := n.getRoutingTable()
	for _, conn := range rtable[groupID] {
		if conn.State == StateConnected {
			onlinesMap[conn.PeerID0] = struct{}{}
			onlinesMap[conn.PeerID1] = struct{}{}
		}
	}

	var onlinePeerIDs []peer.ID
	for peerID := range onlinesMap {
		if _, exists := n.groupPeers[groupID][peerID]; exists {
			onlinePeerIDs = append(onlinePeerIDs, peerID)
		}
	}

	if len(onlinePeerIDs) == 0 {
		return fmt.Errorf("not found valid peer")
	}

	// 获取组网节点
	connAlgo := NewConnAlgo(hostID, onlinePeerIDs)
	for _, peerID := range connAlgo.GetClosestPeers() {
		n.connect(groupID, peerID)
	}

	return nil
}
