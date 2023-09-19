package ds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/adminproto/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type AdminIface interface {
	ipfsds.Batching

	GetLamptime(ctx context.Context, groupID string) (uint64, error)
	MergeLamptime(ctx context.Context, groupID string, lamptime uint64) error
	TickLamptime(ctx context.Context, groupID string) (uint64, error)

	SaveLog(ctx context.Context, log *pb.Log) error

	GetState(ctx context.Context, groupID string) (string, error)
	GetName(ctx context.Context, groupID string) (string, error)
	GetAvatar(ctx context.Context, groupID string) (string, error)
	GetNotice(ctx context.Context, groupID string) (string, error)
	GetAutoJoinGroup(ctx context.Context, groupID string) (bool, error)
	GetCreator(ctx context.Context, groupID string) (peer.ID, error)
	GetCreateTime(ctx context.Context, groupID string) (int64, error)
	GetSessionIDs(ctx context.Context) ([]string, error)
	GetMembers(ctx context.Context, groupID string) ([]types.GroupMember, error)      // 正式成员
	GetAgreeMembers(ctx context.Context, groupID string) ([]types.GroupMember, error) // 所有审核通过的成员

	SetState(ctx context.Context, groupID string, state string) error
	SetName(ctx context.Context, groupID string, name string) error
	SetAvatar(ctx context.Context, groupID string, avatar string) error
	SetAlias(ctx context.Context, groupID string, alias string) error
	SetSession(ctx context.Context, groupID string) error

	DeleteSession(ctx context.Context, groupID string) error
	DeleteGroup(ctx context.Context, groupID string) error

	GetLogHead(ctx context.Context, groupID string) (string, error)
	GetLogTail(ctx context.Context, groupID string) (string, error)
	GetLogLength(ctx context.Context, groupID string) (int, error)

	GetRangeLogs(groupID string, startID string, endID string) ([]*pb.Log, error)
	GetRangeLogIDs(groupID string, startID string, endID string) ([]string, error)
	GetLogsByIDs(string, []string) ([]*pb.Log, error)
}
