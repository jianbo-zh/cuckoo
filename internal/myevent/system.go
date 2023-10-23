package myevent

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

// EvtInviteJoinGroup 邀请加入群组
type GroupInvite struct {
	PeerID         peer.ID
	DepositAddress peer.ID
	GroupLog       []byte
}
type EvtInviteJoinGroup struct {
	GroupID     string
	GroupName   string
	GroupAvatar string
	Content     string
	Invites     []GroupInvite
}

// EvtApplyAddContact 申请添加为联系人
type EvtApplyAddContact struct {
	PeerID      peer.ID
	DepositAddr peer.ID
	Content     string
	Result      chan<- error
}
