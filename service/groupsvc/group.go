package groupsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/admin"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/admin/ds"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/message"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/network"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

type GroupService struct {
	networkSvc *network.NetworkService
	adminSvc   *admin.AdminService
	messageSvc *message.MessageService

	accountSvc accountsvc.AccountServiceIface

	emitters struct {
		evtGroupsInit event.Emitter
	}
}

func NewGroupService(ctx context.Context, conf config.GroupServiceConfig, lhost host.Host, ids ipfsds.Batching,
	ebus event.Bus, rdiscvry *drouting.RoutingDiscovery, accountSvc accountsvc.AccountServiceIface) (*GroupService, error) {
	var err error

	groupsvc = &GroupService{
		accountSvc: accountSvc,
	}

	groupsvc.adminSvc, err = admin.NewAdminService(lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("admin.NewAdminService %s", err.Error())
	}

	groupsvc.networkSvc, err = network.NewNetworkService(lhost, rdiscvry, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("network.NewNetworkService %s", err.Error())
	}

	groupsvc.messageSvc, err = message.NewMessageService(lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("network.NewNetworkService %s", err.Error())
	}

	// 触发器
	groupsvc.emitters.evtGroupsInit, err = ebus.Emitter(&gevent.EvtGroupsInit{})
	if err != nil {
		return nil, fmt.Errorf("ebus.Emitter: %s", err.Error())
	}

	// 订阅器
	sub, err := ebus.Subscribe([]any{new(gevent.EvtHostBootComplete)})
	if err != nil {
		return nil, fmt.Errorf("subscribe boot complete error: %v", err)

	} else {
		go groupsvc.handleSubscribe(context.Background(), sub)
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

func (group *GroupService) handleSubscribe(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}
			switch evt := e.(type) {
			case gevent.EvtHostBootComplete:
				if evt.IsSucc {
					groupIDs, err := group.adminSvc.GetGroupIDs(ctx)
					if err != nil {
						fmt.Printf("GetGroupIDs error: %s", err.Error())
						return
					}
					var groups []gevent.Groups
					for _, groupID := range groupIDs {
						memeberIDs, err := group.adminSvc.ListMembers(ctx, groupID)
						if err != nil {
							fmt.Println("ListMembers error: %s", err.Error())
							return
						}
						groups = append(groups, gevent.Groups{
							GroupID: groupID,
							PeerIDs: memeberIDs,
						})
					}

					fmt.Printf("groups: %v\n", groups)

					err = groupsvc.emitters.evtGroupsInit.Emit(gevent.EvtGroupsInit{
						Groups: groups,
					})
					if err != nil {
						fmt.Printf("emitters.evtGroupsInit error: %s", err.Error())
					}
					return
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// 创建群
func (group *GroupService) CreateGroup(ctx context.Context, name string, avatar string, memberIDs []peer.ID) (string, error) {
	return group.adminSvc.CreateGroup(ctx, name, avatar, memberIDs)
}

// 加入群
func (group *GroupService) JoinGroup(ctx context.Context, groupID string, groupName string, groupAvatar string, lamptime uint64) error {
	return group.adminSvc.JoinGroup(ctx, groupID, groupName, groupAvatar, lamptime)
}

func (group *GroupService) GetGroup(ctx context.Context, groupID string) (*Group, error) {
	grp, err := group.adminSvc.GetGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("adminSvc.GetGroup error: %w", err)
	}

	return &Group{
		ID:     grp.ID,
		Name:   grp.Name,
		Avatar: grp.Avatar,
	}, nil
}

// 解散群
func (group *GroupService) DisbandGroup(ctx context.Context, groupID string) error {

	if err := group.adminSvc.DisbandGroup(ctx, groupID); err != nil {
		return err
	}

	// todo: 广播其他节点
	return nil
}

// 退出群
func (group *GroupService) ExitGroup(ctx context.Context, groupID string) error {

	if err := group.adminSvc.ExitGroup(ctx, groupID); err != nil {
		return err
	}

	// todo: 广播其他节点
	return nil
}

// 退出并删除群
func (group *GroupService) DeleteGroup(ctx context.Context, groupID string) error {

	if err := group.adminSvc.DeleteGroup(ctx, groupID); err != nil {
		return fmt.Errorf("adminSvc delete group error: %w", err)
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

// 设置群头像
func (group *GroupService) SetGroupAvatar(ctx context.Context, groupID string, avatar string) error {
	return group.adminSvc.SetGroupAvatar(ctx, groupID, avatar)
}

// 设置群本地名称
func (group *GroupService) SetGroupLocalName(ctx context.Context, groupID string, name string) error {
	return group.adminSvc.SetGroupLocalName(ctx, groupID, name)
}

// 设置群本地头像
func (group *GroupService) SetGroupLocalAvatar(ctx context.Context, groupID string, avatar string) error {
	return group.adminSvc.SetGroupLocalAvatar(ctx, groupID, avatar)
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
func (group *GroupService) ListMembers(ctx context.Context, groupID string) ([]peer.ID, error) {
	memberIDs, err := group.adminSvc.ListMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return memberIDs, nil
}

// 发送消息
func (group *GroupService) SendMessage(ctx context.Context, groupID string, msgType types.MsgType, mimeType string, payload []byte) error {

	account, err := group.accountSvc.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("accountSvc.GetAccount error: %w", err)
	}

	if err := group.messageSvc.SendTextMessage(ctx, groupID, account.Name, account.Avatar, string(payload)); err != nil {
		return fmt.Errorf("group.messageSvc.SendTextMessage error: %w", err)
	}
	return nil
}

// 消息列表
func (group *GroupService) ListMessages(ctx context.Context, groupID string, offset int, limit int) ([]Message, error) {
	msgs, err := group.messageSvc.GetMessageList(ctx, groupID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("messageSvc.GetMessageList error: %w", err)
	}
	var messageList []Message
	for _, msg := range msgs {
		messageList = append(messageList, Message{
			ID:       msg.Id,
			GroupID:  msg.GroupId,
			MsgType:  decodeMsgType(msg.MsgType),
			MimeType: msg.MimeType,
			FromPeer: Peer{
				PeerID: peer.ID(msg.PeerId),
				Name:   msg.PeerName,
				Avatar: msg.PeerAvatar,
			},
			Payload:    msg.Payload,
			Timestamp:  msg.Timestamp,
			Lamportime: 0,
		})
	}

	return messageList, nil
}
