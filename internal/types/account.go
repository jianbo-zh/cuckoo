package types

import "github.com/libp2p/go-libp2p/core/peer"

type Account struct {
	ID             peer.ID
	Name           string
	Avatar         string
	AutoAddContact bool
	AutoJoinGroup  bool
}
