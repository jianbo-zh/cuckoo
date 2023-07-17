package peer

import "github.com/libp2p/go-libp2p/core/peer"

type PeerInfo struct {
	PeerID   peer.ID
	Nickname string
	AddTs    int64
	AccessTs int64
}
