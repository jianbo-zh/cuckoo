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
	"github.com/jianbo-zh/dchat/service/contactsvc"
	"github.com/jianbo-zh/dchat/service/sessionsvc"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var log = logging.Logger("group-service")

type GroupService struct {
	host myhost.Host

	networkProto *groupnetworkproto.NetworkProto
	adminProto   *groupadminproto.AdminProto
	messageProto *groupmsgproto.MessageProto

	contactSvc contactsvc.ContactServiceIface

	accountGetter mytype.AccountGetter

	emitters struct {
		evtGroupsInit              event.Emitter
		evtLogSessionAttachment    event.Emitter
		evtPushDepositGroupMessage event.Emitter
		evtClearSessionResources   event.Emitter
		evtClearSessionFiles       event.Emitter
		evtClearSession            event.Emitter
		evtDeleteSession           event.Emitter
	}
}

func NewGroupService(ctx context.Context, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus,
	rdiscvry *drouting.RoutingDiscovery, accountGetter mytype.AccountGetter,
	contactSvc contactsvc.ContactServiceIface, sessionSvc sessionsvc.SessionServiceIface) (*GroupService, error) {

	var err error

	groupsvc := &GroupService{
		host:          lhost,
		contactSvc:    contactSvc,
		accountGetter: accountGetter,
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

	if groupsvc.emitters.evtLogSessionAttachment, err = ebus.Emitter(&myevent.EvtLogSessionAttachment{}); err != nil {
		return nil, fmt.Errorf("set send resource request emitter error: %w", err)
	}

	// 触发器：发送离线消息
	if groupsvc.emitters.evtPushDepositGroupMessage, err = ebus.Emitter(&myevent.EvtPushDepositGroupMessage{}); err != nil {
		return nil, fmt.Errorf("set pull deposit msg emitter error: %v", err)
	}

	// 触发器：删除会话资源
	if groupsvc.emitters.evtClearSessionResources, err = ebus.Emitter(&myevent.EvtClearSessionResources{}); err != nil {
		return nil, fmt.Errorf("set clear session resources emitter error: %w", err)
	}

	// 触发器：删除会话文件
	if groupsvc.emitters.evtClearSessionFiles, err = ebus.Emitter(&myevent.EvtClearSessionFiles{}); err != nil {
		return nil, fmt.Errorf("set clear session files emitter error: %w", err)
	}

	// 触发器：清空会话
	if groupsvc.emitters.evtClearSession, err = ebus.Emitter(&myevent.EvtClearSession{}); err != nil {
		return nil, fmt.Errorf("set clear session emitter error: %w", err)
	}

	// 触发器：删除会话
	if groupsvc.emitters.evtDeleteSession, err = ebus.Emitter(&myevent.EvtDeleteSession{}); err != nil {
		return nil, fmt.Errorf("set delete session emitter error: %w", err)
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
						acptPeerIDs, err := g.adminProto.GetAgreePeerIDs(ctx, groupID)
						if err != nil {
							log.Errorf("get accept member ids error: %v", err)
							return
						}
						groups = append(groups, myevent.Groups{
							GroupID:     groupID,
							PeerIDs:     connMemberIDs,
							AcptPeerIDs: acptPeerIDs,
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
func (g *GroupService) CreateGroup(ctx context.Context, name string, avatar string, content string, memberIDs []peer.ID) (*mytype.Group, error) {

	account, err := g.accountGetter.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get account error: %w", err)
	}

	contacts, err := g.contactSvc.GetContactsByPeerIDs(ctx, memberIDs)
	if err != nil {
		return nil, fmt.Errorf("get contacts by ids error: %w", err)
	}

	return g.adminProto.CreateGroup(ctx, account, name, avatar, content, contacts)
}

// GetGroup 获取群信息
func (g *GroupService) GetGroup(ctx context.Context, groupID string) (*mytype.Group, error) {
	return g.adminProto.GetGroup(ctx, groupID)
}

// GetGroupDetail 获取群详情
func (g *GroupService) GetGroupDetail(ctx context.Context, groupID string) (*mytype.GroupDetail, error) {
	return g.adminProto.GetGroupDetail(ctx, groupID)
}

// GetGroupMemberIDs 获取群成员列表
func (g *GroupService) GetGroupMemberIDs(ctx context.Context, groupID string) ([]peer.ID, error) {
	return g.adminProto.GetMemberIDs(ctx, groupID)
}

// GetGroupOnlineMemberIDs 获取群在线成员IDs
func (g *GroupService) GetGroupOnlineMemberIDs(ctx context.Context, groupID string) ([]peer.ID, error) {
	return g.networkProto.GetGroupOnlinePeers(groupID)
}

// InviteJoinGroup 邀请进群
func (g *GroupService) InviteJoinGroup(ctx context.Context, groupID string, contactIDs []peer.ID, content string) error {

	account, err := g.accountGetter.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("svc get account error: %w", err)
	}

	contacts, err := g.contactSvc.GetContactsByPeerIDs(ctx, contactIDs)
	if err != nil {
		return fmt.Errorf("svc.GetContact error: %w", err)
	}

	fmt.Println("222")

	return g.adminProto.InviteJoinGroup(ctx, account, groupID, contacts, content)
}

// AgreeJoinGroup 同意进群
func (g *GroupService) AgreeJoinGroup(ctx context.Context, groupID string, groupName string, groupAvatar string, groupLog []byte) error {

	account, err := g.accountGetter.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("svc get account error: %w", err)
	}

	group := &mytype.Group{
		ID:     groupID,
		Name:   groupName,
		Avatar: groupAvatar,
	}

	return g.adminProto.AgreeJoinGroup(ctx, account, group, groupLog)
}

// ExitGroup 退出群
func (g *GroupService) ExitGroup(ctx context.Context, groupID string) error {

	account, err := g.accountGetter.GetAccount(ctx)
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

	account, err := g.accountGetter.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("svc get account error: %w", err)
	}

	// 删除群聊
	if err := g.adminProto.DeleteGroup(ctx, account, groupID); err != nil {
		return fmt.Errorf("adminSvc delete group error: %w", err)
	}

	sessionID := mytype.GroupSessionID(groupID)

	// 删除会话
	resultCh := make(chan error, 1)
	if err := g.emitters.evtDeleteSession.Emit(myevent.EvtDeleteSession{
		SessionID: sessionID.String(),
		Result:    resultCh,
	}); err != nil {
		return fmt.Errorf("emit delete session error: %w", err)
	}
	if err := <-resultCh; err != nil {
		return fmt.Errorf("clear delete session error: %w", err)
	}

	// 删除会话资源
	resultCh1 := make(chan error, 1)
	if err := g.emitters.evtClearSessionResources.Emit(myevent.EvtClearSessionResources{
		SessionID: sessionID.String(),
		Result:    resultCh1,
	}); err != nil {
		return fmt.Errorf("emit clear session resources error: %w", err)
	}
	if err := <-resultCh1; err != nil {
		return fmt.Errorf("clear session resources error: %w", err)
	}

	// 删除会话文件
	resultCh2 := make(chan error, 1)
	if err := g.emitters.evtClearSessionFiles.Emit(myevent.EvtClearSessionFiles{
		SessionID: sessionID.String(),
		Result:    resultCh2,
	}); err != nil {
		return fmt.Errorf("emit clear session files error: %w", err)
	}
	if err := <-resultCh2; err != nil {
		return fmt.Errorf("clear session files error: %w", err)
	}

	// todo: 广播其他节点
	return nil
}

// 解散群
func (g *GroupService) DisbandGroup(ctx context.Context, groupID string) error {

	account, err := g.accountGetter.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("svc get account error: %w", err)
	}

	if err := g.adminProto.DisbandGroup(ctx, account, groupID); err != nil {
		return err
	}

	return nil
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
func (g *GroupService) RemoveGroupMember(ctx context.Context, groupID string, memberIDs []peer.ID) error {
	return g.adminProto.RemoveGroupMember(ctx, groupID, memberIDs)
}

// 成员列表
func (g *GroupService) GetGroupMembers(ctx context.Context, groupID string, keywords string, offset int, limit int) ([]mytype.GroupMember, error) {
	return g.adminProto.GetGroupMembers(ctx, groupID, keywords, offset, limit)
}

// 发送消息
func (g *GroupService) SendGroupMessage(ctx context.Context, groupID string, msgType string, mimeType string, payload []byte,
	attachmentID string, file *mytype.FileInfo) (<-chan mytype.GroupMessage, error) {

	// 处理资源文件
	if attachmentID != "" || file != nil {
		resultCh := make(chan error)
		sessionID := mytype.GroupSessionID(groupID)
		if err := g.emitters.evtLogSessionAttachment.Emit(myevent.EvtLogSessionAttachment{
			SessionID:  sessionID.String(),
			ResourceID: attachmentID,
			File:       file,
			Result:     resultCh,
		}); err != nil {
			return nil, fmt.Errorf("emit record session attachment error: %w", err)
		}
		if err := <-resultCh; err != nil {
			return nil, fmt.Errorf("record session attachment error: %w", err)
		}
	}

	account, err := g.accountGetter.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("accountSvc.GetAccount error: %w", err)
	}

	// 创建消息
	msg, err := g.messageProto.CreateMessage(ctx, account, groupID, msgType, mimeType, payload, attachmentID)
	if err != nil {
		return nil, fmt.Errorf("generate message error: %w", err)
	}

	resultCh := make(chan mytype.GroupMessage, 1)
	resultCh <- *convertMessage(msg)

	go func(msgID string) {
		defer func() {
			close(resultCh)
		}()

		isSucc := true
		isDeposit, err := g.sendGroupMessage(ctx, groupID, msgID)
		if err != nil {
			isSucc = false
			// log error
			log.Error("send message error: %v", err)
		}

		msg, err := g.messageProto.UpdateMessageState(ctx, groupID, msgID, isDeposit, isSucc)
		if err != nil {
			// log error
			log.Errorf("msgProto.UpdateMessageState error: %w", err)
			return
		}

		resultCh <- *convertMessage(msg)

	}(msg.Id)

	return resultCh, nil
}

func (g *GroupService) sendGroupMessage(ctx context.Context, groupID string, msgID string) (isDeposit bool, err error) {

	if msgData, err1 := g.messageProto.SendGroupMessage(ctx, groupID, msgID); err1 != nil {
		// 发送失败
		if len(msgData) > 0 {
			// 可能对方不在线
			if account, err2 := g.accountGetter.GetAccount(ctx); err2 != nil {
				return false, fmt.Errorf("get account error: %w", err2)

			} else if account.AutoDepositMessage {
				// 开启了自动寄存
				if group, err3 := g.adminProto.GetGroup(ctx, groupID); err3 != nil {
					return false, fmt.Errorf("proto.GetGroup error: %w", err3)

				} else if group.DepositAddress != peer.ID("") {
					// 群组设置了自动寄存
					resultCh := make(chan error, 1)
					if err4 := g.emitters.evtPushDepositGroupMessage.Emit(myevent.EvtPushDepositGroupMessage{
						DepositAddress: group.DepositAddress,
						ToGroupID:      group.ID,
						MsgID:          msgID,
						MsgData:        msgData,
						Result:         resultCh,
					}); err4 != nil {
						return false, fmt.Errorf("emit EvtPushDepositGroupMessage error: %w", err4)
					}

					if err5 := <-resultCh; err5 != nil {
						// 发送寄存信息失败
						return false, fmt.Errorf("push deposit msg error: %w", err5)
					} else {
						return true, nil
					}
				}
			}
		}

		return false, fmt.Errorf("proto.SendGroupMessage error: %w", err1)
	}

	return false, nil
}

// 获取消息消息
func (g *GroupService) GetGroupMessage(ctx context.Context, groupID string, msgID string) (*mytype.GroupMessage, error) {
	msg, err := g.messageProto.GetMessage(ctx, groupID, msgID)
	if err != nil {
		return nil, fmt.Errorf("messageSvc.GetMessageList error: %w", err)
	}

	message := convertMessage(msg)

	return message, nil
}

func (g *GroupService) GetGroupMessageData(ctx context.Context, groupID string, msgID string) ([]byte, error) {
	return g.messageProto.GetMessageData(ctx, groupID, msgID)
}

// 消息列表
func (g *GroupService) GetGroupMessages(ctx context.Context, groupID string, offset int, limit int) ([]*mytype.GroupMessage, error) {
	msgs, err := g.messageProto.GetMessageList(ctx, groupID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("messageSvc.GetMessageList error: %w", err)
	}

	var messageList []*mytype.GroupMessage
	for _, msg := range msgs {
		messageList = append(messageList, convertMessage(msg))
	}

	return messageList, nil
}

// 消息列表
func (g *GroupService) ClearGroupMessage(ctx context.Context, groupID string) error {

	sessionID := mytype.GroupSessionID(groupID)

	// 删除消息记录
	if err := g.messageProto.ClearGroupMessage(ctx, groupID); err != nil {
		return fmt.Errorf("proto.ClearGroupMessage error: %w", err)
	}

	// 清空会话
	resultCh := make(chan error, 1)
	if err := g.emitters.evtClearSession.Emit(myevent.EvtClearSession{
		SessionID: sessionID.String(),
		Result:    resultCh,
	}); err != nil {
		return fmt.Errorf("emit clear session error: %w", err)
	}
	if err := <-resultCh; err != nil {
		return fmt.Errorf("clear session error: %w", err)
	}

	// 删除会话资源
	resultCh1 := make(chan error, 1)
	if err := g.emitters.evtClearSessionResources.Emit(myevent.EvtClearSessionResources{
		SessionID: sessionID.String(),
		Result:    resultCh1,
	}); err != nil {
		return fmt.Errorf("emit clear session resources error: %w", err)
	}
	if err := <-resultCh1; err != nil {
		return fmt.Errorf("clear session resources error: %w", err)
	}

	return nil
}
