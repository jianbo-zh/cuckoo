package groupnetworkproto

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

const (
	StateConnected ConnState = iota
	StateDisconnected
)

type ConnState int

type ConnKey = string

type ConnectPair struct {
	GroupID    GroupID   `json:"gi"`
	PeerID0    peer.ID   `json:"p0"`
	PeerID1    peer.ID   `json:"p1"`
	BootTs0    uint64    `json:"b0"`
	BootTs1    uint64    `json:"b1"`
	ConnTimes0 uint64    `json:"c0"`
	ConnTimes1 uint64    `json:"c1"`
	State      ConnState `json:"st"`
}

type GroupPeer struct {
	PeerIDs       map[peer.ID]struct{} // 正式成员，主动连接
	AcptPeerIDs   map[peer.ID]struct{} // 包含正式成员及准成员，（群主邀请，但可能还未接受）
	RefusePeerIDs map[peer.ID]string   // 包含拒绝的成员（移除，退出，拒绝的成员）
}

type RoutingTable = map[GroupID]GroupRoutingTable

type GroupRoutingTable = map[ConnKey]ConnectPair

type GroupPeers = map[GroupID]GroupPeer

type GroupID = string

// ------------ 组网相关 -------------
type NetworkState string

const (
	DoingNetworkState NetworkState = "doing"
	DoneNetworkState  NetworkState = "done"
)

type Connect struct {
	PeerID peer.ID
	reader pbio.ReadCloser
	writer pbio.WriteCloser
	sendCh chan ConnectPair
	doneCh chan struct{}
}

type GroupNetwork struct {
	Conns map[peer.ID]*Connect
	State NetworkState // 组网状态（是否正在启动组网）
}

type Network map[GroupID]*GroupNetwork
