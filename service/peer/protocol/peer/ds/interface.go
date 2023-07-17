package ds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/peer/protocol/peer/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerIface interface {
	ipfsds.Batching

	SavePeer(context.Context, *pb.PeerInfo) error
	GetPeer(context.Context, peer.ID) (*pb.PeerInfo, error)
	GetPeers(context.Context) ([]*pb.PeerInfo, error)
	GetPeerIDs(context.Context) ([]peer.ID, error)
	DeletePeer(context.Context, peer.ID) error
}
