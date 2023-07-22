package ds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/group/protocol/admin/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type AdminIface interface {
	ipfsds.Batching

	GetLamportTime(context.Context, GroupID) (uint64, error)
	SetLamportTime(context.Context, GroupID, uint64) error
	TickLamportTime(context.Context, GroupID) (uint64, error)

	LogAdminOperation(context.Context, peer.ID, GroupID, *pb.AdminLog) error
	ListGroups(context.Context) ([]Group, error)

	GroupName(context.Context, GroupID) (string, error)
	GroupRemark(context.Context, GroupID) (string, error)
	GroupNotice(context.Context, GroupID) (string, error)
	SetGroupRemark(context.Context, GroupID, string) error
	GroupMemberLogs(context.Context, GroupID) ([]*pb.AdminLog, error)
}
