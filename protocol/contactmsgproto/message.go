package contactmsgproto

import (
	"context"
	"errors"
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
)

type PeerMessageProto struct {
	host        myhost.Host
	data        ds.PeerMessageIface
	sessionData sessionds.SessionIface

	emitters struct {
		evtPushOfflineMessage event.Emitter
		evtPullOfflineMessage event.Emitter
		evtReceiveMessage     event.Emitter
		evtGetResourceData    event.Emitter
		evtSaveResourceData   event.Emitter
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

	// 触发器：获取离线消息
	if msgsvc.emitters.evtPullOfflineMessage, err = eventBus.Emitter(&myevent.EvtPullDepositContactMessage{}); err != nil {
		return nil, fmt.Errorf("set pull deposit msg emitter error: %v", err)
	}

	// 触发器：接收到消息
	if msgsvc.emitters.evtReceiveMessage, err = eventBus.Emitter(&myevent.EvtReceiveContactMessage{}); err != nil {
		return nil, fmt.Errorf("set receive msg emitter error: %v", err)
	}

	if msgsvc.emitters.evtGetResourceData, err = eventBus.Emitter(&myevent.EvtGetResourceData{}); err != nil {
		return nil, fmt.Errorf("set receive msg emitter error: %v", err)
	}

	if msgsvc.emitters.evtSaveResourceData, err = eventBus.Emitter(&myevent.EvtSaveResourceData{}); err != nil {
		return nil, fmt.Errorf("set receive msg emitter error: %v", err)
	}

	// 订阅：同步Peer的消息，拉取离线消息事件
	sub, err := eventBus.Subscribe([]any{new(myevent.EvtSyncAccountMessage), new(myevent.EvtSyncContactMessage)})
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
			case myevent.EvtSyncAccountMessage:
				if err := p.emitters.evtPullOfflineMessage.Emit(myevent.EvtPullDepositContactMessage{
					DepositAddress: evt.DepositAddress,
					MessageHandler: p.SaveDepositMessage,
				}); err != nil {
					log.Errorf("emit EvtPullDepositContactMessage error: %w", err)
				} else {
					log.Debugln("emit pull offline message")
				}

			case myevent.EvtSyncContactMessage:
				// todo: 同步功能暂时不开，所以功能可以暂时不测
				// go p.goSyncMessage(evt.ContactID)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (p *PeerMessageProto) msgIDHandler(stream network.Stream) {
	remotePeerID := stream.Conn().RemotePeer()

	// 开始处理流
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeMessage)
	defer rd.Close()

	stream.SetDeadline(time.Now().Add(StreamTimeout))

	var switchMsg pb.MessageEnvelope
	if err := rd.ReadMsg(&switchMsg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		stream.Reset()
		return
	}
	stream.SetReadDeadline(time.Time{})

	ctx := context.Background()
	exists, err := p.data.HasMessage(ctx, peer.ID(switchMsg.CoreMessage.FromPeerId), switchMsg.CoreMessage.Id)
	if err != nil {
		log.Errorf("data.HasMessage Error: %w", err)
		stream.Reset()
		return

	} else if exists {
		return
	}

	coreMsg := switchMsg.CoreMessage

	if coreMsg.AttachmentId != "" {
		if len(switchMsg.Attachment) <= 0 {
			log.Errorf("msg attachment empty")
			stream.Reset()
			return
		}

		// 触发接收消息
		resultCh := make(chan error, 1)
		sessionID := mytype.ContactSessionID(peer.ID(coreMsg.FromPeerId))
		if err = p.emitters.evtSaveResourceData.Emit(myevent.EvtSaveResourceData{
			SessionID:  sessionID.String(),
			ResourceID: coreMsg.AttachmentId,
			Data:       switchMsg.Attachment,
			Result:     resultCh,
		}); err != nil {
			log.Errorf("emit receive group msg error: %w", err)
			stream.Reset()
			return
		}

		if err := <-resultCh; err != nil {
			log.Errorf("save resource data error: %w", err)
			stream.Reset()
			return
		}
	}

	if err := p.saveCoreMessage(context.Background(), peer.ID(coreMsg.FromPeerId), coreMsg, false); err != nil {
		log.Errorf("store message error %v", err)
		stream.Reset()
		return
	}

	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&pb.ContactMessageAck{Id: switchMsg.Id}); err != nil {
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

func (p *PeerMessageProto) SaveDepositMessage(ctx context.Context, fromPeerID peer.ID, msgID string, msgData []byte) error {

	var switchMsg pb.MessageEnvelope
	if err := proto.Unmarshal(msgData, &switchMsg); err != nil {
		return fmt.Errorf("proto.Unmarshal error: %w", err)
	}

	coreMsg := switchMsg.CoreMessage

	if err := p.data.MergeLamportTime(ctx, fromPeerID, coreMsg.Lamportime); err != nil {
		return fmt.Errorf("data.MergeLamportTime error: %w", err)
	}

	if coreMsg.AttachmentId != "" {
		if len(switchMsg.Attachment) <= 0 {
			return fmt.Errorf("attachment data is empty")
		}

		// 触发接收消息
		resultCh := make(chan error, 1)
		sessionID := mytype.ContactSessionID(peer.ID(coreMsg.FromPeerId))
		if err := p.emitters.evtSaveResourceData.Emit(myevent.EvtSaveResourceData{
			SessionID:  sessionID.String(),
			ResourceID: coreMsg.AttachmentId,
			Data:       switchMsg.Attachment,
			Result:     resultCh,
		}); err != nil {
			return fmt.Errorf("emit EvtSaveResourceData error: %w", err)
		}

		if err := <-resultCh; err != nil {
			return fmt.Errorf("save resource data error: %w", err)
		}
	}

	if err := p.saveCoreMessage(ctx, peer.ID(coreMsg.FromPeerId), coreMsg, true); err != nil {
		return fmt.Errorf("saveCoreMessage error: %w", err)
	}

	return nil
}

// CreateMessage 创建新消息
func (p *PeerMessageProto) CreateMessage(ctx context.Context, contactID peer.ID, msgType string, mimeType string, payload []byte, attachmentID string) (*pb.ContactMessage, error) {

	hostID := p.host.ID()

	lamptime, err := p.data.TickLamportTime(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("data tick lamptime error: %w", err)
	}

	id := msgID(lamptime, hostID)
	nowtime := time.Now().Unix()

	msg := pb.ContactMessage{
		Id: id,
		CoreMessage: &pb.CoreMessage{
			Id:           id,
			FromPeerId:   []byte(hostID),
			ToPeerId:     []byte(contactID),
			MsgType:      msgType,
			MimeType:     mimeType,
			Payload:      payload,
			AttachmentId: attachmentID,
			Lamportime:   lamptime,
			Timestamp:    nowtime,
			Signature:    []byte{},
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

func (p *PeerMessageProto) GetMessageEnvelope(ctx context.Context, contactID peer.ID, msgID string) ([]byte, error) {
	msgEnvelop, err := p.getMessageEnvelope(ctx, contactID, msgID)
	if err != nil {
		return nil, fmt.Errorf("getMessageEnvelope error: %w", err)
	}

	bs, err := proto.Marshal(msgEnvelop)
	if err != nil {
		return nil, fmt.Errorf("proto.Marshal error: %w", err)
	}

	return bs, nil
}

func (p *PeerMessageProto) getMessageEnvelope(ctx context.Context, contactID peer.ID, msgID string) (*pb.MessageEnvelope, error) {
	hostID := p.host.ID()

	coreMsg, err := p.data.GetCoreMessage(ctx, contactID, msgID)
	if err != nil {
		return nil, fmt.Errorf("data.GetMessage error: %w", err)
	}

	if peer.ID(coreMsg.FromPeerId) != hostID {
		return nil, fmt.Errorf("msg from peer id not equal host id")
	}

	var attachment []byte
	if coreMsg.AttachmentId != "" {
		// 获取附件数据
		resultCh := make(chan *myevent.GetResourceDataResult, 1)
		if err := p.emitters.evtGetResourceData.Emit(myevent.EvtGetResourceData{
			ResourceID: coreMsg.AttachmentId,
			Result:     resultCh,
		}); err != nil {
			return nil, fmt.Errorf("emit evtGetResourceData error: %w", err)
		}

		result := <-resultCh
		if result == nil {
			return nil, fmt.Errorf("GetResourceDataResult is nil")

		} else if result.Error != nil {
			return nil, fmt.Errorf("GetResourceDataResult error: %w", result.Error)
		}

		attachment = result.Data
	}

	return &pb.MessageEnvelope{
		Id:          coreMsg.Id,
		CoreMessage: coreMsg,
		Attachment:  attachment,
	}, nil
}

// SendMessage 发送消息
func (p *PeerMessageProto) SendMessage(ctx context.Context, contactID peer.ID, msgID string) ([]byte, error) {

	switchMsg, err := p.getMessageEnvelope(ctx, contactID, msgID)
	if err != nil {
		return nil, fmt.Errorf("get message envelope error: %w", err)
	}

	if err := p.sendMessage(ctx, contactID, switchMsg); err != nil {
		if errors.As(err, &myerror.StreamErr{}) {
			// 流错误，表示可能对方不在线
			bs, err2 := proto.Marshal(switchMsg)
			if err2 != nil {
				return nil, fmt.Errorf("proto.Marshal error: %w", err2)
			}

			return bs, fmt.Errorf("send message stream error: %w", err)
		}
		return nil, err
	}

	return nil, nil
}

func (p *PeerMessageProto) sendMessage(ctx context.Context, peerID peer.ID, msg *pb.MessageEnvelope) error {

	stream, err := p.host.NewStream(network.WithDialPeerTimeout(ctx, time.Second), peerID, MSG_ID)
	if err != nil {
		return myerror.WrapStreamError("host new stream error", err)
	}

	pw := pbio.NewDelimitedWriter(stream)
	defer pw.Close()

	if err = pw.WriteMsg(msg); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("pbio write msg error", err)
	}

	var msgAck pb.ContactMessageAck
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	if err = rd.ReadMsg(&msgAck); err != nil {
		return myerror.WrapStreamError("pbio read ack msg error", err)

	} else if msgAck.Id != msg.Id {
		return myerror.WrapStreamError("msg ack id error", nil)
	}

	return nil
}

func (p *PeerMessageProto) GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*pb.ContactMessage, error) {
	return p.data.GetMessage(ctx, peerID, msgID)
}

func (p *PeerMessageProto) UpdateMessageState(ctx context.Context, peerID peer.ID, msgID string, isDeposit bool, isSucc bool) (*pb.ContactMessage, error) {
	return p.data.UpdateMessageSendState(ctx, peerID, msgID, isDeposit, isSucc)
}

func (p *PeerMessageProto) DeleteMessage(ctx context.Context, peerID peer.ID, msgID string) error {
	return p.data.DeleteMessage(ctx, peerID, msgID)
}

func (p *PeerMessageProto) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]*pb.ContactMessage, error) {
	return p.data.GetMessages(ctx, peerID, offset, limit)
}

func (p *PeerMessageProto) ClearMessage(ctx context.Context, contactID peer.ID) error {
	return p.data.ClearMessage(ctx, contactID)
}

func (p *PeerMessageProto) saveCoreMessage(ctx context.Context, contactID peer.ID, coreMsg *pb.CoreMessage, isDeposit bool) error {

	exists, err := p.data.HasMessage(ctx, contactID, coreMsg.Id)
	if err != nil {
		return fmt.Errorf("data.HasMessage error: %w", err)

	} else if exists {
		return nil
	}

	isLatest, err := p.data.SaveMessage(ctx, contactID, &pb.ContactMessage{
		Id:          coreMsg.Id,
		CoreMessage: coreMsg,
		IsDeposit:   isDeposit,
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
