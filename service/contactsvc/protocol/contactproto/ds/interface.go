package ds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/contactproto/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerIface interface {
	ipfsds.Batching

	AddContact(context.Context, *pb.ContactMsg) error
	GetContact(context.Context, peer.ID) (*pb.ContactMsg, error)
	GetContacts(context.Context) ([]*pb.ContactMsg, error)
	GetContactsByIDs(ctx context.Context, peerIDs []peer.ID) ([]*pb.ContactMsg, error)
	UpdateContact(context.Context, *pb.ContactMsg) error
	DeleteContact(context.Context, peer.ID) error

	GetState(ctx context.Context, peerID peer.ID) (string, error)
	SetState(ctx context.Context, peerID peer.ID, state string) error
	SetSession(ctx context.Context, peerID peer.ID) error
	GetSessionIDs(ctx context.Context) ([]peer.ID, error)
}
