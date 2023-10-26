package myevent

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

type Groups struct {
	GroupID       string
	PeerIDs       []peer.ID
	AcptPeerIDs   []peer.ID
	RefusePeerIDs map[peer.ID]string
}

type EvtGroupAdded struct {
	ID             string
	Name           string
	Avatar         string
	DepositAddress peer.ID
}

// EvtGroupConnectChange 成员连接状态改变
type EvtGroupConnectChange struct {
	GroupID     string
	PeerID      peer.ID
	IsConnected bool
}

// EvtGroupInitNetwork 启动成功后，初始化群组
type EvtGroupInitNetwork struct{}

// EvtGroupNetworkSuccess 群组组网成功
type EvtGroupNetworkSuccess struct {
	GroupID string
	PeerIDs []peer.ID
}

// EvtSyncGroupMessage 同步群组消息
type EvtSyncGroupMessage struct {
	GroupID        string
	DepositAddress peer.ID
}

// EvtGroupsChange 群组发生改变（删除群、新增群）
type EvtGroupsChange struct {
	DeleteGroupID string
	AddGroupID    string
}

// EvtGroupMemberChange 群成员发生改变
type EvtGroupMemberChange struct {
	GroupID string
}

type EvtPullGroupLog struct {
	GroupID string
	PeerID  peer.ID
	Result  chan<- error
}
