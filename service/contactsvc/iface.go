package contactsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/types"
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
	AddContact(ctx context.Context, peerID peer.ID, name string, avatar string) error
	GetContact(ctx context.Context, peerID peer.ID) (*Contact, error)
	GetContacts(ctx context.Context) ([]Contact, error)
	GetContactsByPeerIDs(ctx context.Context, peerIDs []peer.ID) ([]Contact, error)
	DeleteContact(ctx context.Context, peerID peer.ID) error
	SetContactName(ctx context.Context, peerID peer.ID, name string) error

	GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*Message, error)
	GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]Message, error)
	SendMessage(ctx context.Context, peerID peer.ID, msgType types.MsgType, mimeType string, payload []byte) error
	ClearMessage(ctx context.Context, peerID peer.ID) error

	Close()
}

type Message struct {
	ID         string
	MsgType    types.MsgType
	MimeType   string
	FromPeerID peer.ID
	ToPeerID   peer.ID
	Payload    []byte
	Timestamp  int64
	Lamportime uint64
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
