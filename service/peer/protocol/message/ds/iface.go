package ds

import (
	"context"

	ds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/peer/protocol/message/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerMessageIface interface {
	ds.Batching

	SaveMessage(context.Context, peer.ID, *pb.Message) error
	GetMessages(context.Context, peer.ID, int, int) ([]*pb.Message, error)
	HasMessage(context.Context, peer.ID, string) (bool, error)

	GetLamportTime(context.Context, peer.ID) (uint64, error)
	TickLamportTime(context.Context, peer.ID) (uint64, error)
	MergeLamportTime(context.Context, peer.ID, uint64) error

	GetMessageHead(ctx context.Context, peerID peer.ID) (string, error)
	GetMessageTail(ctx context.Context, peerID peer.ID) (string, error)
	GetMessageLength(ctx context.Context, peerID peer.ID) (int32, error)
	GetRangeIDs(peerID peer.ID, startID string, endID string) ([]string, error)
	GetRangeMessages(peerID peer.ID, startID string, endID string) ([]*pb.Message, error)
	GetMessagesByIDs(peerID peer.ID, msgIDs []string) ([]*pb.Message, error)
}
