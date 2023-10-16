package myevent

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type EvtPushDepositContactMessage struct {
	DepositAddress peer.ID
	ToPeerID       peer.ID
	MsgID          string
	MsgData        []byte
	Result         chan<- error
}

type EvtPushDepositGroupMessage struct {
	DepositAddress peer.ID
	ToGroupID      string
	MsgID          string
	MsgData        []byte
	Result         chan<- error
}

type EvtPullDepositContactMessage struct {
	DepositAddress peer.ID
	MessageHandler func(ctx context.Context, fromPeerID peer.ID, msgID string, msgData []byte) error
}

type EvtPullDepositGroupMessage struct {
	GroupID        string
	DepositAddress peer.ID
	MessageHandler func(ctx context.Context, fromGroupID string, msgID string, msgData []byte) error
}
