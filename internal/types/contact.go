package types

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type Contact struct {
	ID            peer.ID
	Name          string
	Avatar        string
	DepositPeerID peer.ID
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

type ContactGetter interface {
	GetContact(ctx context.Context, peerID peer.ID) (*Contact, error)
}
