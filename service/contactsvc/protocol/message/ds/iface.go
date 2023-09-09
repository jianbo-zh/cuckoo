package ds

import (
	"context"

	ds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/message/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerMessageIface interface {
	ds.Batching

	SaveMessage(ctx context.Context, peerID peer.ID, msg *pb.Message) error
	GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*pb.Message, error)
	GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]*pb.Message, error)
	HasMessage(ctx context.Context, peerID peer.ID, msgID string) (bool, error)

	GetLamportTime(ctx context.Context, peerID peer.ID) (uint64, error)
	TickLamportTime(ctx context.Context, peerID peer.ID) (uint64, error)
	MergeLamportTime(ctx context.Context, peerID peer.ID, lamptime uint64) error

	GetMessageHead(ctx context.Context, peerID peer.ID) (string, error)
	GetMessageTail(ctx context.Context, peerID peer.ID) (string, error)
	GetMessageLength(ctx context.Context, peerID peer.ID) (int32, error)
	GetRangeIDs(peerID peer.ID, startID string, endID string) ([]string, error)
	GetRangeMessages(peerID peer.ID, startID string, endID string) ([]*pb.Message, error)
	GetMessagesByIDs(peerID peer.ID, msgIDs []string) ([]*pb.Message, error)
}
