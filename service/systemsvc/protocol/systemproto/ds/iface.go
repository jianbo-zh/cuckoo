package ds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/systemsvc/protocol/systemproto/pb"
)

type SystemIface interface {
	ipfsds.Batching

	AddSystemMessage(ctx context.Context, msg *pb.SystemMsg) error
	GetSystemMessage(ctx context.Context, msgID string) (*pb.SystemMsg, error)
	GetSystemMessageList(ctx context.Context, offset int, limit int) ([]*pb.SystemMsg, error)
	UpdateSystemMessageState(ctx context.Context, msgID string, state string) error
	DeleteSystemMessage(ctx context.Context, msgIDs []string) error
}
