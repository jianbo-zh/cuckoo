package contactsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/libp2p/go-libp2p/core/peer"
)

type ContactServiceIface interface {
	ApplyAddContact(ctx context.Context, peer0 *types.Peer, content string) error
	AgreeAddContact(ctx context.Context, peer0 *types.Peer) error
	GetContact(ctx context.Context, peerID peer.ID) (*types.Contact, error)
	GetContacts(ctx context.Context) ([]types.Contact, error)
	GetContactSessions(ctx context.Context) ([]types.ContactSession, error)
	GetContactsByPeerIDs(ctx context.Context, peerIDs []peer.ID) ([]types.Contact, error)
	DeleteContact(ctx context.Context, peerID peer.ID) error
	SetContactName(ctx context.Context, peerID peer.ID, name string) error

	GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*types.ContactMessage, error)
	GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]types.ContactMessage, error)
	SendMessage(ctx context.Context, peerID peer.ID, msgType string, mimeType string, payload []byte) error
	ClearMessage(ctx context.Context, peerID peer.ID) error

	Close()
}
