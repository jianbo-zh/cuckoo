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
	GetContactIDs(context.Context) ([]peer.ID, error)
	DeleteContact(context.Context, peer.ID) error
}
