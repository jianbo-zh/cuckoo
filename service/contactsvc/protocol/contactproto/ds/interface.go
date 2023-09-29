package ds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/contactproto/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerIface interface {
	ipfsds.Batching

	AddContact(context.Context, *pb.Contact) error
	GetContact(context.Context, peer.ID) (*pb.Contact, error)
	GetContacts(context.Context) ([]*pb.Contact, error)
	GetContactsByIDs(ctx context.Context, peerIDs []peer.ID) ([]*pb.Contact, error)
	UpdateContact(context.Context, *pb.Contact) error
	DeleteContact(context.Context, peer.ID) error

	SetApply(ctx context.Context, peerID peer.ID) error
	DeleteApply(ctx context.Context, peerID peer.ID) error
	GetApplyIDs(ctx context.Context) ([]peer.ID, error)

	SetSession(ctx context.Context, peerID peer.ID) error
	GetSessionIDs(ctx context.Context) ([]peer.ID, error)
}
