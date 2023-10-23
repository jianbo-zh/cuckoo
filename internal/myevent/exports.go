package myevent

import (
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

type EvtReceiveContactMessage struct {
	MsgID      string
	FromPeerID peer.ID
	MsgType    string
	MimeType   string
	Payload    []byte
	Timestamp  int64
}

type EvtReceiveGroupMessage struct {
	MsgID      string
	GroupID    string
	FromPeerID peer.ID
	MsgType    string
	MimeType   string
	Payload    []byte
	Timestamp  int64
}

// EvtPeerStateChanged Peer在线状态
type EvtPeerStateChanged struct {
	PeerID peer.ID
	Online bool
}

type EvtSessionAdded struct {
	Type   mytype.SessionType
	ID     string
	Name   string
	Avatar string
	RelID  string
}

type UpdateSession struct {
	ID      string
	LastMsg string
	Unreads int64
}

type EvtSessionUpdated struct {
	Sessions []UpdateSession
}
