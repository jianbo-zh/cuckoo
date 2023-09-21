package event

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

// 开启peer消息同步
type EvtSyncPeers struct {
	ContactIDs []peer.ID
}

type EvtReceivePeerStream struct {
	PeerID peer.ID
}