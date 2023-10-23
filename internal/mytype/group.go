package mytype

import "github.com/libp2p/go-libp2p/core/peer"

const (
	GroupStateExit    = "exit"    // 0,0
	GroupStateApply   = "apply"   // 0,1
	GroupStateAgree   = "agree"   // 1,0
	GroupStateNormal  = "normal"  // 1,1
	GroupStateDisband = "disband" // 0,0
)

type Group struct {
	ID             string
	Name           string
	Avatar         string
	DepositAddress peer.ID
}

type GroupDetail struct {
	ID             string
	CreatorID      peer.ID
	Name           string
	Avatar         string
	Notice         string
	AutoJoinGroup  bool
	DepositAddress peer.ID
	CreateTime     int64
	UpdateTime     int64
}

// type GroupSession struct {
// 	ID     string
// 	Name   string
// 	Avatar string
// }

type GroupMember struct {
	ID     peer.ID
	Name   string
	Avatar string
}

type GroupMessage struct {
	ID        string
	GroupID   string
	FromPeer  GroupMember
	MsgType   string
	MimeType  string
	Payload   []byte
	IsDeposit bool
	State     MessageState
	Timestamp int64
}
