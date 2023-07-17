package ds

import (
	"context"
	"strings"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/service/peer/protocol/peer/pb"
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

func (p *PeerDS) SavePeer(ctx context.Context, info *pb.PeerInfo) error {
	value, err := proto.Marshal(info)
	if err != nil {
		return err
	}

	key := ipfsds.NewKey("/dchat/peer/peer/" + peer.ID(info.PeerId).String())
	return p.Put(ctx, key, value)
}

func (p *PeerDS) GetPeer(ctx context.Context, peerID peer.ID) (*pb.PeerInfo, error) {
	key := ipfsds.NewKey("/dchat/peer/peer/" + peerID.String())

	value, err := p.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var info pb.PeerInfo
	if err := proto.Unmarshal(value, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (p *PeerDS) GetPeers(ctx context.Context) ([]*pb.PeerInfo, error) {
	results, err := p.Query(ctx, query.Query{
		Prefix: "/dchat/peer/peer/",
	})
	if err != nil {
		return nil, err
	}

	var peeris []*pb.PeerInfo
	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		var peeri pb.PeerInfo
		if err := proto.Unmarshal(result.Entry.Value, &peeri); err != nil {
			return nil, err
		}

		peeris = append(peeris, &peeri)
	}

	return peeris, nil
}

func (p *PeerDS) GetPeerIDs(ctx context.Context) ([]peer.ID, error) {
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

func (p *PeerDS) DeletePeer(ctx context.Context, peerID peer.ID) error {
	key := ipfsds.NewKey("/dchat/peer/peer/" + peerID.String())
	return p.Delete(ctx, key)
}
