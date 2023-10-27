package systemds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	pb "github.com/jianbo-zh/dchat/service/systemsvc/protobuf/pb/systempb"
)

type SystemIface interface {
	ipfsds.Batching

	AddSystemMessage(ctx context.Context, msg *pb.SystemMessage) error
	GetSystemMessage(ctx context.Context, msgID string) (*pb.SystemMessage, error)
	GetSystemMessageList(ctx context.Context, offset int, limit int) ([]*pb.SystemMessage, error)
	UpdateSystemMessageState(ctx context.Context, msgID string, state string) error
	DeleteSystemMessage(ctx context.Context, msgIDs []string) error
}
