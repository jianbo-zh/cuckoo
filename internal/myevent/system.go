package myevent

import "github.com/libp2p/go-libp2p/core/peer"

// EvtInviteJoinGroup 邀请加入群组
type EvtInviteJoinGroup struct {
	PeerIDs       []peer.ID
	GroupID       string
	GroupName     string
	GroupAvatar   string
	GroupLamptime uint64
}

// EvtApplyAddContact 申请添加为联系人
type EvtApplyAddContact struct {
	PeerID  peer.ID
	Content string
}
