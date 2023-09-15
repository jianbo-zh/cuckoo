package types

import "github.com/libp2p/go-libp2p/core/peer"

type Contact struct {
	ID     peer.ID
	Name   string
	Avatar string
}

type ContactSession struct {
	ID     peer.ID
	Name   string
	Avatar string
}

type ContactMessage struct {
	ID         string
	MsgType    string
	MimeType   string
	FromPeerID peer.ID
	ToPeerID   peer.ID
	Payload    []byte
	Timestamp  int64
	Lamportime uint64
}
