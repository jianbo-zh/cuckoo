package types

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type Account struct {
	ID                   peer.ID
	Name                 string
	Avatar               string
	AutoAddContact       bool
	AutoJoinGroup        bool
	AutoSendDeposit      bool
	DepositPeerID        peer.ID
	EnableDepositService bool
}

type AccountPeer struct {
	ID            peer.ID
	Name          string
	Avatar        string
	DepositPeerID peer.ID
}

type PeerState struct {
	PeerID   peer.ID
	IsOnline bool
}

type AccountGetter interface {
	GetAccount(ctx context.Context) (*Account, error)
}
