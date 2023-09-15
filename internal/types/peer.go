package types

import "github.com/libp2p/go-libp2p/core/peer"

type Peer struct {
	ID     peer.ID
	Name   string
	Avatar string
}
