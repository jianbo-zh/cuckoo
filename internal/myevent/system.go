package myevent

import (
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

// EvtInviteJoinGroup 邀请加入群组
type EvtInviteJoinGroup struct {
	GroupID       string
	GroupName     string
	GroupAvatar   string
	GroupLamptime uint64
	Content       string
	Contacts      []mytype.Contact
}

// EvtApplyAddContact 申请添加为联系人
type EvtApplyAddContact struct {
	PeerID      peer.ID
	DepositAddr peer.ID
	Content     string
	Result      chan<- error
}
