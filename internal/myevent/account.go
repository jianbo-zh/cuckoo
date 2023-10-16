package myevent

import (
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

// EvtAccountBaseChange
type EvtAccountBaseChange struct {
	AccountPeer mytype.AccountPeer
}

// EvtSyncAccountMessage 同步账号消息
type EvtSyncAccountMessage struct {
	DepositAddress peer.ID
}
