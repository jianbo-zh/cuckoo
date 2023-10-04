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

	ResultCallback func(toPeerID peer.ID, msgID string, err error)
}

type EvtPushDepositGroupMessage struct {
	DepositAddress peer.ID
	ToGroupID      string
	MsgID          string
	MsgData        []byte

	ResultCallback func(toGroupID string, msgID string, err error)
}

type EvtPullDepositContactMessage struct {
	DepositAddress peer.ID
	MessageHandler func(ctx context.Context, fromPeerID peer.ID, msgID string, msgData []byte) error
}

type EvtPullDepositGroupMessage struct {
	DepositAddress peer.ID
	GroupID        string
	MessageHandler func(fromGroupID string, msgID string, msgData []byte) error
}
