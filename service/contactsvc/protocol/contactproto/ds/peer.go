package ds

import (
	"context"
	"strings"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/contactproto/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

var _ PeerIface = (*PeerDS)(nil)

type PeerDS struct {
	ipfsds.Batching
}

func Wrap(b ipfsds.Batching) *PeerDS {
	return &PeerDS{Batching: b}
}

func (p *PeerDS) AddContact(ctx context.Context, info *pb.ContactMsg) error {
	value, err := proto.Marshal(info)
	if err != nil {
		return err
	}

	key := ipfsds.NewKey("/dchat/peer/peer/" + peer.ID(info.PeerId).String())
	return p.Put(ctx, key, value)
}

func (p *PeerDS) GetContact(ctx context.Context, peerID peer.ID) (*pb.ContactMsg, error) {
	key := ipfsds.NewKey("/dchat/peer/peer/" + peerID.String())

	value, err := p.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var info pb.ContactMsg
	if err := proto.Unmarshal(value, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (p *PeerDS) GetContacts(ctx context.Context) ([]*pb.ContactMsg, error) {
	results, err := p.Query(ctx, query.Query{
		Prefix: "/dchat/peer/peer/",
	})
	if err != nil {
		return nil, err
	}

	var peeris []*pb.ContactMsg
	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		var peeri pb.ContactMsg
		if err := proto.Unmarshal(result.Entry.Value, &peeri); err != nil {
			return nil, err
		}

		peeris = append(peeris, &peeri)
	}

	return peeris, nil
}

func (p *PeerDS) GetContactIDs(ctx context.Context) ([]peer.ID, error) {
	prefix := "/dchat/peer/peer/"
	results, err := p.Query(ctx, query.Query{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	var peerIDs []peer.ID
	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		peerID, err := peer.Decode(strings.TrimLeft(result.Entry.Key, prefix))
		if err != nil {
			return nil, err
		}

		peerIDs = append(peerIDs, peerID)
	}

	return peerIDs, nil
}

func (p *PeerDS) DeleteContact(ctx context.Context, peerID peer.ID) error {
	key := ipfsds.NewKey("/dchat/peer/peer/" + peerID.String())
	return p.Delete(ctx, key)
}
