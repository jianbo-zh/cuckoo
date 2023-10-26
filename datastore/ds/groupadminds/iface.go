package groupadminds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/grouppb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type AdminIface interface {
	ipfsds.Batching

	GetLamptime(ctx context.Context, groupID string) (uint64, error)
	MergeLamptime(ctx context.Context, groupID string, lamptime uint64) error
	TickLamptime(ctx context.Context, groupID string) (uint64, error)

	SaveLog(ctx context.Context, log *pb.GroupLog) error
	GetLog(ctx context.Context, groupID string, logID string) (*pb.GroupLog, error)

	UpdateCreator(ctx context.Context, groupID string) error
	UpdateName(ctx context.Context, groupID string) error
	UpdateAvatar(ctx context.Context, groupID string) error
	UpdateNotice(ctx context.Context, groupID string) error
	UpdateAutoJoinGroup(ctx context.Context, groupID string) error
	UpdateDepositAddress(ctx context.Context, groupID string) error
	UpdateMembers(ctx context.Context, groupID string, hostID peer.ID) error
	UpdateDisband(ctx context.Context, groupID string) error

	GetState(ctx context.Context, groupID string) (string, error)
	GetName(ctx context.Context, groupID string) (string, error)
	GetAvatar(ctx context.Context, groupID string) (string, error)
	GetNotice(ctx context.Context, groupID string) (string, error)
	GetAutoJoinGroup(ctx context.Context, groupID string) (bool, error)
	GetDepositAddress(ctx context.Context, groupID string) (peer.ID, error)
	GetCreator(ctx context.Context, groupID string) (peer.ID, error)
	GetCreateTime(ctx context.Context, groupID string) (int64, error)
	GetGroupIDs(ctx context.Context) ([]string, error)
	GetMembers(ctx context.Context, groupID string) ([]*pb.GroupMember, error) // 正式成员

	GetLogIDs(ctx context.Context, groupID string) ([]string, error)                   // 所有日志IDs
	GetMemberIDs(ctx context.Context, groupID string) ([]peer.ID, error)               // 正式成员IDs
	GetAgreePeerIDs(ctx context.Context, groupID string) ([]peer.ID, error)            // 所有审核通过的成员IDs
	GetRefusePeerLogs(ctx context.Context, groupID string) (map[peer.ID]string, error) // 所有拒绝连接的成员

	SetState(ctx context.Context, groupID string, state string) error
	SetName(ctx context.Context, groupID string, name string) error
	SetAvatar(ctx context.Context, groupID string, avatar string) error
	SetAlias(ctx context.Context, groupID string, alias string) error
	SetListID(ctx context.Context, groupID string) error

	DeleteListID(ctx context.Context, groupID string) error
	DeleteGroup(ctx context.Context, groupID string) error

	GetLogHead(ctx context.Context, groupID string) (string, error)
	GetLogTail(ctx context.Context, groupID string) (string, error)
	GetLogLength(ctx context.Context, groupID string) (int, error)

	GetRangeLogs(groupID string, startID string, endID string) ([]*pb.GroupLog, error)
	GetRangeLogIDs(groupID string, startID string, endID string) ([]string, error)
	GetLogsByIDs(string, []string) ([]*pb.GroupLog, error)
}
