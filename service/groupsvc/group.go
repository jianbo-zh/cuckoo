package groupsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/contactsvc"
	admin "github.com/jianbo-zh/dchat/service/groupsvc/protocol/adminproto"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/adminproto/pb"
	message "github.com/jianbo-zh/dchat/service/groupsvc/protocol/messageproto"
	network "github.com/jianbo-zh/dchat/service/groupsvc/protocol/networkproto"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var log = logging.Logger("group-service")

type GroupService struct {
	networkProto *network.NetworkProto
	adminProto   *admin.AdminProto
	messageProto *message.MessageProto

	accountSvc accountsvc.AccountServiceIface
	contactSvc contactsvc.ContactServiceIface

	emitters struct {
		evtGroupsInit event.Emitter
	}
}

func NewGroupService(ctx context.Context, conf config.GroupServiceConfig, lhost host.Host, ids ipfsds.Batching, ebus event.Bus,
	rdiscvry *drouting.RoutingDiscovery, accountSvc accountsvc.AccountServiceIface, contactSvc contactsvc.ContactServiceIface) (*GroupService, error) {

	var err error

	groupsvc = &GroupService{
		accountSvc: accountSvc,
		contactSvc: contactSvc,
	}

	groupsvc.adminProto, err = admin.NewAdminProto(lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("admin.NewAdminService %s", err.Error())
	}

	groupsvc.networkProto, err = network.NewNetworkProto(lhost, rdiscvry, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("network.NewNetworkService %s", err.Error())
	}

	groupsvc.messageProto, err = message.NewMessageProto(lhost, ids, ebus)
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
func (g *GroupService) Close() {}

func (g *GroupService) handleSubscribe(ctx context.Context, sub event.Subscription) {
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
					groupIDs, err := g.adminProto.GetGroupIDs(ctx)
					if err != nil {
						log.Errorf("get group ids error: %s", err.Error())
						return
					}
					var groups []gevent.Groups
					for _, groupID := range groupIDs {
						memeberIDs, err := g.adminProto.GetMemberIDs(ctx, groupID)
						if err != nil {
							log.Errorf("get member ids error: %s", err.Error())
							return
						}
						groups = append(groups, gevent.Groups{
							GroupID: groupID,
							PeerIDs: memeberIDs,
						})
					}

					err = groupsvc.emitters.evtGroupsInit.Emit(gevent.EvtGroupsInit{
						Groups: groups,
					})
					if err != nil {
						log.Errorf("emit group init error: %s", err.Error())
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
func (g *GroupService) CreateGroup(ctx context.Context, name string, avatar string, memberIDs []peer.ID) (*types.Group, error) {

	contacts, err := g.contactSvc.GetContactsByPeerIDs(ctx, memberIDs)
	if err != nil {
		return nil, fmt.Errorf("get contacts by ids error: %w", err)
	}

	members := make([]*pb.Log_Member, len(contacts))
	for i, contact := range contacts {
		members[i] = &pb.Log_Member{
			Id:     []byte(contact.ID),
			Name:   contact.Name,
			Avatar: contact.Avatar,
		}
	}

	groupID, err := g.adminProto.CreateGroup(ctx, name, avatar, members)
	if err != nil {
		return nil, fmt.Errorf("proto create group error: %w", err)
	}

	return &types.Group{
		ID:     groupID,
		Name:   name,
		Avatar: avatar,
	}, nil
}

func (g *GroupService) GetGroup(ctx context.Context, groupID string) (*types.Group, error) {
	grp, err := g.adminProto.GetGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("adminSvc.GetGroup error: %w", err)
	}

	return &types.Group{
		ID:     grp.ID,
		Name:   grp.Name,
		Avatar: grp.Avatar,
	}, nil
}

func (g *GroupService) GetGroupDetail(ctx context.Context, groupID string) (*types.GroupDetail, error) {
	grp, err := g.adminProto.GetGroupDetail(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("adminSvc.GetGroup error: %w", err)
	}

	return &types.GroupDetail{
		ID:     grp.ID,
		Name:   grp.Name,
		Avatar: grp.Avatar,
	}, nil
}

// AgreeJoinGroup 同意加入群
func (g *GroupService) AgreeJoinGroup(ctx context.Context, groupID string, groupName string, groupAvatar string, lamptime uint64) error {

	account, err := g.accountSvc.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("svc get account error: %w", err)
	}

	group := &types.Group{
		ID:     groupID,
		Name:   groupName,
		Avatar: groupAvatar,
	}

	return g.adminProto.AgreeJoinGroup(ctx, account, group, lamptime)
}

// 退出群
func (g *GroupService) ExitGroup(ctx context.Context, groupID string) error {

	account, err := g.accountSvc.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("svc get account error: %w", err)
	}

	if err := g.adminProto.ExitGroup(ctx, account, groupID); err != nil {
		return err
	}

	// todo: 广播其他节点
	return nil
}

// 退出并删除群
func (g *GroupService) DeleteGroup(ctx context.Context, groupID string) error {

	account, err := g.accountSvc.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("svc get account error: %w", err)
	}

	if err := g.adminProto.DeleteGroup(ctx, account, groupID); err != nil {
		return fmt.Errorf("adminSvc delete group error: %w", err)
	}

	// todo: 广播其他节点
	return nil
}

// 解散群
func (g *GroupService) DisbandGroup(ctx context.Context, groupID string) error {

	account, err := g.accountSvc.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("svc get account error: %w", err)
	}

	if err := g.adminProto.DisbandGroup(ctx, account, groupID); err != nil {
		return err
	}

	// todo: 广播其他节点
	return nil
}

// 群列表
func (g *GroupService) GetGroups(ctx context.Context) ([]types.Group, error) {
	grps, err := g.adminProto.GetGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("proto get groups error: %w", err)
	}

	var groups []types.Group
	for _, group := range grps {
		groups = append(groups, types.Group{
			ID:     group.ID,
			Name:   group.Name,
			Avatar: group.Avatar,
		})
	}

	return groups, nil
}

// 群列表
func (g *GroupService) GetGroupSessions(ctx context.Context) ([]types.GroupSession, error) {
	grps, err := g.adminProto.GetGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("proto get groups error: %w", err)
	}

	var groups []types.GroupSession
	for _, group := range grps {
		groups = append(groups, types.GroupSession{
			ID:     group.ID,
			Name:   group.Name,
			Avatar: group.Avatar,
		})
	}

	return groups, nil
}

// 设置群名称
func (g *GroupService) SetGroupName(ctx context.Context, groupID string, name string) error {
	return g.adminProto.SetGroupName(ctx, groupID, name)
}

// 设置群头像
func (g *GroupService) SetGroupAvatar(ctx context.Context, groupID string, avatar string) error {
	return g.adminProto.SetGroupAvatar(ctx, groupID, avatar)
}

// 设置群公告
func (g *GroupService) SetGroupNotice(ctx context.Context, groupID string, notice string) error {
	return g.adminProto.SetGroupNotice(ctx, groupID, notice)
}

func (g *GroupService) SetGroupAutoJoin(ctx context.Context, groupID string, isAutoJoin bool) error {
	return g.adminProto.SetGroupAutoJoin(ctx, groupID, isAutoJoin)
}

// 申请进群
func (g *GroupService) ApplyJoinGroup(ctx context.Context, groupID string) error {
	// 1. 找到群节点（3个）

	// 2. 向群节点发送申请入群消息
	// group.adminSvc.ApplyMember(ctx, groupID, memberID, lamporttime)

	return nil
}

// 进群审核
func (g *GroupService) ReviewJoinGroup(ctx context.Context, groupID string, member *types.Peer, isAgree bool) error {
	return g.adminProto.ReviewJoinGroup(ctx, groupID, member, isAgree)
}

// 移除成员
func (g *GroupService) RemoveGroupMember(ctx context.Context, groupID string, memberID peer.ID) error {
	return g.adminProto.RemoveMember(ctx, groupID, memberID)
}

// 成员列表
func (g *GroupService) GetGroupMembers(ctx context.Context, groupID string, keywords string, offset int, limit int) ([]types.GroupMember, error) {
	members, err := g.adminProto.GetGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	membersList := make([]types.GroupMember, len(members))
	for _, member := range members {
		membersList = append(membersList, types.GroupMember{
			ID:     peer.ID(member.Id),
			Name:   member.Name,
			Avatar: member.Avatar,
		})
	}

	return membersList, nil
}

// 发送消息
func (g *GroupService) SendGroupMessage(ctx context.Context, groupID string, msgType string, mimeType string, payload []byte) (*types.GroupMessage, error) {

	account, err := g.accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("accountSvc.GetAccount error: %w", err)
	}

	msg, err := g.messageProto.SendGroupMessage(ctx, account, groupID, msgType, mimeType, payload)
	if err != nil {
		return nil, fmt.Errorf("group.messageSvc.SendTextMessage error: %w", err)
	}

	message := &types.GroupMessage{
		ID:      msg.Id,
		GroupID: msg.GroupId,
		FromPeer: types.GroupMember{
			ID:     peer.ID(msg.Member.Id),
			Name:   msg.Member.Name,
			Avatar: msg.Member.Avatar,
		},
		MsgType:    msg.MsgType,
		MimeType:   msg.MimeType,
		Payload:    msg.Payload,
		CreateTime: msg.CreateTime,
	}

	return message, nil
}

// 消息列表
func (g *GroupService) GetGroupMessages(ctx context.Context, groupID string, offset int, limit int) ([]types.GroupMessage, error) {
	msgs, err := g.messageProto.GetMessageList(ctx, groupID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("messageSvc.GetMessageList error: %w", err)
	}

	var messageList []types.GroupMessage
	for _, msg := range msgs {
		messageList = append(messageList, types.GroupMessage{
			ID:       msg.Id,
			GroupID:  msg.GroupId,
			MsgType:  msg.MsgType,
			MimeType: msg.MimeType,
			FromPeer: types.GroupMember{
				ID:     peer.ID(msg.Member.Id),
				Name:   msg.Member.Name,
				Avatar: msg.Member.Avatar,
			},
			Payload:    msg.Payload,
			CreateTime: msg.CreateTime,
		})
	}

	return messageList, nil
}

// 消息列表
func (g *GroupService) ClearGroupMessage(ctx context.Context, groupID string) error {
	return g.messageProto.ClearGroupMessage(ctx, groupID)
}
