package mytype

import "github.com/libp2p/go-libp2p/core/peer"

const (
	GroupStateApply   = "peer_agree"
	GroupStateAgree   = "admin_agree"
	GroupStateReject  = "reject"
	GroupStateExit    = "exit"
	GroupStateNormal  = "normal"
	GroupStateDisband = "disband"
)

type Group struct {
	ID             string
	Name           string
	Avatar         string
	DepositAddress peer.ID
}

type GroupDetail struct {
	ID             string
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
