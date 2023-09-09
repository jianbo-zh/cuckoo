package systemsvc

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/contactsvc"
	"github.com/jianbo-zh/dchat/service/systemsvc/protocol/systemproto"
	"github.com/jianbo-zh/dchat/service/systemsvc/protocol/systemproto/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

var log = logging.Logger("system")

type SystemSvc struct {
	host host.Host

	systemProto *systemproto.SystemProto

	accountSvc accountsvc.AccountServiceIface
	contactSvc contactsvc.ContactServiceIface

	msgCh chan *pb.SystemMsg
}

func NewSystemService(lhost host.Host, ids ipfsds.Batching,
	accountSvc accountsvc.AccountServiceIface, contactSvc contactsvc.ContactServiceIface) (*SystemSvc, error) {

	var err error

	systemsvc := &SystemSvc{
		host:       lhost,
		msgCh:      make(chan *pb.SystemMsg, 5),
		accountSvc: accountSvc,
		contactSvc: contactSvc,
	}

	systemsvc.systemProto, err = systemproto.NewSystemProto(lhost, ids, systemsvc.msgCh)
	if err != nil {
		return nil, fmt.Errorf("peerpeer.NewAccountSvc error: %s", err.Error())
	}

	// 后台处理系统消息
	go systemsvc.goHandleMessage()

	return systemsvc, nil
}

func (s *SystemSvc) goHandleMessage() {
	for msg := range s.msgCh {

		toPeerID := peer.ID(msg.ToPeer.PeerId)
		if toPeerID != s.host.ID() {
			// 不是自己的数据
			log.Warn("toPeerID is not equal")
			continue
		}

		switch msg.Type {
		case pb.SystemMsg_ContactApply: // 申请加好友
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
				if err = s.systemProto.UpdateMessageState(ctx, msg.Id, pb.SystemMsg_IsAgree); err != nil {
					log.Errorf("update message state error: %s", err.Error())
					continue
				}

				// 发送同意加好友消息
				fmt.Println("send message avatar: ", account.Avatar)
				if err = s.systemProto.SendMessage(ctx, &pb.SystemMsg{
					Id:    GenMsgID(account.PeerID),
					AckId: msg.Id,
					Type:  pb.SystemMsg_ContactApplyAck,
					FromPeer: &pb.Peer{
						PeerId: []byte(account.PeerID),
						Name:   account.Name,
						Avatar: account.Avatar,
					},
					ToPeer: &pb.Peer{
						PeerId: msg.FromPeer.PeerId,
						Name:   msg.FromPeer.Name,
						Avatar: msg.FromPeer.Avatar,
					},
					Content: "",
					State:   pb.SystemMsg_IsAgree,
					Ctime:   time.Now().Unix(),
					Utime:   time.Now().Unix(),
				}); err != nil {
					log.Errorf("send messge error: %s", err.Error())
				}

			}

		case pb.SystemMsg_ContactApplyAck: // 同意加好友
			ctx := context.Background()
			ackMsg, err := s.systemProto.GetMessage(ctx, msg.AckId)
			if err != nil {
				log.Errorf("systemProto.GetMessage error: %s-%s-", err.Error(), msg.AckId)
				continue
			}

			ackPeerID := peer.ID(ackMsg.ToPeer.PeerId)
			if peer.ID(msg.FromPeer.PeerId) != ackPeerID {
				log.Errorf("ackPeerID not equal")
				continue
			}

			switch msg.State {
			case pb.SystemMsg_IsAgree:
				// 保存好友信息
				fmt.Println("is agree avatar: ", msg.FromPeer.Avatar)
				if err = s.contactSvc.AddContact(ctx, ackPeerID, msg.FromPeer.Name, msg.FromPeer.Avatar); err != nil {
					log.Errorf("contactSvc.AddContact error: %s", err.Error())
					continue
				}

				// 更新系统消息状态，已加好友
				if err = s.systemProto.UpdateMessageState(ctx, msg.AckId, pb.SystemMsg_IsAgree); err != nil {
					log.Errorf("systemProto.UpdateMessage error: %s", err.Error())
				}
			case pb.SystemMsg_IsReject:
				// 更新系统消息状态，已拒绝
				if err = s.systemProto.UpdateMessageState(ctx, msg.AckId, pb.SystemMsg_IsReject); err != nil {
					log.Errorf("systemProto.UpdateMessage error: %s", err.Error())
				}
			default:
				// nothing
				log.Error("msg state error")
			}
		default:
			log.Error("msg type error")
		}
	}
}

func (s *SystemSvc) ApplyAddContact(ctx context.Context, peerID peer.ID, name string, avatar string, content string) error {
	account, err := s.accountSvc.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("accountSvc.GetAccount error: %w", err)
	}

	msg := pb.SystemMsg{
		Id:      GenMsgID(account.PeerID),
		AckId:   "",
		Type:    pb.SystemMsg_ContactApply,
		GroupId: "",
		FromPeer: &pb.Peer{
			PeerId: []byte(account.PeerID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		ToPeer: &pb.Peer{
			PeerId: []byte(peerID),
			Name:   name,
			Avatar: avatar,
		},
		Content: content,
		State:   pb.SystemMsg_IsSended,
		Ctime:   time.Now().Unix(),
		Utime:   time.Now().Unix(),
	}

	if err := s.systemProto.SaveMessage(ctx, &msg); err != nil {
		return fmt.Errorf("systemProto.SaveMessage error: %w", err)
	}

	if err = s.systemProto.SendMessage(ctx, &msg); err != nil {
		return fmt.Errorf("systemProto.SendMessage error: %w", err)
	}

	return nil
}

func (s *SystemSvc) GetSystemMessageList(ctx context.Context, offset int, limit int) ([]SystemMessage, error) {
	msgs, err := s.systemProto.GetMessageList(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("systemProto.GetMessageList error: %w", err)
	}

	var sysmsgs []SystemMessage
	for _, msg := range msgs {

		var msgType MsgType
		switch msg.Type {
		case pb.SystemMsg_ContactApply:
			msgType = TypeContactApply
		default:
			// nothing
		}

		var msgState MsgState
		switch msg.State {
		case pb.SystemMsg_IsSended:
			msgState = StateIsSended
		case pb.SystemMsg_IsAgree:
			msgState = StateIsAgree
		case pb.SystemMsg_IsReject:
			msgState = StateIsReject
		default:
			// nothing
		}

		sysmsgs = append(sysmsgs, SystemMessage{
			ID:      msg.Id,
			Type:    msgType,
			GroupID: msg.GroupId,
			Sender: Peer{
				PeerID: peer.ID(msg.FromPeer.PeerId),
				Name:   msg.FromPeer.Name,
				Avatar: msg.FromPeer.Avatar,
			},
			Receiver: Peer{
				PeerID: peer.ID(msg.ToPeer.PeerId),
				Name:   msg.ToPeer.Name,
				Avatar: msg.ToPeer.Avatar,
			},
			Content: msg.Content,
			State:   msgState,
			Ctime:   msg.Ctime,
			Utime:   msg.Utime,
		})
	}

	return sysmsgs, nil
}

func (s *SystemSvc) AgreeAddContact(ctx context.Context, ackID string) error {
	// 发送同意加好友消息

	ackMsg, err := s.systemProto.GetMessage(ctx, ackID)
	if err != nil {
		return fmt.Errorf("systemProto.GetMessage error: %w", err)
	}

	if err = s.systemProto.SendMessage(ctx, &pb.SystemMsg{
		Id:      GenMsgID(s.host.ID()),
		AckId:   ackID,
		Type:    pb.SystemMsg_ContactApplyAck,
		GroupId: "",
		FromPeer: &pb.Peer{
			PeerId: ackMsg.ToPeer.PeerId,
			Name:   ackMsg.ToPeer.Name,
			Avatar: ackMsg.ToPeer.Avatar,
		},
		ToPeer: &pb.Peer{
			PeerId: ackMsg.FromPeer.PeerId,
			Name:   ackMsg.FromPeer.Name,
			Avatar: ackMsg.FromPeer.Avatar,
		},
		Content: "",
		State:   pb.SystemMsg_IsAgree,
		Ctime:   time.Now().Unix(),
		Utime:   time.Now().Unix(),
	}); err != nil {
		return fmt.Errorf("systemProto.SendMessage error: %w", err)
	}

	if err = s.systemProto.UpdateMessageState(ctx, ackID, pb.SystemMsg_IsAgree); err != nil {
		return fmt.Errorf("systemProto.UpdateMessageState error: %w", err)
	}

	return nil
}

func (s *SystemSvc) RejectAddContact(ctx context.Context, ackID string) error {

	ackMsg, err := s.systemProto.GetMessage(ctx, ackID)
	if err != nil {
		return fmt.Errorf("systemProto.GetMessage error: %w", err)
	}

	if err = s.systemProto.SendMessage(ctx, &pb.SystemMsg{
		Id:      GenMsgID(s.host.ID()),
		AckId:   ackID,
		Type:    pb.SystemMsg_ContactApplyAck,
		GroupId: "",
		FromPeer: &pb.Peer{
			PeerId: ackMsg.ToPeer.PeerId,
			Name:   ackMsg.ToPeer.Name,
			Avatar: ackMsg.ToPeer.Avatar,
		},
		ToPeer: &pb.Peer{
			PeerId: ackMsg.FromPeer.PeerId,
			Name:   ackMsg.FromPeer.Name,
			Avatar: ackMsg.FromPeer.Avatar,
		},
		Content: "",
		State:   pb.SystemMsg_IsReject,
		Ctime:   time.Now().Unix(),
		Utime:   time.Now().Unix(),
	}); err != nil {
		return fmt.Errorf("systemProto.SendMessage error: %w", err)
	}

	if err = s.systemProto.UpdateMessageState(ctx, ackID, pb.SystemMsg_IsReject); err != nil {
		return fmt.Errorf("systemProto.UpdateMessageState error: %w", err)
	}

	return nil
}

func (s *SystemSvc) Close() {}

func GenMsgID(peerID peer.ID) string {
	return fmt.Sprintf("%d-%s", time.Now().UnixMilli(), peerID.String())
}
