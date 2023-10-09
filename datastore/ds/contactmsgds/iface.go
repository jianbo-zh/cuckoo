package contactmsgds

import (
	"context"

	ds "github.com/ipfs/go-datastore"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/contactpb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerMessageIface interface {
	ds.Batching

	SaveMessage(ctx context.Context, peerID peer.ID, msg *pb.ContactMessage) (isLatest bool, err error)
	GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*pb.ContactMessage, error)
	GetCoreMessage(ctx context.Context, peerID peer.ID, msgID string) (*pb.ContactMessage_CoreMessage, error)
	UpdateMessageSendState(ctx context.Context, peerID peer.ID, msgID string, isSucc bool) (*pb.ContactMessage, error)
	DeleteMessage(ctx context.Context, peerID peer.ID, msgID string) error
	GetMessageData(ctx context.Context, peerID peer.ID, msgID string) ([]byte, error)
	GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]*pb.ContactMessage, error)
	HasMessage(ctx context.Context, peerID peer.ID, msgID string) (bool, error)
	ClearMessage(ctx context.Context, peerID peer.ID) error

	GetLamportTime(ctx context.Context, peerID peer.ID) (uint64, error)
	TickLamportTime(ctx context.Context, peerID peer.ID) (uint64, error)
	MergeLamportTime(ctx context.Context, peerID peer.ID, lamptime uint64) error

	GetMessageHead(ctx context.Context, peerID peer.ID) (string, error)
	GetMessageTail(ctx context.Context, peerID peer.ID) (string, error)
	GetMessageLength(ctx context.Context, peerID peer.ID) (int32, error)
	GetRangeIDs(peerID peer.ID, startID string, endID string) ([]string, error)
	GetRangeMessages(peerID peer.ID, startID string, endID string) ([]*pb.ContactMessage, error)
	GetMessagesByIDs(peerID peer.ID, msgIDs []string) ([]*pb.ContactMessage, error)
}
