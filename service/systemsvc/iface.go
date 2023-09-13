package systemsvc

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type SystemServiceIface interface {
	GetSystemMessageList(ctx context.Context, offset int, limit int) ([]SystemMessage, error)
	ApplyAddContact(ctx context.Context, peerID peer.ID, name string, avatar string, content string) error
	// ApplyJoinGroup(ctx context.Context, groupID string)
	AgreeAddContact(ctx context.Context, ackMsgID string) error
	RejectAddContact(ctx context.Context, ackMsgID string) error

	Close()
}

type MsgType string
type MsgState string

const (
	StateIsSended MsgState = "is_send"
	StateIsAgree  MsgState = "is_agree"
	StateIsReject MsgState = "is_reject"
)

const (
	TypeContactApply MsgType = "contact_apply"
)

type Peer struct {
	PeerID peer.ID
	Name   string
	Avatar string
}

type SystemMessage struct {
	ID       string
	Type     MsgType
	GroupID  string
	Sender   Peer
	Receiver Peer
	Content  string
	State    MsgState
	Ctime    int64
	Utime    int64
}
