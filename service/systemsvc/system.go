package systemsvc

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	myevent "github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/contactsvc"
	"github.com/jianbo-zh/dchat/service/groupsvc"
	"github.com/jianbo-zh/dchat/service/systemsvc/protocol/systemproto"
	"github.com/jianbo-zh/dchat/service/systemsvc/protocol/systemproto/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
)

var log = logging.Logger("system")

type SystemSvc struct {
	host myhost.Host

	systemProto *systemproto.SystemProto

	accountSvc accountsvc.AccountServiceIface
	contactSvc contactsvc.ContactServiceIface
	groupSvc   groupsvc.GroupServiceIface

	msgCh chan *pb.SystemMsg
}

func NewSystemService(ctx context.Context, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus,
	accountSvc accountsvc.AccountServiceIface, contactSvc contactsvc.ContactServiceIface, groupSvc groupsvc.GroupServiceIface) (*SystemSvc, error) {

	var err error

	systemsvc := &SystemSvc{
		host:       lhost,
		msgCh:      make(chan *pb.SystemMsg, 5),
		accountSvc: accountSvc,
		contactSvc: contactSvc,
		groupSvc:   groupSvc,
	}

	systemsvc.systemProto, err = systemproto.NewSystemProto(lhost, ids, systemsvc.msgCh)
	if err != nil {
		return nil, fmt.Errorf("peerpeer.NewAccountSvc error: %s", err.Error())
	}

	sub, err := ebus.Subscribe([]any{new(myevent.EvtInviteJoinGroup), new(myevent.EvtApplyAddContact)}, eventbus.Name("send_system_message"))
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
			case myevent.EvtApplyAddContact: // 申请添加联系人
				account, err := s.accountSvc.GetAccount(ctx)
				if err != nil {
					log.Errorf("get account error: %s", err.Error())
					continue
				}

				msg := pb.SystemMsg{
					Id:         GenMsgID(account.ID),
					SystemType: mytype.SystemTypeApplyAddContact,
					Group:      nil,
					FromPeer: &pb.SystemMsg_Peer{
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

				if err := s.systemProto.SaveMessage(ctx, &msg); err != nil {
					log.Errorf("systemProto.SaveMessage error: %v", err)
					continue
				}

				if err = s.systemProto.SendMessage(ctx, &msg); err != nil {
					log.Errorf("systemProto.SendMessage error: %v", err)
					continue
				}

			case myevent.EvtInviteJoinGroup: // 邀请加入群组
				account, err := s.accountSvc.GetAccount(ctx)
				if err != nil {
					log.Errorf("get account error: %s", err.Error())
					continue
				}

				for _, peerID := range evt.PeerIDs {
					msg := pb.SystemMsg{
						Id:         GenMsgID(account.ID),
						SystemType: mytype.SystemTypeInviteJoinGroup,
						Group: &pb.SystemMsg_Group{
							Id:       evt.GroupID,
							Name:     evt.GroupName,
							Avatar:   evt.GroupAvatar,
							Lamptime: evt.GroupLamptime,
						},
						FromPeer: &pb.SystemMsg_Peer{
							PeerId: []byte(account.ID),
							Name:   account.Name,
							Avatar: account.Avatar,
						},
						ToPeerId:    []byte(peerID),
						Content:     "",
						SystemState: mytype.SystemStateSended,
						CreateTime:  time.Now().Unix(),
						UpdateTime:  time.Now().Unix(),
					}

					if err := s.systemProto.SaveMessage(ctx, &msg); err != nil {
						log.Errorf("systemProto.SaveMessage error: %v", err)
						continue
					}

					fmt.Println("send system message: ", msg.String())

					err = s.systemProto.SendMessage(ctx, &msg)
					if err != nil {
						log.Errorf("systemProto.SendMessage error: %v", err)
						continue
					}
				}
			}

		case <-ctx.Done():
			return
		}
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
			account, err := s.accountSvc.GetAccount(ctx)
			if err != nil {
				log.Errorf("get account error: %s", err.Error())
				continue
			}

			if account.AutoAddContact {
				// 添加为好友
				fmt.Println("auto add contact avatar: ", msg.FromPeer.Avatar)
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
			account, err := s.accountSvc.GetAccount(ctx)
			if err != nil {
				log.Errorf("get account error: %s", err.Error())
				continue
			}

			if account.AutoJoinGroup {

				// 创建群组
				if err = s.groupSvc.AgreeJoinGroup(ctx, msg.Group.Id, msg.Group.Name, msg.Group.Avatar, msg.Group.Lamptime); err != nil {
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

func (s *SystemSvc) ApplyAddContact(ctx context.Context, peer0 *mytype.Peer, content string) error {
	account, err := s.accountSvc.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("accountSvc.GetAccount error: %w", err)
	}

	if err = s.contactSvc.ApplyAddContact(ctx, peer0, content); err != nil {
		return fmt.Errorf("svc apply add contact error: %w", err)
	}

	msg := pb.SystemMsg{
		Id:         GenMsgID(account.ID),
		SystemType: mytype.SystemTypeApplyAddContact,
		Group:      nil,
		FromPeer: &pb.SystemMsg_Peer{
			PeerId: []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		ToPeerId:    []byte(peer0.ID),
		Content:     content,
		SystemState: mytype.SystemStateSended,
		CreateTime:  time.Now().Unix(),
		UpdateTime:  time.Now().Unix(),
	}

	if err := s.systemProto.SaveMessage(ctx, &msg); err != nil {
		return fmt.Errorf("systemProto.SaveMessage error: %w", err)
	}

	if err = s.systemProto.SendMessage(ctx, &msg); err != nil {
		return fmt.Errorf("systemProto.SendMessage error: %w", err)
	}

	return nil
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
	// 发送同意加好友消息
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

func (s *SystemSvc) Close() {}

func GenMsgID(peerID peer.ID) string {
	return fmt.Sprintf("%d-%s", time.Now().UnixMilli(), peerID.String())
}
