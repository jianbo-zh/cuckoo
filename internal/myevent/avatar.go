package myevent

import "github.com/libp2p/go-libp2p/core/peer"

type EvtCheckAvatar struct {
	Avatar  string
	PeerIDs []peer.ID
}
