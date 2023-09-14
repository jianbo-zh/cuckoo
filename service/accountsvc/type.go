package accountsvc

import "github.com/libp2p/go-libp2p/core/peer"

type Account struct {
	PeerID         peer.ID
	Name           string
	Avatar         string
	AutoAddContact bool
	AutoJoinGroup  bool
}

type Peer struct {
	PeerID peer.ID
	Name   string
	Avatar string
}
