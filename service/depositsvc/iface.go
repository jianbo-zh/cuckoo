package depositsvc

import "github.com/libp2p/go-libp2p/core/peer"

type DepositServiceIface interface {
	PushContactMessage(depositPeerID peer.ID, toPeerID peer.ID, msgID string, msgData []byte) error
	PushGroupMessage(depositPeerID peer.ID, groupID string, msgID string, msgData []byte) error
	Close()
}
