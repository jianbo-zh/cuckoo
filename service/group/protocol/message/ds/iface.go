package ds

import (
	"context"

	ds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/group/protocol/message/pb"
)

type MessageIface interface {
	ds.Batching

	GetLamportTime(context.Context, GroupID) (uint64, error)
	MergeLamportTime(context.Context, GroupID, uint64) error
	TickLamportTime(context.Context, GroupID) (uint64, error)

	GetMessage(context.Context, GroupID, string) (*pb.GroupMsg, error)
	SaveMessage(context.Context, GroupID, *pb.GroupMsg) error
	ListMessages(context.Context, GroupID) ([]*pb.GroupMsg, error)

	GetMessageHead(context.Context, GroupID) (string, error)
	GetMessageTail(context.Context, GroupID) (string, error)
	GetMessageLength(context.Context, GroupID) (int32, error)

	GetRangeMessages(string, string, string) ([]*pb.GroupMsg, error)
	GetRangeIDs(string, string, string) ([]string, error)
	GetMessagesByIDs(string, []string) ([]*pb.GroupMsg, error)
}
