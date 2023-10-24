package mytype

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type ContactState = string

const (
	ContactStateUnknown ContactState = ""
	ContactStateDelete  ContactState = "delete"
	ContactStateNormal  ContactState = "normal"
)

type Contact struct {
	ID             peer.ID
	Name           string
	Avatar         string
	DepositAddress peer.ID
	State          ContactState
}

// type ContactSession struct {
// 	ID     peer.ID
// 	Name   string
// 	Avatar string
// }

type ContactMessage struct {
	ID         string
	MsgType    string
	MimeType   string
	FromPeerID peer.ID
	ToPeerID   peer.ID
	Payload    []byte
	IsDeposit  bool
	State      MessageState
	Timestamp  int64
}

type ContactGetter interface {
	GetContact(ctx context.Context, peerID peer.ID) (*Contact, error)
}
