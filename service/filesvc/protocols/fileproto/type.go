package fileproto

import "github.com/libp2p/go-libp2p/core/peer"

type PeerQueryResult struct {
	PeerID     peer.ID
	FileExists bool
}
