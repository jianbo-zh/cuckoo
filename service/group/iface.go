package group

import (
	"context"

	"github.com/jianbo-zh/dchat/service/group/datastore"
	"github.com/libp2p/go-libp2p/core/peer"
)

type GroupServiceIface interface {
	CreateGroup(ctx context.Context, name string, memberIDs []peer.ID) (string, error) // 创建群
	DisbandGroup(ctx context.Context, groupID string) error                            // 解散群
	ListGroups(ctx context.Context) ([]datastore.Group, error)                         // 群列表

	GroupName(ctx context.Context, groupID string) (string, error)           // 群备注/名称
	SetGroupName(ctx context.Context, groupID string, name string) error     // 设置群名称
	SetGroupRemark(ctx context.Context, groupID string, remark string) error // 设置群备注

	GroupNotice(ctx context.Context, groupID string) (string, error)         // 群公告
	SetGroupNotice(ctx context.Context, groupID string, notice string) error // 设置群公告

	InviteMember(ctx context.Context, groupID string, peerID peer.ID) error                 // 邀请进群
	ApplyMember(ctx context.Context, groupID string) error                                  // 申请进群
	ReviewMember(ctx context.Context, groupID string, memberID peer.ID, isAgree bool) error // 进群审核
	RemoveMember(ctx context.Context, groupID string, memberID peer.ID) error               // 移除成员
	ListMembers(ctx context.Context, groupID string) ([]Member, error)                      // 成员列表

	SendMessage(SendMessageParam) error                // 发送消息
	ListMessages(ListMessagesParam) ([]Message, error) // 消息列表
}

type Group struct{}
type ListGroupsParam struct{}
type GroupNameParam struct{}
type SetGroupNameParam struct{}
type SetGroupRemarkParam struct{}
type GroupNoticeParam struct{}
type SetGroupNoticeParam struct{}

type Member struct {
	PeerID peer.ID
}
type InviteMemberParam struct{}
type ApplyMemberParam struct{}
type ReviewMemberParam struct{}
type RemoveMemberParam struct{}
type ListMembersParam struct{}

type Message struct{}
type SendMessageParam struct{}
type ListMessagesParam struct{}
