package groupsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/admin"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/admin/ds"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/network"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

type GroupService struct {
	adminSvc *admin.AdminService

	networkSvc *network.NetworkService
}

func NewGroupService(conf config.GroupMessageConfig, lhost host.Host, ids ipfsds.Batching, ebus event.Bus, rdiscvry *drouting.RoutingDiscovery, opts ...Option) (*GroupService, error) {
	var err error

	groupsvc = &GroupService{}

	if err := groupsvc.Apply(opts...); err != nil {
		return nil, err
	}

	groupsvc.adminSvc, err = admin.NewAdminService(lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("admin.NewAdminService %s", err.Error())
	}

	groupsvc.networkSvc, err = network.NewNetworkService(lhost, rdiscvry, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("network.NewNetworkService %s", err.Error())
	}

	return groupsvc, nil
}

func Get() GroupServiceIface {
	if groupsvc == nil {
		panic("group must init before use")
	}

	return groupsvc
}

// 关闭服务
func (group *GroupService) Close() {}

// 创建群
func (group *GroupService) CreateGroup(ctx context.Context, name string, memberIDs []peer.ID) (string, error) {
	return group.adminSvc.CreateGroup(ctx, name, memberIDs)
}

// 解散群
func (group *GroupService) DisbandGroup(ctx context.Context, groupID string) error {

	if err := group.adminSvc.DisbandGroup(ctx, groupID); err != nil {
		return err
	}

	// todo: 广播其他节点
	return nil
}

// 群列表
func (group *GroupService) ListGroups(ctx context.Context) ([]ds.Group, error) {
	return group.adminSvc.ListGroups(ctx)
}

// 群名称
func (group *GroupService) GroupName(ctx context.Context, groupID string) (string, error) {
	return group.adminSvc.GroupName(ctx, groupID)
}

// 设置群名称
func (group *GroupService) SetGroupName(ctx context.Context, groupID string, name string) error {
	return group.adminSvc.SetGroupName(ctx, groupID, name)
}

// 设置群备注
func (group *GroupService) SetGroupRemark(ctx context.Context, groupID string, remark string) error {
	return group.adminSvc.SetGroupRemark(ctx, groupID, remark)
}

// 群公告
func (group *GroupService) GroupNotice(ctx context.Context, groupID string) (string, error) {
	return group.adminSvc.GroupNotice(ctx, groupID)
}

// 设置群公告
func (group *GroupService) SetGroupNotice(ctx context.Context, groupID string, notice string) error {
	return group.adminSvc.SetGroupNotice(ctx, groupID, notice)
}

// 邀请进群
func (group *GroupService) InviteMember(ctx context.Context, groupID string, peerID peer.ID) error {
	if err := group.adminSvc.InviteMember(ctx, groupID, peerID); err != nil {
		return err
	}

	// todo: 群广播操作
	return nil
}

// 申请进群
func (group *GroupService) ApplyMember(ctx context.Context, groupID string) error {
	// 1. 找到群节点（3个）

	// 2. 向群节点发送申请入群消息
	// group.adminSvc.ApplyMember(ctx, groupID, memberID, lamporttime)

	return nil
}

// 进群审核
func (group *GroupService) ReviewMember(ctx context.Context, groupID string, memberID peer.ID, isAgree bool) error {
	return group.adminSvc.ReviewMember(ctx, groupID, memberID, isAgree)
}

// 移除成员
func (group *GroupService) RemoveMember(ctx context.Context, groupID string, memberID peer.ID) error {
	return group.adminSvc.RemoveMember(ctx, groupID, memberID)
}

// 成员列表
func (group *GroupService) ListMembers(ctx context.Context, groupID string) ([]Member, error) {
	members, err := group.adminSvc.ListMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var mms []Member
	for _, member := range members {
		mms = append(mms, Member{
			PeerID: member.PeerID,
		})
	}

	return mms, nil
}

// 发送消息
func (group *GroupService) SendMessage(SendMessageParam) error {
	return nil
}

// 消息列表
func (group *GroupService) ListMessages(ListMessagesParam) ([]Message, error) {
	return nil, nil
}
