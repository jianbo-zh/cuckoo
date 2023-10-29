package mytype

import "github.com/libp2p/go-libp2p/core/peer"

type GroupState = string

const (
	GroupStateUnknown GroupState = ""
	GroupStateExit    GroupState = "exit"    // 0,0
	GroupStateApply   GroupState = "apply"   // 0,1
	GroupStateAgree   GroupState = "agree"   // 1,0
	GroupStateNormal  GroupState = "normal"  // 1,1
	GroupStateDisband GroupState = "disband" // 0,0
)

type Group struct {
	ID             string
	CreatorID      peer.ID
	Name           string
	Avatar         string
	DepositAddress peer.ID
	State          GroupState
}

type GroupDetail struct {
	ID             string
	CreatorID      peer.ID
	Name           string
	Avatar         string
	Notice         string
	AutoJoinGroup  bool
	DepositAddress peer.ID
	State          GroupState
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
