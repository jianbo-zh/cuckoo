package systemsvc

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	myevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/contactsvc"
	"github.com/jianbo-zh/dchat/service/groupsvc"
	"github.com/jianbo-zh/dchat/service/systemsvc/protocol/systemproto"
	"github.com/jianbo-zh/dchat/service/systemsvc/protocol/systemproto/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
)

var log = logging.Logger("system")

type SystemSvc struct {
	host host.Host

	systemProto *systemproto.SystemProto

	accountSvc accountsvc.AccountServiceIface
	contactSvc contactsvc.ContactServiceIface
	groupSvc   groupsvc.GroupServiceIface

	msgCh chan *pb.SystemMsg
}

func NewSystemService(ctx context.Context, lhost host.Host, ids ipfsds.Batching, ebus event.Bus,
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

	sub, err := ebus.Subscribe([]any{new(myevent.EvtInviteJoinGroup)}, eventbus.Name("system_message"))
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

			evt := e.(myevent.EvtInviteJoinGroup)

			account, err := s.accountSvc.GetAccount(ctx)
			if err != nil {
				log.Errorf("get account error: %s", err.Error())
				continue
			}

			for _, peerID := range evt.PeerIDs {

				msg := pb.SystemMsg{
					Id:            GenMsgID(account.ID),
					AckId:         "",
					SystemType:    types.SystemTypeInviteJoinGroup,
					GroupId:       evt.GroupID,
					GroupName:     evt.GroupName,
					GroupAvatar:   evt.GroupAvatar,
					GroupLamptime: evt.GroupLamptime,
					FromPeer: &pb.Peer{
						PeerId: []byte(account.ID),
						Name:   account.Name,
						Avatar: account.Avatar,
					},
					ToPeer: &pb.Peer{
						PeerId: []byte(peerID),
						Name:   "",
						Avatar: "",
					},
					Content:     "",
					SystemState: types.SystemStateSended,
					CreateTime:  time.Now().Unix(),
					UpdateTime:  time.Now().Unix(),
				}

				if err := s.systemProto.SaveMessage(ctx, &msg); err != nil {
					log.Errorf("systemProto.SaveMessage error: %s", err.Error())
					continue
				}

				err = s.systemProto.SendMessage(ctx, &msg)
				if err != nil {
					log.Errorf("systemProto.SendMessage error: %s", err.Error())
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func (s *SystemSvc) goHandleMessage() {
	for msg := range s.msgCh {

		toPeerID := peer.ID(msg.ToPeer.PeerId)
		if toPeerID != s.host.ID() {
			// 不是自己的数据
			log.Warn("toPeerID is not equal")
			continue
		}

		switch msg.SystemType {
		case types.SystemTypeApplyAddContact: // 申请加好友
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
				if err = s.contactSvc.AddContact(ctx, peer.ID(msg.FromPeer.PeerId), msg.FromPeer.Name, msg.FromPeer.Avatar); err != nil {
					log.Errorf("contactSvc.AddContact error: %s", err.Error())
					continue
				}
				// 如果是自动加好友，则更新系统消息状态为已同意
				if err = s.systemProto.UpdateMessageState(ctx, msg.Id, types.SystemStateAgreed); err != nil {
					log.Errorf("update message state error: %s", err.Error())
					continue
				}
			}

		case types.SystemTypeInviteJoinGroup:
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
				fmt.Println("groupID ", msg.GroupId, "name ", msg.GroupName, "avatar ", msg.GroupAvatar)
				if err = s.groupSvc.AgreeJoinGroup(ctx, msg.GroupId, msg.GroupName, msg.GroupAvatar, msg.GroupLamptime); err != nil {
					log.Errorf("groupSvc.JoinGroup error: %s", err.Error())
					continue
				}

				// 如果是自动加好友，则更新系统消息状态为已同意
				if err = s.systemProto.UpdateMessageState(ctx, msg.Id, types.SystemStateAgreed); err != nil {
					log.Errorf("update message state error: %s", err.Error())
					continue
				}
			}

		default:
			log.Error("msg type error")
		}
	}
}

func (s *SystemSvc) ApplyAddContact(ctx context.Context, peer0 *types.Peer, content string) error {
	account, err := s.accountSvc.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("accountSvc.GetAccount error: %w", err)
	}

	msg := pb.SystemMsg{
		Id:         GenMsgID(account.ID),
		AckId:      "",
		SystemType: types.SystemTypeApplyAddContact,
		GroupId:    "",
		FromPeer: &pb.Peer{
			PeerId: []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		ToPeer: &pb.Peer{
			PeerId: []byte(peer0.ID),
			Name:   peer0.Name,
			Avatar: peer0.Avatar,
		},
		Content:     content,
		SystemState: types.SystemStateSended,
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

func (s *SystemSvc) GetSystemMessageList(ctx context.Context, offset int, limit int) ([]types.SystemMessage, error) {
	msgs, err := s.systemProto.GetMessageList(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("systemProto.GetMessageList error: %w", err)
	}

	var sysmsgs []types.SystemMessage
	for _, msg := range msgs {

		sysmsgs = append(sysmsgs, types.SystemMessage{
			ID:         msg.Id,
			SystemType: msg.SystemType,
			GroupID:    msg.GroupId,
			Sender: types.Peer{
				ID:     peer.ID(msg.FromPeer.PeerId),
				Name:   msg.FromPeer.Name,
				Avatar: msg.FromPeer.Avatar,
			},
			Receiver: types.Peer{
				ID:     peer.ID(msg.ToPeer.PeerId),
				Name:   msg.ToPeer.Name,
				Avatar: msg.ToPeer.Avatar,
			},
			Content:     msg.Content,
			SystemState: msg.SystemState,
			CreateTime:  msg.CreateTime,
			UpdateTime:  msg.UpdateTime,
		})
	}

	return sysmsgs, nil
}

func (s *SystemSvc) AgreeAddContact(ctx context.Context, ackID string) error {
	// 发送同意加好友消息
	// todo: ...
	if err := s.systemProto.UpdateMessageState(ctx, ackID, types.SystemStateAgreed); err != nil {
		return fmt.Errorf("systemProto.UpdateMessageState error: %w", err)
	}

	return nil
}

func (s *SystemSvc) RejectAddContact(ctx context.Context, ackID string) error {

	if err := s.systemProto.UpdateMessageState(ctx, ackID, types.SystemStateRejected); err != nil {
		return fmt.Errorf("systemProto.UpdateMessageState error: %w", err)
	}

	return nil
}

func (s *SystemSvc) Close() {}

func GenMsgID(peerID peer.ID) string {
	return fmt.Sprintf("%d-%s", time.Now().UnixMilli(), peerID.String())
}
