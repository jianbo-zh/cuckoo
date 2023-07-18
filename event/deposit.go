package event

import "github.com/libp2p/go-libp2p/core/peer"

type PushOfflineMessageEvt struct {
	ToPeerID peer.ID
	MsgID    string
	MsgData  []byte
}

type PullOfflineMessageEvt struct {
	HasMessage  func(peerID peer.ID, msgID string) (bool, error)
	SaveMessage func(peerID peer.ID, msgID string, msgData []byte) error
}
