package contactsvc

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

type ContactServiceIface interface {
	GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*Message, error)
	GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]Message, error)
	SendTextMessage(ctx context.Context, peerID peer.ID, content string) error
	SendGroupInviteMessage(ctx context.Context, peerID peer.ID, content string) error

	AddContact(ctx context.Context, peerID peer.ID, name string, avatar string) error
	GetContact(ctx context.Context, peerID peer.ID) (*Contact, error)
	GetContacts(ctx context.Context) ([]Contact, error)
	GetContactsByPeerIDs(ctx context.Context, peerIDs []peer.ID) ([]Contact, error)

	Close()
}

type MsgType int

const (
	MsgTypeText MsgType = iota
	MsgTypeAudio
	MsgTypeVideo
	MsgTypeInvite
)

type Message struct {
	ID         string  `json:"id"`
	MsgType    MsgType `json:"msg_type"`
	MimeType   string  `json:"mime_type"`
	FromPeerID peer.ID `json:"from_peer_id"`
	ToPeerID   peer.ID `json:"to_peer_id"`
	Payload    []byte  `json:"payload"`
	Timestamp  int64   `json:"timestamp"`
	Lamportime uint64  `json:"lamportime"`
}

type Peer struct {
	PeerID peer.ID
	Name   string
	Avatar string
}

type Contact struct {
	PeerID   peer.ID
	Name     string
	Avatar   string
	AddTs    int64
	AccessTs int64
}
