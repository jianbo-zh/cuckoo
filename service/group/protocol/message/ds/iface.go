package ds

import (
	"context"

	ds "github.com/ipfs/go-datastore"
	msgpb "github.com/jianbo-zh/dchat/service/group/protocol/message/pb"
)

type MessageIface interface {
	ds.Batching

	GetLamportTime(context.Context, GroupID) (uint64, error)
	MergeLamportTime(context.Context, GroupID, uint64) error
	TickLamportTime(context.Context, GroupID) (uint64, error)

	GetMessage(context.Context, GroupID, string) (*msgpb.GroupMsg, error)
	PutMessage(context.Context, GroupID, *msgpb.GroupMsg) error
	ListMessages(context.Context, GroupID) ([]*msgpb.GroupMsg, error)

	GetMessageHeadID(context.Context, GroupID) (string, error)
	GetMessageTailID(context.Context, GroupID) (string, error)
	GetMessageLength(context.Context, GroupID) (int32, error)
}
