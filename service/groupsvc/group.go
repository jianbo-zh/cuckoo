package groupsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/protocol/groupadminproto"
	"github.com/jianbo-zh/dchat/protocol/groupmsgproto"
	"github.com/jianbo-zh/dchat/protocol/groupnetworkproto"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/contactsvc"
	"github.com/jianbo-zh/dchat/service/sessionsvc"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var log = logging.Logger("group-service")

type GroupService struct {
	networkProto *groupnetworkproto.NetworkProto
	adminProto   *groupadminproto.AdminProto
	messageProto *groupmsgproto.MessageProto

	accountSvc accountsvc.AccountServiceIface
	contactSvc contactsvc.ContactServiceIface

	emitters struct {
		evtGroupsInit event.Emitter
	}
}

func NewGroupService(ctx context.Context, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus,
	rdiscvry *drouting.RoutingDiscovery, accountSvc accountsvc.AccountServiceIface,
	contactSvc contactsvc.ContactServiceIface, sessionSvc sessionsvc.SessionServiceIface) (*GroupService, error) {

	var err error

	groupsvc := &GroupService{
		accountSvc: accountSvc,
		contactSvc: contactSvc,
	}

	groupsvc.adminProto, err = groupadminproto.NewAdminProto(lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("admin.NewAdminService %s", err.Error())
	}

	groupsvc.networkProto, err = groupnetworkproto.NewNetworkProto(lhost, rdiscvry, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("network.NewNetworkService %s", err.Error())
	}

	groupsvc.messageProto, err = groupmsgproto.NewMessageProto(lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("network.NewNetworkService %s", err.Error())
	}

	// 触发器
	groupsvc.emitters.evtGroupsInit, err = ebus.Emitter(&myevent.EvtGroupsInit{})
	if err != nil {
		return nil, fmt.Errorf("ebus.Emitter: %s", err.Error())
	}

	// 订阅器
	sub, err := ebus.Subscribe([]any{new(myevent.EvtHostBootComplete)})
	if err != nil {
		return nil, fmt.Errorf("subscribe boot complete error: %v", err)

	} else {
		go groupsvc.handleSubscribe(context.Background(), sub)
	}

	return groupsvc, nil
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
			case myevent.EvtHostBootComplete:
				if evt.IsSucc {
					groupIDs, err := g.adminProto.GetGroupIDs(ctx)
					if err != nil {
						log.Errorf("get group ids error: %v", err)
						return
					}
					var groups []myevent.Groups
					for _, groupID := range groupIDs {
						connMemberIDs, err := g.adminProto.GetMemberIDs(ctx, groupID)
						if err != nil {
							log.Errorf("get member ids error: %v", err)
						}
						acptMemeberIDs, err := g.adminProto.GetAgreeMemberIDs(ctx, groupID)
						if err != nil {
							log.Errorf("get accept member ids error: %v", err)
							return
						}
						groups = append(groups, myevent.Groups{
							GroupID:     groupID,
							PeerIDs:     connMemberIDs,
							AcptPeerIDs: acptMemeberIDs,
						})
					}

					err = g.emitters.evtGroupsInit.Emit(myevent.EvtGroupsInit{
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
func (g *GroupService) CreateGroup(ctx context.Context, name string, avatar string, memberIDs []peer.ID) (*mytype.Group, error) {

	account, err := g.accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get account error: %w", err)
	}

	contacts, err := g.contactSvc.GetContactsByPeerIDs(ctx, memberIDs)
	if err != nil {
		return nil, fmt.Errorf("get contacts by ids error: %w", err)
	}

	return g.adminProto.CreateGroup(ctx, account, name, avatar, contacts)
}

// GetGroup 获取群信息
func (g *GroupService) GetGroup(ctx context.Context, groupID string) (*mytype.Group, error) {
	return g.adminProto.GetGroup(ctx, groupID)
}

// GetGroupDetail 获取群详情
func (g *GroupService) GetGroupDetail(ctx context.Context, groupID string) (*mytype.GroupDetail, error) {
	return g.adminProto.GetGroupDetail(ctx, groupID)
}

// AgreeJoinGroup 同意加入群
func (g *GroupService) AgreeJoinGroup(ctx context.Context, groupID string, groupName string, groupAvatar string, lamptime uint64) error {

	account, err := g.accountSvc.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("svc get account error: %w", err)
	}

	group := &mytype.Group{
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

	return nil
}

// 群列表
func (g *GroupService) GetGroups(ctx context.Context) ([]mytype.Group, error) {
	return g.adminProto.GetSessions(ctx)
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
func (g *GroupService) SetGroupDepositAddress(ctx context.Context, groupID string, depositPeerID peer.ID) error {
	return g.adminProto.SetGroupDepositAddress(ctx, groupID, depositPeerID)
}

// 申请进群
func (g *GroupService) ApplyJoinGroup(ctx context.Context, groupID string) error {
	// 1. 找到群节点（3个）

	// 2. 向群节点发送申请入群消息
	// group.adminSvc.ApplyMember(ctx, groupID, memberID, lamporttime)

	return nil
}

// 进群审核
func (g *GroupService) ReviewJoinGroup(ctx context.Context, groupID string, member *mytype.Peer, isAgree bool) error {
	return g.adminProto.ReviewJoinGroup(ctx, groupID, member, isAgree)
}

// 移除成员
func (g *GroupService) RemoveGroupMember(ctx context.Context, groupID string, memberID peer.ID) error {
	return g.adminProto.RemoveMember(ctx, groupID, memberID)
}

// 成员列表
func (g *GroupService) GetGroupMembers(ctx context.Context, groupID string, keywords string, offset int, limit int) ([]mytype.GroupMember, error) {
	return g.adminProto.GetGroupMembers(ctx, groupID, keywords, offset, limit)
}

// 发送消息
func (g *GroupService) SendGroupMessage(ctx context.Context, groupID string, msgType string, mimeType string, payload []byte) (string, error) {

	account, err := g.accountSvc.GetAccount(ctx)
	if err != nil {
		return "", fmt.Errorf("accountSvc.GetAccount error: %w", err)
	}

	return g.messageProto.SendGroupMessage(ctx, account, groupID, msgType, mimeType, payload)
}

// 获取消息消息
func (g *GroupService) GetGroupMessage(ctx context.Context, groupID string, msgID string) (*mytype.GroupMessage, error) {
	msg, err := g.messageProto.GetMessage(ctx, groupID, msgID)
	if err != nil {
		return nil, fmt.Errorf("messageSvc.GetMessageList error: %w", err)
	}

	message := &mytype.GroupMessage{
		ID:      msg.Id,
		GroupID: msg.GroupId,
		FromPeer: mytype.GroupMember{
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

func (g *GroupService) GetGroupMessageData(ctx context.Context, groupID string, msgID string) ([]byte, error) {
	return g.messageProto.GetMessageData(ctx, groupID, msgID)
}

func (g *GroupService) DeleteGroupMessage(ctx context.Context, groupID string, msgID string) error {
	return g.messageProto.DeleteMessage(ctx, groupID, msgID)
}

// 消息列表
func (g *GroupService) GetGroupMessages(ctx context.Context, groupID string, offset int, limit int) ([]mytype.GroupMessage, error) {
	msgs, err := g.messageProto.GetMessageList(ctx, groupID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("messageSvc.GetMessageList error: %w", err)
	}

	var messageList []mytype.GroupMessage
	for _, msg := range msgs {
		messageList = append(messageList, mytype.GroupMessage{
			ID:       msg.Id,
			GroupID:  msg.GroupId,
			MsgType:  msg.MsgType,
			MimeType: msg.MimeType,
			FromPeer: mytype.GroupMember{
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
