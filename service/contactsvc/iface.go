package contactsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

type ContactServiceIface interface {
	ApplyAddContact(ctx context.Context, peer0 *mytype.Peer, content string) error
	AgreeAddContact(ctx context.Context, peer0 *mytype.Peer) error
	GetContact(ctx context.Context, peerID peer.ID) (*mytype.Contact, error)
	GetContacts(ctx context.Context) ([]mytype.Contact, error)
	GetContactsByPeerIDs(ctx context.Context, peerIDs []peer.ID) ([]mytype.Contact, error)
	DeleteContact(ctx context.Context, peerID peer.ID) error
	SetContactName(ctx context.Context, peerID peer.ID, name string) error

	GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*mytype.ContactMessage, error)
	DeleteMessage(ctx context.Context, peerID peer.ID, msgID string) error
	GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]mytype.ContactMessage, error)
	SendMessage(ctx context.Context, peerID peer.ID, msgType string, mimeType string, payload []byte, attachmentID string, file *mytype.FileInfo) (resultCh <-chan mytype.ContactMessage, err error)
	ClearMessage(ctx context.Context, peerID peer.ID) error

	Close()
}
