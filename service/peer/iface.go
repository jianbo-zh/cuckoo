package peer

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

// 联系人相关服务

/*

通讯录-----
查看通讯录
查找联系人
添加进通讯录
发起聊天
查看聊天内容
新消息（未读）

*/

type PeerServiceIface interface {
	GetMessages(context.Context, peer.ID, int, int) ([]PeerMessage, error)
	SendTextMessage(context.Context, peer.ID, string) error
	SendGroupInviteMessage(context.Context, peer.ID, string) error
}

type MsgType int

const (
	MsgTypeText MsgType = iota
	MsgTypeAudio
	MsgTypeVideo
	MsgTypeInvite
)

type PeerMessage struct {
	ID         string  `json:"id"`
	Type       MsgType `json:"type"`
	SenderID   peer.ID `json:"sender_id"`
	ReceiverID peer.ID `json:"receiver_id"`
	Payload    []byte  `json:"payload"`
	Timestamp  int64   `json:"timestamp"`
	Lamportime uint64  `json:"lamportime"`
}
