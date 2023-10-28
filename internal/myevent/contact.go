package myevent

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

// EvtSyncContactMessage 同步联系人消息（程序启动后）
type EvtSyncContactMessage struct {
	ContactID peer.ID
}

// EvtContactAdded 已添加联系人
type EvtContactAdded struct {
	ID             peer.ID
	Name           string
	Avatar         string
	DepositAddress peer.ID
}
