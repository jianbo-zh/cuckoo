package myevent

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

type Groups struct {
	GroupID     string
	PeerIDs     []peer.ID
	AcptPeerIDs []peer.ID
}

// EvtGroupConnectChange 成员连接状态改变
type EvtGroupConnectChange struct {
	GroupID     string
	PeerID      peer.ID
	IsConnected bool
}

// EvtGroupsInit 启动成功后，初始化群组
type EvtGroupsInit struct {
	Groups []Groups
}

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
	DeleteGroups []string
	AddGroups    []Groups
}

// EvtGroupMemberChange 群成员发生改变
type EvtGroupMemberChange struct {
	GroupID     string
	PeerIDs     []peer.ID
	AcptPeerIDs []peer.ID
}
