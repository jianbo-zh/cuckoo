package systemsvc

import (
	"context"
	"errors"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myerror"
	myevent "github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/service/contactsvc"
	"github.com/jianbo-zh/dchat/service/groupsvc"
	pb "github.com/jianbo-zh/dchat/service/systemsvc/protobuf/pb/systempb"
	"github.com/jianbo-zh/dchat/service/systemsvc/protocols/systemproto"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"google.golang.org/protobuf/proto"
)

var log = logging.Logger("systemsvc")

var _ SystemServiceIface = (*SystemSvc)(nil)

type SystemSvc struct {
	host myhost.Host

	systemProto *systemproto.SystemProto

	accountGetter mytype.AccountGetter

	contactSvc contactsvc.ContactServiceIface
	groupSvc   groupsvc.GroupServiceIface

	msgCh chan *pb.SystemMessage

	emitters struct {
		evtPushDepositSystemMessage event.Emitter
		evtPullDepositSystemMessage event.Emitter
	}
}

func NewSystemService(ctx context.Context, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus,
	accountGetter mytype.AccountGetter, contactSvc contactsvc.ContactServiceIface, groupSvc groupsvc.GroupServiceIface) (*SystemSvc, error) {

	var err error

	systemsvc := &SystemSvc{
		host:          lhost,
		msgCh:         make(chan *pb.SystemMessage, 5),
		accountGetter: accountGetter,
		contactSvc:    contactSvc,
		groupSvc:      groupSvc,
	}

	systemsvc.systemProto, err = systemproto.NewSystemProto(lhost, ids, systemsvc.msgCh)
	if err != nil {
		return nil, fmt.Errorf("peerpeer.NewAccountSvc error: %s", err.Error())
	}

	if systemsvc.emitters.evtPushDepositSystemMessage, err = ebus.Emitter(&myevent.EvtPushDepositSystemMessage{}); err != nil {
		return nil, fmt.Errorf("set send deposit system msg emitter error: %w", err)
	}

	if systemsvc.emitters.evtPullDepositSystemMessage, err = ebus.Emitter(&myevent.EvtPullDepositSystemMessage{}); err != nil {
		return nil, fmt.Errorf("set pull deposit system msg emitter error: %w", err)
	}

	sub, err := ebus.Subscribe([]any{new(myevent.EvtInviteJoinGroup), new(myevent.EvtApplyAddContact), new(myevent.EvtSyncSystemMessage)}, eventbus.Name("send_system_message"))
	if err != nil {
		return nil, fmt.Errorf("subscription failed. group admin server error: %v", err)
	}

	go systemsvc.goSubscribeHandler(ctx, sub)

	// 后台处理系统消息 todo: add context
	go systemsvc.goHandleMessage()

	return systemsvc, nil
}

// goSubscribeHandler 发送系统消息的监听订阅
func (s *SystemSvc) goSubscribeHandler(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}

			switch evt := e.(type) {
			case myevent.EvtApplyAddContact:
				go s.handleApplyAddContactEvent(ctx, evt)

			case myevent.EvtInviteJoinGroup:
				go s.handleInviteJoinGroupEvent(ctx, evt)

			case myevent.EvtSyncSystemMessage:
				go s.handleSyncSystemMessage(ctx, evt)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (s *SystemSvc) handleSyncSystemMessage(ctx context.Context, evt myevent.EvtSyncSystemMessage) {
	if err := s.emitters.evtPullDepositSystemMessage.Emit(myevent.EvtPullDepositSystemMessage{
		DepositAddress: evt.DepositAddress,
		MessageHandler: s.SaveDepositMessage,
	}); err != nil {
		log.Errorf("emit pull deposit system msg error: %v", err)
		return
	}
}

func (s *SystemSvc) SaveDepositMessage(ctx context.Context, fromPeerID peer.ID, msgID string, msgData []byte) error {

	hostID := s.host.ID()

	var msg pb.SystemMessage
	if err := proto.Unmarshal(msgData, &msg); err != nil {
		return fmt.Errorf("proto.Unmarshal deposit system error: %w", err)
	}
	if fromPeerID != peer.ID(msg.FromPeer.PeerId) || hostID != peer.ID(msg.ToPeerId) {
		return fmt.Errorf("from or to peer error")
	}

	log.Debugln("push system msg to msgCh")
	s.msgCh <- &msg

	return nil
}

func (s *SystemSvc) handleApplyAddContactEvent(ctx context.Context, evt myevent.EvtApplyAddContact) {
	var resultErr error
	defer func() {
		evt.Result <- resultErr
		close(evt.Result)
	}()

	account, err := s.accountGetter.GetAccount(ctx)
	if err != nil {
		resultErr = fmt.Errorf("get account error: %w", err)
		return
	}

	msg := pb.SystemMessage{
		Id:         GenMsgID(account.ID),
		SystemType: mytype.SystemTypeApplyAddContact,
		Group:      nil,
		FromPeer: &pb.SystemMessage_Peer{
			PeerId: []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		ToPeerId:    []byte(evt.PeerID),
		Content:     evt.Content,
		SystemState: mytype.SystemStateSended,
		CreateTime:  time.Now().Unix(),
		UpdateTime:  time.Now().Unix(),
	}

	if err = s.sendSystemMessage(ctx, evt.PeerID, evt.DepositAddr, &msg); err != nil {
		resultErr = fmt.Errorf("sendSystemMessage error: %w", err)
		return
	}
}

func (s *SystemSvc) handleInviteJoinGroupEvent(ctx context.Context, evt myevent.EvtInviteJoinGroup) {
	account, err := s.accountGetter.GetAccount(ctx)
	if err != nil {
		log.Errorf("get account error: %s", err.Error())
		return
	}

	for _, invite := range evt.Invites {
		msg := pb.SystemMessage{
			Id:         GenMsgID(account.ID),
			SystemType: mytype.SystemTypeInviteJoinGroup,
			Group: &pb.SystemMessage_Group{
				Id:     evt.GroupID,
				Name:   evt.GroupName,
				Avatar: evt.GroupAvatar,
			},
			FromPeer: &pb.SystemMessage_Peer{
				PeerId: []byte(account.ID),
				Name:   account.Name,
				Avatar: account.Avatar,
			},
			ToPeerId:    []byte(invite.PeerID),
			Content:     evt.Content,
			Payload:     invite.GroupLog,
			SystemState: mytype.SystemStateSended,
			CreateTime:  time.Now().Unix(),
			UpdateTime:  time.Now().Unix(),
		}

		err = s.sendSystemMessage(ctx, invite.PeerID, invite.DepositAddress, &msg)
		if err != nil {
			log.Errorf("systemProto.SendMessage error: %v", err)
			return
		}
	}
}

func (s *SystemSvc) sendSystemMessage(ctx context.Context, toPeerID peer.ID, depositAddr peer.ID, msg *pb.SystemMessage) error {

	onlineState := s.host.OnlineState(toPeerID)

	switch onlineState {
	case mytype.OnlineStateOnline, mytype.OnlineStateUnknown:
		if err := s.systemProto.SendMessage(ctx, msg); err != nil {
			// 发送离线消息
			if errors.As(err, &myerror.StreamErr{}) && depositAddr != peer.ID("") {
				account, err := s.accountGetter.GetAccount(ctx)
				if err != nil {
					return fmt.Errorf("get account error: %w", err)
				}

				if account.AutoDepositMessage {
					msgData, err := proto.Marshal(msg)
					if err != nil {
						return fmt.Errorf("proto.Marshal msg error: %w", err)
					}

					log.Debugln("start deposit ", depositAddr.String())

					resultCh := make(chan error, 1)
					if err := s.emitters.evtPushDepositSystemMessage.Emit(myevent.EvtPushDepositSystemMessage{
						DepositAddress: depositAddr,
						ToPeerID:       toPeerID,
						MsgID:          msg.Id,
						MsgData:        msgData,
						Result:         resultCh,
					}); err != nil {
						return fmt.Errorf("emit push deposit system msg error: %w", err)
					}

					if err := <-resultCh; err != nil {

						log.Debugln("start deposit error")
						return fmt.Errorf("send deposit msg error: %w", err)
					}

					log.Debugln("start deposit end")
					return nil
				}
			}

			return fmt.Errorf("proto.SendMessage error: %w", err)
		}

		return nil

	case mytype.OnlineStateOffline:
		if depositAddr != peer.ID("") {
			account, err := s.accountGetter.GetAccount(ctx)
			if err != nil {
				return fmt.Errorf("get account error: %w", err)
			}

			if account.AutoDepositMessage {
				msgData, err := proto.Marshal(msg)
				if err != nil {
					return fmt.Errorf("proto.Marshal msg error: %w", err)
				}

				resultCh := make(chan error, 1)
				if err := s.emitters.evtPushDepositSystemMessage.Emit(myevent.EvtPushDepositSystemMessage{
					DepositAddress: depositAddr,
					ToPeerID:       toPeerID,
					MsgID:          msg.Id,
					MsgData:        msgData,
					Result:         resultCh,
				}); err != nil {
					return fmt.Errorf("emit push deposit system msg error: %w", err)
				}

				if err := <-resultCh; err != nil {
					return fmt.Errorf("send deposit msg error: %w", err)
				}
				return nil
			}
		}

		return fmt.Errorf("peer is offline")

	default:
		return fmt.Errorf("unsupport online state")
	}
}

func (s *SystemSvc) goHandleMessage() {
	for msg := range s.msgCh {

		toPeerID := peer.ID(msg.ToPeerId)
		if toPeerID != s.host.ID() {
			// 不是自己的数据
			log.Warn("toPeerID is not equal")
			continue
		}

		switch msg.SystemType {
		case mytype.SystemTypeApplyAddContact: // 申请加好友
			ctx := context.Background()
			// 系统消息入库
			if err := s.systemProto.SaveMessage(ctx, msg); err != nil {
				log.Errorf("save message error: %s", err.Error())
				continue
			}

			// 判断是否自动加好友
			account, err := s.accountGetter.GetAccount(ctx)
			if err != nil {
				log.Errorf("get account error: %s", err.Error())
				continue
			}

			if account.AutoAddContact {
				// 添加为好友
				if err = s.contactSvc.AgreeAddContact(ctx, &mytype.Peer{
					ID:     peer.ID(msg.FromPeer.PeerId),
					Name:   msg.FromPeer.Name,
					Avatar: msg.FromPeer.Avatar,
				}); err != nil {
					log.Errorf("contactSvc.AddContact error: %s", err.Error())
					continue
				}
				// 如果是自动加好友，则更新系统消息状态为已同意
				if err = s.systemProto.UpdateMessageState(ctx, msg.Id, mytype.SystemStateAgreed); err != nil {
					log.Errorf("update message state error: %s", err.Error())
					continue
				}
			}

		case mytype.SystemTypeInviteJoinGroup:
			ctx := context.Background()
			// 系统消息入库
			if err := s.systemProto.SaveMessage(ctx, msg); err != nil {
				log.Errorf("save message error: %s", err.Error())
				continue
			}

			// 判断是否自动加好友
			account, err := s.accountGetter.GetAccount(ctx)
			if err != nil {
				log.Errorf("get account error: %s", err.Error())
				continue
			}

			if account.AutoJoinGroup {
				// 创建群组
				if err = s.groupSvc.AgreeJoinGroup(ctx, msg.Group.Id, msg.Group.Name, msg.Group.Avatar, msg.Payload); err != nil {
					log.Errorf("groupSvc.JoinGroup error: %s", err.Error())
					continue
				}

				// 如果是自动加好友，则更新系统消息状态为已同意
				if err = s.systemProto.UpdateMessageState(ctx, msg.Id, mytype.SystemStateAgreed); err != nil {
					log.Errorf("update message state error: %s", err.Error())
					continue
				}
			}

		default:
			log.Error("msg type error")
		}
	}
}

func (s *SystemSvc) GetSystemMessageList(ctx context.Context, offset int, limit int) ([]mytype.SystemMessage, error) {
	msgs, err := s.systemProto.GetMessageList(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("systemProto.GetMessageList error: %w", err)
	}

	var sysmsgs []mytype.SystemMessage
	for _, msg := range msgs {

		sysmsgs = append(sysmsgs, mytype.SystemMessage{
			ID:         msg.Id,
			SystemType: msg.SystemType,
			FromPeer: mytype.Peer{
				ID:     peer.ID(msg.FromPeer.PeerId),
				Name:   msg.FromPeer.Name,
				Avatar: msg.FromPeer.Avatar,
			},
			ToPeerID:    peer.ID(msg.ToPeerId),
			Content:     msg.Content,
			SystemState: msg.SystemState,
			CreateTime:  msg.CreateTime,
			UpdateTime:  msg.UpdateTime,
		})
	}

	return sysmsgs, nil
}

func (s *SystemSvc) AgreeAddContact(ctx context.Context, msgID string) error {

	msg, err := s.systemProto.GetMessage(ctx, msgID)
	if err != nil {
		return fmt.Errorf("proto get message error: %w", err)
	}

	// 添加为好友
	if err := s.contactSvc.AgreeAddContact(ctx, &mytype.Peer{
		ID:     peer.ID(msg.FromPeer.PeerId),
		Name:   msg.FromPeer.Name,
		Avatar: msg.FromPeer.Avatar,
	}); err != nil {
		return fmt.Errorf("contactSvc.AddContact error: %w", err)
	}

	// 更新系统消息状态
	if err := s.systemProto.UpdateMessageState(ctx, msgID, mytype.SystemStateAgreed); err != nil {
		return fmt.Errorf("systemProto.UpdateMessageState error: %w", err)
	}

	return nil
}

func (s *SystemSvc) RejectAddContact(ctx context.Context, ackID string) error {

	if err := s.systemProto.UpdateMessageState(ctx, ackID, mytype.SystemStateRejected); err != nil {
		return fmt.Errorf("systemProto.UpdateMessageState error: %w", err)
	}

	return nil
}

func (s *SystemSvc) AgreeJoinGroup(ctx context.Context, msgID string) error {

	msg, err := s.systemProto.GetMessage(ctx, msgID)
	if err != nil {
		return fmt.Errorf("proto get message error: %w", err)

	} else if msg.SystemType != mytype.SystemTypeInviteJoinGroup {
		return fmt.Errorf("message type is not invite group")
	}

	// 创建群组
	if err = s.groupSvc.AgreeJoinGroup(ctx, msg.Group.Id, msg.Group.Name, msg.Group.Avatar, msg.Payload); err != nil {
		return fmt.Errorf("groupSvc.JoinGroup error: %w", err)

	}

	// 更新系统消息状态
	if err = s.systemProto.UpdateMessageState(ctx, msg.Id, mytype.SystemStateAgreed); err != nil {
		return fmt.Errorf("update message state error: %w", err)
	}

	return nil
}

func (s *SystemSvc) RejectJoinGroup(ctx context.Context, ackID string) error {
	if err := s.systemProto.UpdateMessageState(ctx, ackID, mytype.SystemStateRejected); err != nil {
		return fmt.Errorf("systemProto.UpdateMessageState error: %w", err)
	}

	return nil
}

func (s *SystemSvc) DeleteSystemMessage(ctx context.Context, msgIDs []string) error {
	return s.systemProto.DeleteSystemMessage(ctx, msgIDs)
}

func (s *SystemSvc) Close() {}

func GenMsgID(peerID peer.ID) string {
	return fmt.Sprintf("%d-%s", time.Now().UnixMilli(), peerID.String())
}
