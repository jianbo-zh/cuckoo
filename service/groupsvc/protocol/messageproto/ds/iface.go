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

	GetMessage(context.Context, GroupID, string) (*pb.Message, error)
	SaveMessage(context.Context, GroupID, *pb.Message) error
	ListMessages(ctx context.Context, groupID GroupID, offset int, limit int) ([]*pb.Message, error)
	ClearMessage(ctx context.Context, groupID GroupID) error

	GetMessageHead(context.Context, GroupID) (string, error)
	GetMessageTail(context.Context, GroupID) (string, error)
	GetMessageLength(context.Context, GroupID) (int32, error)

	GetRangeMessages(string, string, string) ([]*pb.Message, error)
	GetRangeIDs(string, string, string) ([]string, error)
	GetMessagesByIDs(string, []string) ([]*pb.Message, error)
}
