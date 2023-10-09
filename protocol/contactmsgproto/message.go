package contactmsgproto

import (
	"context"
	"fmt"
	"time"

	ds "github.com/jianbo-zh/dchat/datastore/ds/contactmsgds"
	"github.com/jianbo-zh/dchat/datastore/ds/sessionds"
	"github.com/jianbo-zh/dchat/internal/myerror"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/myprotocol"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/contactpb"
	"github.com/jianbo-zh/dchat/protobuf/pb/sessionpb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
	"google.golang.org/protobuf/proto"

	ipfsds "github.com/ipfs/go-datastore"
)

// 消息服务

var log = logging.Logger("contact-message")

var StreamTimeout = 1 * time.Minute

const (
	MSG_ID      = myprotocol.ContactMessageID_v100
	MSG_SYNC_ID = myprotocol.ContactMessageSyncID_v100

	ServiceName = "peer.message"
	maxMsgSize  = 4 * 1024 // 4K
)

type PeerMessageProto struct {
	host        myhost.Host
	data        ds.PeerMessageIface
	sessionData sessionds.SessionIface

	emitters struct {
		evtPushOfflineMessage event.Emitter
		evtPullOfflineMessage event.Emitter
		evtReceiveMessage     event.Emitter
	}
}

func NewMessageSvc(lhost myhost.Host, ids ipfsds.Batching, eventBus event.Bus) (*PeerMessageProto, error) {
	var err error
	msgsvc := PeerMessageProto{
		host:        lhost,
		data:        ds.Wrap(ids),
		sessionData: sessionds.SessionWrap(ids),
	}

	lhost.SetStreamHandler(MSG_ID, msgsvc.msgIDHandler)
	lhost.SetStreamHandler(MSG_SYNC_ID, msgsvc.syncIDHandler)

	// 触发器：发送离线消息
	if msgsvc.emitters.evtPushOfflineMessage, err = eventBus.Emitter(&myevent.EvtPushDepositContactMessage{}); err != nil {
		return nil, fmt.Errorf("set pull deposit msg emitter error: %v", err)
	}

	// 触发器：获取离线消息
	if msgsvc.emitters.evtPullOfflineMessage, err = eventBus.Emitter(&myevent.EvtPullDepositContactMessage{}); err != nil {
		return nil, fmt.Errorf("set pull deposit msg emitter error: %v", err)
	}

	// 触发器：接收到消息
	if msgsvc.emitters.evtReceiveMessage, err = eventBus.Emitter(&myevent.EvtReceiveContactMessage{}); err != nil {
		return nil, fmt.Errorf("set receive msg emitter error: %v", err)
	}

	// 订阅：同步Peer的消息，拉取离线消息事件
	sub, err := eventBus.Subscribe([]any{new(myevent.EvtSyncContactMessage), new(event.EvtLocalAddressesUpdated)})
	if err != nil {
		return nil, fmt.Errorf("subscribe boot complete event error: %v", err)

	} else {
		go msgsvc.subscribeHandler(context.Background(), sub)
	}

	return &msgsvc, nil
}

func (p *PeerMessageProto) subscribeHandler(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				log.Warnf("subscribe out not ok")
				return
			}

			switch evt := e.(type) {
			case myevent.EvtSyncContactMessage:
				// todo: 同步功能暂时不开，所以功能可以暂时不测
				// go p.goSyncMessage(evt.ContactID)

			case event.EvtLocalAddressesUpdated:
				if !evt.Diffs {
					log.Warnf("event local address no ok or no diff")
					return
				}

				for _, curr := range evt.Current {
					if isPublicAddr(curr.Address) {
						p.emitters.evtPullOfflineMessage.Emit(myevent.EvtPullDepositContactMessage{
							DepositAddress: peer.ID(""),
							MessageHandler: p.SaveCoreMessage,
						})
						break
					}
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func (p *PeerMessageProto) msgIDHandler(stream network.Stream) {

	// 开始处理流
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rd.Close()

	stream.SetDeadline(time.Now().Add(StreamTimeout))
	var coreMsg pb.ContactMessage_CoreMessage
	if err := rd.ReadMsg(&coreMsg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		stream.Reset()
		return
	}
	stream.SetReadDeadline(time.Time{})

	remotePeerID := stream.Conn().RemotePeer()

	if err := p.saveCoreMessage(context.Background(), remotePeerID, &coreMsg); err != nil {
		log.Errorf("store message error %v", err)
		stream.Reset()
		return
	}

	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&pb.ContactMessageAck{Id: coreMsg.Id}); err != nil {
		log.Errorf("ack msg error: %w", err)
		stream.Reset()
		return
	}

	if err := p.data.MergeLamportTime(context.Background(), remotePeerID, coreMsg.Lamportime); err != nil {
		log.Errorf("update lamport time error: %v", err)
		stream.Reset()
		return
	}

	p.emitters.evtReceiveMessage.Emit(myevent.EvtReceiveContactMessage{
		MsgID:      coreMsg.Id,
		FromPeerID: peer.ID(coreMsg.FromPeerId),
		MsgType:    coreMsg.MsgType,
		MimeType:   coreMsg.MimeType,
		Payload:    coreMsg.Payload,
		Timestamp:  coreMsg.Timestamp,
	})
}

func (p *PeerMessageProto) SaveCoreMessage(ctx context.Context, peerID peer.ID, msgID string, coreMsgData []byte) error {
	var coreMsg pb.ContactMessage_CoreMessage
	if err := proto.Unmarshal(coreMsgData, &coreMsg); err != nil {
		return err

	} else if coreMsg.Id != msgID {
		return fmt.Errorf("msg id not equal")
	}

	return p.saveCoreMessage(ctx, peerID, &coreMsg)
}

// CreateMessage 创建新消息
func (p *PeerMessageProto) CreateMessage(ctx context.Context, contactID peer.ID, msgType string, mimeType string, payload []byte) (*pb.ContactMessage, error) {

	hostID := p.host.ID()

	lamptime, err := p.data.TickLamportTime(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("data tick lamptime error: %w", err)
	}

	id := msgID(lamptime, hostID)
	nowtime := time.Now().Unix()

	msg := pb.ContactMessage{
		Id: id,
		CoreMessage: &pb.ContactMessage_CoreMessage{
			Id:         id,
			FromPeerId: []byte(hostID),
			ToPeerId:   []byte(contactID),
			MsgType:    msgType,
			MimeType:   mimeType,
			Payload:    payload,
			Lamportime: lamptime,
			Timestamp:  nowtime,
			Signature:  []byte{},
		},
		SendState:  pb.ContactMessage_Sending,
		CreateTime: nowtime,
		UpdateTime: nowtime,
	}

	isLatest, err := p.data.SaveMessage(ctx, contactID, &msg)
	if err != nil {
		return nil, fmt.Errorf("data save message error: %w", err)
	}

	// 更新session时间
	sessionID := mytype.ContactSessionID(contactID)
	if err = p.sessionData.UpdateSessionTime(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("data.UpdateSessionTime error: %w", err)
	}

	// 重置session未读消息
	if err = p.sessionData.ResetUnreads(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("data.IncrUnreads error: %w", err)
	}

	if isLatest {
		// 更新session最新消息
		var content string
		switch msg.CoreMessage.MsgType {
		case mytype.TextMsgType:
			content = string(msg.CoreMessage.Payload)
		case mytype.ImageMsgType:
			content = string("[image]")
		case mytype.VoiceMsgType:
			content = string("[voice]")
		case mytype.AudioMsgType:
			content = string("[audio]")
		case mytype.VideoMsgType:
			content = string("[video]")
		case mytype.FileMsgType:
			content = string("[file]")
		default:
			content = string("[unknown]")
		}

		if err = p.sessionData.SetLastMessage(ctx, sessionID.String(), &sessionpb.SessionLastMessage{
			Username: "",
			Content:  content,
		}); err != nil {
			return nil, fmt.Errorf("data.SetLastMessage error: %w", err)
		}
	}

	return &msg, nil
}

// SendMessage 发送消息，实际发送的是核心消息
func (p *PeerMessageProto) SendMessage(ctx context.Context, peerID peer.ID, msgID string) error {

	hostID := p.host.ID()

	coreMsg, err := p.data.GetCoreMessage(ctx, peerID, msgID)
	if err != nil {
		return fmt.Errorf("data.GetMessage error: %w", err)
	}

	if peer.ID(coreMsg.FromPeerId) != hostID {
		return fmt.Errorf("msg from peer id not equal host id")
	}

	stream, err := p.host.NewStream(network.WithDialPeerTimeout(ctx, time.Second), peerID, MSG_ID)
	if err != nil {
		return myerror.WrapStreamError("host new stream error", err)
	}

	pw := pbio.NewDelimitedWriter(stream)
	defer pw.Close()

	if err = pw.WriteMsg(coreMsg); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("pbio write msg error", err)
	}

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	var msgAck pb.ContactMessageAck
	if err = rd.ReadMsg(&msgAck); err != nil {
		return myerror.WrapStreamError("pbio read ack msg error", err)

	} else if msgAck.Id != coreMsg.Id {
		return myerror.WrapStreamError("msg ack id error", nil)
	}

	return nil
}

func (p *PeerMessageProto) GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*pb.ContactMessage, error) {
	return p.data.GetMessage(ctx, peerID, msgID)
}

func (p *PeerMessageProto) UpdateMessageState(ctx context.Context, peerID peer.ID, msgID string, isSucc bool) (*pb.ContactMessage, error) {
	return p.data.UpdateMessageSendState(ctx, peerID, msgID, isSucc)
}

func (p *PeerMessageProto) DeleteMessage(ctx context.Context, peerID peer.ID, msgID string) error {
	return p.data.DeleteMessage(ctx, peerID, msgID)
}

func (p *PeerMessageProto) GetMessageData(ctx context.Context, peerID peer.ID, msgID string) ([]byte, error) {
	return p.data.GetMessageData(ctx, peerID, msgID)
}

func (p *PeerMessageProto) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]*pb.ContactMessage, error) {
	return p.data.GetMessages(ctx, peerID, offset, limit)
}

func (p *PeerMessageProto) ClearMessage(ctx context.Context, peerID peer.ID) error {
	return p.data.ClearMessage(ctx, peerID)
}

func (p *PeerMessageProto) saveCoreMessage(ctx context.Context, contactID peer.ID, coreMsg *pb.ContactMessage_CoreMessage) error {

	exists, err := p.data.HasMessage(ctx, contactID, coreMsg.Id)
	if err != nil {
		return fmt.Errorf("data.HasMessage error: %w", err)

	} else if exists {
		return nil
	}

	isLatest, err := p.data.SaveMessage(ctx, contactID, &pb.ContactMessage{
		Id:          coreMsg.Id,
		CoreMessage: coreMsg,
		SendState:   pb.ContactMessage_SendSucc,
		CreateTime:  time.Now().Unix(),
		UpdateTime:  time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("data save message error: %w", err)
	}

	sessionID := mytype.ContactSessionID(contactID)
	// 更新session时间
	if err = p.sessionData.UpdateSessionTime(ctx, sessionID.String()); err != nil {
		return fmt.Errorf("data.UpdateSessionTime error: %w", err)
	}

	// 更新session未读消息
	if peer.ID(coreMsg.FromPeerId) == p.host.ID() {
		if err = p.sessionData.ResetUnreads(ctx, sessionID.String()); err != nil {
			return fmt.Errorf("data.IncrUnreads error: %w", err)
		}
	} else {
		if err = p.sessionData.IncrUnreads(ctx, sessionID.String()); err != nil {
			return fmt.Errorf("data.IncrUnreads error: %w", err)
		}
	}

	if isLatest {
		// 更新session最新消息
		var content string
		switch coreMsg.MsgType {
		case mytype.TextMsgType:
			content = string(coreMsg.Payload)
		case mytype.ImageMsgType:
			content = string("[image]")
		case mytype.VoiceMsgType:
			content = string("[voice]")
		case mytype.AudioMsgType:
			content = string("[audio]")
		case mytype.VideoMsgType:
			content = string("[video]")
		case mytype.FileMsgType:
			content = string("[file]")
		default:
			content = string("[unknown]")
		}

		if err = p.sessionData.SetLastMessage(ctx, sessionID.String(), &sessionpb.SessionLastMessage{
			Username: "",
			Content:  content,
		}); err != nil {
			return fmt.Errorf("data.SetLastMessage error: %w", err)
		}
	}

	return nil
}
