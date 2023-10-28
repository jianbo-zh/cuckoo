package myevent

import (
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

// EvtAccountBaseChange 账号基本信息改变触发(头像、昵称、寄存地址)
type EvtAccountBaseChange struct {
	AccountPeer mytype.AccountPeer
}

// EvtSyncAccountMessage 同步账号消息（启动成功后，拉取账号寄存的联系人消息）
type EvtSyncAccountMessage struct {
	DepositAddress peer.ID
}

// EvtSyncSystemMessage 同步系统消息（启动完成后：拉取账号系统消息）
type EvtSyncSystemMessage struct {
	DepositAddress peer.ID
}
