package groupsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/admin/ds"
	"github.com/libp2p/go-libp2p/core/peer"
)

type GroupServiceIface interface {
	CreateGroup(ctx context.Context, name string, avatar string, memberIDs []peer.ID) (groupID string, err error) // 创建群
	JoinGroup(ctx context.Context, groupID string, groupName string, groupAvatar string, lamptime uint64) error   // 加入群
	DisbandGroup(ctx context.Context, groupID string) error                                                       // 解散群
	ExitGroup(ctx context.Context, groupID string) error                                                          // 退出群
	DeleteGroup(ctx context.Context, groupID string) error                                                        // 删除群
	ListGroups(ctx context.Context) ([]ds.Group, error)                                                           // 群列表

	GetGroup(ctx context.Context, groupID string) (*Group, error) // 获取群组

	GroupName(ctx context.Context, groupID string) (string, error)            // 群备注/名称
	SetGroupName(ctx context.Context, groupID string, name string) error      // 设置群名称
	SetGroupLocalName(ctx context.Context, groupID string, name string) error // 设置群备注

	SetGroupAvatar(ctx context.Context, groupID string, avatar string) error      // 设置群头像
	SetGroupLocalAvatar(ctx context.Context, groupID string, avatar string) error // 设置群本地头像

	GroupNotice(ctx context.Context, groupID string) (string, error)         // 群公告
	SetGroupNotice(ctx context.Context, groupID string, notice string) error // 设置群公告

	InviteMember(ctx context.Context, groupID string, peerID peer.ID) error                 // 邀请进群
	ApplyMember(ctx context.Context, groupID string) error                                  // 申请进群
	ReviewMember(ctx context.Context, groupID string, memberID peer.ID, isAgree bool) error // 进群审核
	RemoveMember(ctx context.Context, groupID string, memberID peer.ID) error               // 移除成员
	ListMembers(ctx context.Context, groupID string) ([]peer.ID, error)                     // 成员列表

	SendMessage(ctx context.Context, groupID string, msgType string, mimeType string, payload []byte) error // 发送消息
	ListMessages(ctx context.Context, groupID string, offset int, limit int) ([]Message, error)             // 消息列表

	Close()
}

type Group struct {
	ID     string
	Name   string
	Avatar string
}
type ListGroupsParam struct{}
type GroupNameParam struct{}
type SetGroupNameParam struct{}
type SetGroupLocalNameParam struct{}
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

type MsgType int

const (
	MsgTypeText MsgType = iota
	MsgTypeAudio
	MsgTypeVideo
	MsgTypeInvite
)

type Message struct {
	ID         string
	GroupID    string
	MsgType    MsgType
	MimeType   string
	FromPeer   Peer
	Payload    []byte
	Timestamp  int64
	Lamportime uint64
}

type Peer struct {
	PeerID peer.ID
	Name   string
	Avatar string
}

type SendMessageParam struct{}
type ListMessagesParam struct{}
