package event

import "github.com/libp2p/go-libp2p/core/peer"

const (
	MsgTypeText MsgType = "text"
)

type EvtReceivePeerMessage struct {
	MsgID      string
	FromPeerID peer.ID
	MsgType    MsgType
	MimeType   string
	Payload    []byte
	Timestamp  int64
}

type EvtReceiveGroupMessage struct {
	MsgID      string
	GroupID    string
	FromPeerID peer.ID
	MsgType    MsgType
	MimeType   string
	Payload    []byte
	Timestamp  int64
}

type MsgType string
