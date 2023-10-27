package groupmsgds

import (
	"context"

	ds "github.com/ipfs/go-datastore"
	pb "github.com/jianbo-zh/dchat/service/groupsvc/protobuf/pb/grouppb"
)

type MessageIface interface {
	ds.Batching

	GetLamportTime(ctx context.Context, groupID string) (uint64, error)
	MergeLamportTime(ctx context.Context, groupID string, lamptime uint64) error
	TickLamportTime(ctx context.Context, groupID string) (uint64, error)

	HasMessage(ctx context.Context, groupID string, msgID string) (bool, error)
	GetCoreMessage(ctx context.Context, groupID string, msgID string) (*pb.CoreMessage, error)
	GetMessage(ctx context.Context, groupID string, msgID string) (*pb.GroupMessage, error)
	GetMessageData(ctx context.Context, groupID string, msgID string) ([]byte, error)
	DeleteMessage(ctx context.Context, groupID string, msgID string) error
	SaveMessage(ctx context.Context, groupID string, pbmsg *pb.GroupMessage) (isLatest bool, err error)
	GetMessages(ctx context.Context, groupID string, offset int, limit int) ([]*pb.GroupMessage, error)
	ClearMessage(ctx context.Context, groupID string) error
	UpdateMessageSendState(ctx context.Context, groupID string, msgID string, isDeposit bool, isSucc bool) (*pb.GroupMessage, error)

	GetMessageHead(ctx context.Context, groupID string) (string, error)
	GetMessageTail(ctx context.Context, groupID string) (string, error)
	GetMessageLength(ctx context.Context, groupID string) (int32, error)

	GetRangeMessages(string, string, string) ([]*pb.GroupMessage, error)
	GetRangeIDs(string, string, string) ([]string, error)
	GetMessagesByIDs(string, []string) ([]*pb.GroupMessage, error)
}
