package network

import (
	"context"
	"encoding/json"

	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/networkproto/pb"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

// switchRouting 交换路由信息
func (n *NetworkProto) switchRoutingTable(groupID string, peerID peer.ID) error {
	stream, err := n.host.NewStream(network.WithUseTransient(context.Background(), ""), peerID, ROUTING_ID)
	if err != nil {
		return err
	}

	// 获取本地路由表
	localGRT := n.getRoutingTable(groupID)
	bs, err := json.Marshal(localGRT)
	if err != nil {
		return err
	}

	// 发送本地路由表
	wt := pbio.NewDelimitedWriter(stream)
	defer wt.Close()

	if err = wt.WriteMsg(&pb.RoutingTable{GroupId: groupID, Payload: bs}); err != nil {
		return err
	}

	// 接收对方路由表
	var msg pb.RoutingTable
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	if err = rd.ReadMsg(&msg); err != nil {
		return err
	}

	var remoteGRT GroupRoutingTable
	if err = json.Unmarshal(msg.Payload, &remoteGRT); err != nil {
		return err
	}

	// 合并更新到本地路由
	n.mergeRoutingTable(groupID, remoteGRT)

	return nil
}

// updateRoutingTable 更新本地路由表
func (n *NetworkProto) updateRoutingTable(groupID GroupID, conn ConnectPair) bool {
	n.routingTableMutex.Lock()
	defer n.routingTableMutex.Unlock()

	if _, exists := n.routingTable[groupID]; !exists {
		n.routingTable[groupID] = make(map[ConnKey]ConnectPair)
		n.routingTable[groupID][connKey(conn.PeerID0, conn.PeerID1)] = conn
		return true
	}

	if _, exists := n.routingTable[groupID][connKey(conn.PeerID0, conn.PeerID1)]; !exists {
		n.routingTable[groupID][connKey(conn.PeerID0, conn.PeerID1)] = conn
		return true
	}

	for key, value := range n.routingTable[groupID] {
		// 更新路由
		if key == connKey(conn.PeerID0, conn.PeerID1) {
			if value.BootTs0 < conn.BootTs0 ||
				value.BootTs1 < conn.BootTs1 ||
				(value.BootTs0 == conn.BootTs0 && value.ConnTimes0 < conn.ConnTimes0) ||
				(value.BootTs1 == conn.BootTs1 && value.ConnTimes1 < conn.ConnTimes1) {

				n.routingTable[groupID][key] = conn
				return true
			}
		}

		// 剔除过期
		if value.PeerID0 == conn.PeerID0 {
			if value.BootTs0 < conn.BootTs0 || (value.BootTs0 == conn.BootTs0 && value.ConnTimes0 < conn.ConnTimes0) {
				delete(n.routingTable[groupID], key)
				continue
			}
		}

		if value.PeerID1 == conn.PeerID1 {
			if value.BootTs1 < conn.BootTs1 || (value.BootTs1 == conn.BootTs1 && value.ConnTimes1 < conn.ConnTimes1) {
				delete(n.routingTable[groupID], key)
				continue
			}
		}
	}
	return false
}

// getRoutingTable 获取路由表
func (n *NetworkProto) getRoutingTable(groupID string) GroupRoutingTable {
	n.routingTableMutex.Lock()
	defer n.routingTableMutex.Unlock()

	if n.routingTable == nil {
		n.routingTable = make(RoutingTable)
	}

	if _, exists := n.routingTable[groupID]; !exists {
		n.routingTable[groupID] = make(GroupRoutingTable)
	}

	return n.routingTable[groupID]
}

// mergeRoutingTable 合并路由表
func (n *NetworkProto) mergeRoutingTable(groupID string, remoteGRT GroupRoutingTable) {
	n.routingTableMutex.Lock()
	defer n.routingTableMutex.Unlock()

	if n.routingTable == nil {
		n.routingTable = make(RoutingTable)
	}

	if _, exists := n.routingTable[groupID]; !exists {
		n.routingTable[groupID] = remoteGRT

	} else {
		// todo: 复制一份路由表
		localGRT := n.routingTable[groupID]

		// 合并路由
		for key, conn := range remoteGRT {
			connPair, exists := localGRT[key]
			if !exists {
				localGRT[key] = conn
			}

			if conn.BootTs0 > connPair.BootTs0 || conn.BootTs1 > connPair.BootTs1 {
				localGRT[key] = conn

			} else if conn.BootTs0 == connPair.BootTs0 && conn.BootTs1 == connPair.BootTs1 {
				if conn.ConnTimes0 > connPair.ConnTimes0 && conn.ConnTimes1 > connPair.ConnTimes1 {
					localGRT[key] = conn
				}
			}
		}

		// 找到 peer 最大值 rtime, lamportime
		pmax := make(map[peer.ID][2]uint64)
		for _, conn := range localGRT {
			if _, exists := pmax[conn.PeerID0]; !exists {
				pmax[conn.PeerID0] = [2]uint64{conn.BootTs0, conn.ConnTimes0}

			} else {
				if conn.BootTs0 > pmax[conn.PeerID0][0] {
					pmax[conn.PeerID0] = [2]uint64{conn.BootTs0, conn.ConnTimes0}

				} else if conn.BootTs0 == pmax[conn.PeerID0][0] && conn.ConnTimes0 > pmax[conn.PeerID0][1] {
					pmax[conn.PeerID0] = [2]uint64{conn.BootTs0, conn.ConnTimes0}
				}
			}

			if _, exists := pmax[conn.PeerID1]; !exists {
				pmax[conn.PeerID1] = [2]uint64{conn.BootTs0, conn.ConnTimes0}

			} else {
				if conn.BootTs0 > pmax[conn.PeerID1][0] {
					pmax[conn.PeerID1] = [2]uint64{conn.BootTs0, conn.ConnTimes0}

				} else if conn.BootTs0 == pmax[conn.PeerID1][0] && conn.ConnTimes0 > pmax[conn.PeerID1][1] {
					pmax[conn.PeerID1] = [2]uint64{conn.BootTs0, conn.ConnTimes0}
				}
			}
		}

		// 剔除无效的连接
		for key, conn := range localGRT {
			if conn.BootTs0 < pmax[conn.PeerID0][0] ||
				conn.BootTs1 < pmax[conn.PeerID1][0] ||
				(conn.BootTs0 == pmax[conn.PeerID0][0] && conn.ConnTimes0 < pmax[conn.PeerID0][1]) ||
				(conn.BootTs1 == pmax[conn.PeerID1][0] && conn.ConnTimes1 < pmax[conn.PeerID1][1]) {

				delete(localGRT, key)
			}
		}

		n.routingTable[groupID] = localGRT
	}
}
