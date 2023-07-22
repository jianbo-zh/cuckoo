package network

import (
	"context"
	"encoding/json"

	"github.com/jianbo-zh/dchat/service/group/protocol/network/pb"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

// switchRouting 交换路由信息
func (n *NetworkService) switchRoutingTable(groupID string, peerID peer.ID) error {
	stream, err := n.host.NewStream(network.WithUseTransient(context.Background(), ""), peerID, ROUTING_ID)
	if err != nil {
		return err
	}

	// 获取本地路由表
	localRoutingTable := n.getRoutingTable()
	bs, err := json.Marshal(localRoutingTable)
	if err != nil {
		return err
	}

	// 发送本地路由表
	wt := pbio.NewDelimitedWriter(stream)
	defer wt.Close()

	if err = wt.WriteMsg(&pb.RoutingTable{Payload: bs}); err != nil {
		return err
	}

	// 接收对方路由表
	var prt pb.RoutingTable
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	if err = rd.ReadMsg(&prt); err != nil {
		return err
	}

	var remoteRoutingTable RoutingTable
	if err = json.Unmarshal(prt.Payload, &remoteRoutingTable); err != nil {
		return err
	}

	// 合并更新到本地路由
	n.mergeRoutingTable(remoteRoutingTable)

	return nil
}

// updateRoutingTable 更新本地路由表
func (n *NetworkService) updateRoutingTable(groupID GroupID, conn ConnectPair) bool {

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
func (n *NetworkService) getRoutingTable() RoutingTable {
	n.routingTableMutex.Lock()
	defer n.routingTableMutex.Unlock()

	if n.routingTable == nil {
		n.routingTable = make(RoutingTable)
	}

	return n.routingTable
}

// mergeRoutingTable 合并路由表
func (n *NetworkService) mergeRoutingTable(rt RoutingTable) {
	n.routingTableMutex.Lock()
	defer n.routingTableMutex.Unlock()

	if n.routingTable == nil {
		n.routingTable = make(RoutingTable)
	}

	for groupID, conns := range rt {
		if _, exists := n.routingTable[groupID]; !exists {
			n.routingTable[groupID] = conns

		} else {
			// todo: 复制一份路由表
			grtable := n.routingTable[groupID]

			// 合并路由
			for key, conn := range conns {
				connPair, exists := grtable[key]
				if !exists {
					grtable[key] = conn
				}

				if conn.BootTs0 > connPair.BootTs0 || conn.BootTs1 > connPair.BootTs1 {
					grtable[key] = conn

				} else if conn.BootTs0 == connPair.BootTs0 && conn.BootTs1 == connPair.BootTs1 {
					if conn.ConnTimes0 > connPair.ConnTimes0 && conn.ConnTimes1 > connPair.ConnTimes1 {
						grtable[key] = conn
					}
				}
			}

			// 找到 peer 最大值 rtime, lamportime
			pmax := make(map[peer.ID][2]uint64)
			for _, conn := range grtable {
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
			for key, conn := range grtable {
				if conn.BootTs0 < pmax[conn.PeerID0][0] ||
					conn.BootTs1 < pmax[conn.PeerID1][0] ||
					(conn.BootTs0 == pmax[conn.PeerID0][0] && conn.ConnTimes0 < pmax[conn.PeerID0][1]) ||
					(conn.BootTs1 == pmax[conn.PeerID1][0] && conn.ConnTimes1 < pmax[conn.PeerID1][1]) {

					delete(grtable, key)
				}
			}

			n.routingTable[groupID] = grtable
		}
	}
}
