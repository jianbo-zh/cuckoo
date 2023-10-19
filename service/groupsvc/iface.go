package groupsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

type GroupServiceIface interface {
	CreateGroup(ctx context.Context, name string, avatarID string, content string, memberIDs []peer.ID) (*mytype.Group, error) // 创建群
	AgreeJoinGroup(ctx context.Context, groupID string, groupName string, groupAvatar string, lamptime uint64) error           // 同意加入群
	DisbandGroup(ctx context.Context, groupID string) error                                                                    // 解散群
	ExitGroup(ctx context.Context, groupID string) error                                                                       // 退出群
	DeleteGroup(ctx context.Context, groupID string) error                                                                     // 删除群
	GetGroups(ctx context.Context) ([]mytype.Group, error)                                                                     // 群列表

	GetGroup(ctx context.Context, groupID string) (*mytype.Group, error)             // 获取群组
	GetGroupDetail(ctx context.Context, groupID string) (*mytype.GroupDetail, error) // 获取群组

	GetGroupOnlineMemberIDs(ctx context.Context, groupID string) ([]peer.ID, error) // 在线成员IDs

	SetGroupName(ctx context.Context, groupID string, name string) error                     // 设置群名称
	SetGroupAvatar(ctx context.Context, groupID string, avatar string) error                 // 设置群头像
	SetGroupNotice(ctx context.Context, groupID string, notice string) error                 // 设置群公告
	SetGroupAutoJoin(ctx context.Context, groupID string, isAutoJoin bool) error             // 设置入群免确认
	SetGroupDepositAddress(ctx context.Context, groupID string, depositPeerID peer.ID) error // 设置群消息寄存地址

	ApplyJoinGroup(ctx context.Context, groupID string) error                                                                  // 申请进群
	ReviewJoinGroup(ctx context.Context, groupID string, member *mytype.Peer, isAgree bool) error                              // 进群审核
	RemoveGroupMember(ctx context.Context, groupID string, memberID peer.ID) error                                             // 移除成员
	GetGroupMembers(ctx context.Context, groupID string, keywords string, offset int, limit int) ([]mytype.GroupMember, error) // 成员列表

	SendGroupMessage(ctx context.Context, groupID string, msgType string, mimeType string, payload []byte, resourceID string, file *mytype.FileInfo) (resultCh <-chan mytype.GroupMessage, err error) // 发送消息

	GetGroupMessage(ctx context.Context, groupID string, msgID string) (*mytype.GroupMessage, error)             // 获取消息
	GetGroupMessageData(ctx context.Context, groupID string, msgID string) ([]byte, error)                       // 获取消息
	GetGroupMessages(ctx context.Context, groupID string, offset int, limit int) ([]*mytype.GroupMessage, error) // 消息列表
	ClearGroupMessage(ctx context.Context, groupID string) error                                                 // 清空消息

	Close()
}
