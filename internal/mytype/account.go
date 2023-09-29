package mytype

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type Account struct {
	ID                 peer.ID
	Name               string
	Avatar             string
	AutoAddContact     bool
	AutoJoinGroup      bool
	AutoDepositMessage bool
	DepositAddress     peer.ID
}

type AccountPeer struct {
	ID             peer.ID
	Name           string
	Avatar         string
	DepositAddress peer.ID
}

type PeerState struct {
	PeerID   peer.ID
	IsOnline bool
}

type AccountGetter interface {
	GetAccount(ctx context.Context) (*Account, error)
}
