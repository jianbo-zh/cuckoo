package myevent

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

// 开启peer消息同步
type EvtSyncContactMessage struct {
	ContactID peer.ID
}
