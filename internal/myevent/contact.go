package myevent

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

// 开启peer消息同步
type EvtSyncContactMessage struct {
	ContactID peer.ID
}

type EvtContactAdded struct {
	ID             peer.ID
	Name           string
	Avatar         string
	DepositAddress peer.ID
}
