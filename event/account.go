package event

import "github.com/jianbo-zh/dchat/internal/types"

type EvtAccountPeerChange struct {
	AccountPeer types.AccountPeer
}

type EvtAccountDepositServiceChange struct {
	Enable bool
}
