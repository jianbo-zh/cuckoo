package event

import "github.com/libp2p/go-libp2p/core/peer"

type PushDepositContactMessageEvt struct {
	DepositPeerID peer.ID
	ToPeerID      peer.ID
	MsgID         string
	MsgData       []byte

	ResultCallback func(toPeerID peer.ID, msgID string, err error)
}

type PushDepositGroupMessageEvt struct {
	DepositPeerID peer.ID
	ToGroupID     string
	MsgID         string
	MsgData       []byte

	ResultCallback func(toGroupID string, msgID string, err error)
}

type PullDepositContactMessageEvt struct {
	DepositPeerID  peer.ID
	MessageHandler func(fromPeerID peer.ID, msgID string, msgData []byte) error
}

type PullDepositGroupMessageEvt struct {
	DepositPeerID  peer.ID
	GroupID        string
	MessageHandler func(fromGroupID string, msgID string, msgData []byte) error
}
