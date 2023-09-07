package systemsvc

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
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

func NewSystemService(conf config.PeerMessageConfig, lhost host.Host, ids ipfsds.Batching,
	accountSvc accountsvc.AccountServiceIface, contactSvc contactsvc.ContactServiceIface) (*SystemSvc, error) {

	var err error

	systemsvc := &SystemSvc{
		msgCh:      make(chan *pb.SystemMsg, 5),
		accountSvc: accountSvc,
		contactSvc: contactSvc,
	}

	systemsvc.systemProto, err = systemproto.NewSystemProto(lhost, ids, systemsvc.msgCh)
	if err != nil {
		return nil, fmt.Errorf("peerpeer.NewAccountSvc error: %s", err.Error())
	}

	go systemsvc.goHandleMessage()

	return systemsvc, nil
}

func (s *SystemSvc) goHandleMessage() {
	for msg := range s.msgCh {
		toPeerID := peer.ID(msg.ToPeerId)
		if toPeerID != s.host.ID() {
			// 不是自己的数据
			log.Warn("toPeerID is not equal")
			continue
		}

		switch msg.Type {
		case pb.SystemMsg_ContactApply: // 申请加好友
			ctx := context.Background()
			// 系统消息入库
			if err := s.systemProto.SaveMssage(ctx, msg); err != nil {
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
				// 如果是自动加好友，则更新系统消息状态为已同意
				if err = s.systemProto.UpdateMessageState(ctx, msg.Id, pb.SystemMsg_IsAgree); err != nil {
					log.Errorf("update message state error: %s", err.Error())
					continue
				}

				// 发送同意加好友消息
				if err = s.systemProto.SendMessage(ctx, peer.ID(msg.FromPeer.PeerId), &pb.SystemMsg{
					Id:   GenMsgID(account.PeerID),
					Type: pb.SystemMsg_ContactApplyAck,
					FromPeer: &pb.Peer{
						PeerId: []byte(account.PeerID),
						Name:   account.Name,
						Avatar: account.Avatar,
					},
					ToPeerId: msg.FromPeer.PeerId,
					Content:  "",
					State:    pb.SystemMsg_IsAgree,
					Ctime:    time.Now().Unix(),
					Utime:    time.Now().Unix(),
				}); err != nil {
					log.Errorf("send messge error: %s", err.Error())
				}

				// 添加为好友
				if err = s.contactSvc.AddContact(ctx, peer.ID(msg.FromPeer.PeerId), msg.FromPeer.Name, msg.FromPeer.Avatar); err != nil {
					log.Errorf("contactSvc.AddContact error: %s", err.Error())
				}
			}

		case pb.SystemMsg_ContactApplyAck: // 同意加好友
			ctx := context.Background()
			ackMsg, err := s.systemProto.GetMssage(ctx, msg.AckId)
			if err != nil {
				log.Errorf("systemProto.GetMessage error: %s", err.Error())
				continue
			}

			ackPeerID := peer.ID(ackMsg.ToPeerId)
			if peer.ID(msg.FromPeer.PeerId) != ackPeerID {
				log.Errorf("ackPeerID not equal")
				continue
			}

			switch msg.State {
			case pb.SystemMsg_IsAgree:
				// 保存好友信息
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
			}

		}
	}
}

func (s *SystemSvc) ApplyAddContact(ctx context.Context, peerID peer.ID, content string) error {

	return nil
}

func (s *SystemSvc) AgreeAddContact(ctx context.Context, peerID peer.ID, ackID string) error {
	return nil
}

func GenMsgID(peerID peer.ID) string {
	return fmt.Sprintf("%d-%s", time.Now().Unix(), peerID.String())
}
