package ds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/admin/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type AdminIface interface {
	ipfsds.Batching

	GetLamportTime(context.Context, GroupID) (uint64, error)
	MergeLamportTime(context.Context, GroupID, uint64) error
	TickLamportTime(context.Context, GroupID) (uint64, error)

	SaveLog(context.Context, peer.ID, GroupID, *pb.Log) error
	ListGroups(context.Context) ([]Group, error)
	GetGroupIDs(context.Context) ([]string, error)

	JoinGroupSaveLog(context.Context, peer.ID, GroupID, *pb.Log) error
	JoinGroup(ctx context.Context, groupID string, name string, avatar string) error
	DeleteGroup(ctx context.Context, groupID string) error
	GetGroup(ctx context.Context, groupID string) (*Group, error)

	GroupName(context.Context, GroupID) (string, error)
	GroupLocalName(context.Context, GroupID) (string, error)
	GroupAvatar(context.Context, GroupID) (string, error)
	GroupLocalAvatar(context.Context, GroupID) (string, error)
	GroupNotice(context.Context, GroupID) (string, error)
	SetGroupLocalName(context.Context, GroupID, string) error
	SetGroupLocalAvatar(context.Context, GroupID, string) error
	GroupMemberLogs(context.Context, GroupID) ([]*pb.Log, error)

	GetMessageHead(context.Context, GroupID) (string, error)
	GetMessageTail(context.Context, GroupID) (string, error)
	GetMessageLength(context.Context, GroupID) (int32, error)

	GetRangeMessages(string, string, string) ([]*pb.Log, error)
	GetRangeIDs(string, string, string) ([]string, error)
	GetMessagesByIDs(string, []string) ([]*pb.Log, error)
}
