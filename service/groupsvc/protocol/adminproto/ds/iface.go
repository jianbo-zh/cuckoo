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

	GetLamportTime(ctx context.Context, groupID string) (uint64, error)
	MergeLamportTime(ctx context.Context, groupID string, lamptime uint64) error
	TickLamportTime(ctx context.Context, groupID string) (uint64, error)

	SaveLog(ctx context.Context, peerID peer.ID, groupID string, log *pb.Log) error
	GetGroups(ctx context.Context) ([]Group, error)
	GetGroupIDs(ctx context.Context) ([]string, error)

	GetState(ctx context.Context, groupID string) (string, error)

	JoinGroupSaveLog(ctx context.Context, peerID peer.ID, groupID string, log *pb.Log) error
	JoinGroup(ctx context.Context, group *types.Group) error
	DeleteGroup(ctx context.Context, groupID string) error
	GetGroup(ctx context.Context, groupID string) (*Group, error)

	GroupName(ctx context.Context, groupID string) (string, error)
	GroupLocalName(ctx context.Context, groupID string) (string, error)
	GroupAvatar(ctx context.Context, groupID string) (string, error)
	GroupLocalAvatar(ctx context.Context, groupID string) (string, error)
	GroupNotice(ctx context.Context, groupID string) (string, error)
	SetGroupLocalName(ctx context.Context, groupID string, name string) error
	SetGroupLocalAvatar(ctx context.Context, groupID string, avatar string) error
	GroupMemberLogs(ctx context.Context, groupID string) ([]*pb.Log, error)

	GetMessageHead(ctx context.Context, groupID string) (string, error)
	GetMessageTail(ctx context.Context, groupID string) (string, error)
	GetMessageLength(ctx context.Context, groupID string) (int32, error)

	GetRangeMessages(groupID string, startID string, endID string) ([]*pb.Log, error)
	GetRangeIDs(groupID string, startID string, endID string) ([]string, error)
	GetMessagesByIDs(string, []string) ([]*pb.Log, error)
}
