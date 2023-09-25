package ds

import (
	"context"

	ds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/messageproto/pb"
)

type MessageIface interface {
	ds.Batching

	GetLamportTime(context.Context, GroupID) (uint64, error)
	MergeLamportTime(context.Context, GroupID, uint64) error
	TickLamportTime(context.Context, GroupID) (uint64, error)

	GetMessage(ctx context.Context, groupID GroupID, msgID string) (*pb.Message, error)
	GetMessageData(ctx context.Context, groupID GroupID, msgID string) ([]byte, error)
	DeleteMessage(ctx context.Context, groupID GroupID, msgID string) error
	SaveMessage(context.Context, GroupID, *pb.Message) error
	GetMessages(ctx context.Context, groupID GroupID, offset int, limit int) ([]*pb.Message, error)
	ClearMessage(ctx context.Context, groupID GroupID) error

	GetMessageHead(context.Context, GroupID) (string, error)
	GetMessageTail(context.Context, GroupID) (string, error)
	GetMessageLength(context.Context, GroupID) (int32, error)

	GetRangeMessages(string, string, string) ([]*pb.Message, error)
	GetRangeIDs(string, string, string) ([]string, error)
	GetMessagesByIDs(string, []string) ([]*pb.Message, error)
}
