package myevent

import "github.com/libp2p/go-libp2p/core/peer"

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

type EvtOnlineStateDiscover struct {
	OnlineState map[peer.ID]bool
}
