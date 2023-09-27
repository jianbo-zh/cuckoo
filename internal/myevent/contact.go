package myevent

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

// 开启peer消息同步
type EvtSyncPeerMessage struct {
	ContactID peer.ID
}

type EvtReceivePeerStream struct {
	PeerID peer.ID
}
