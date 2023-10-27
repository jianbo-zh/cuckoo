package groupmsgproto

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	ds "github.com/jianbo-zh/dchat/datastore/ds/groupmsgds"
	"github.com/jianbo-zh/dchat/datastore/ds/sessionds"
	"github.com/jianbo-zh/dchat/internal/myerror"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/myprotocol"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/grouppb"
	"github.com/jianbo-zh/dchat/protobuf/pb/sessionpb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/libp2p/go-msgio/pbio"
	"google.golang.org/protobuf/proto"
)

var log = logging.Logger("group-message")

var StreamTimeout = 1 * time.Minute

const (
	MSG_ID  = myprotocol.GroupMessageID_v100
	SYNC_ID = myprotocol.GroupMessageSyncID_v100

	ServiceName = "group.message"
)

type MessageProto struct {
	host myhost.Host

	data        ds.MessageIface
	sessionData sessionds.SessionIface

	groupConns map[string]map[peer.ID]struct{}

	emitters struct {
		evtGetResourceData         event.Emitter
		evtSaveResourceData        event.Emitter
		evtReceiveGroupMessage     event.Emitter
		evtPullDepositGroupMessage event.Emitter
	}
}

func NewMessageProto(h myhost.Host, ids ipfsds.Batching, eventBus event.Bus) (*MessageProto, error) {
	var err error
	msgsvc := &MessageProto{
		host:        h,
		data:        ds.MessageWrap(ids),
		sessionData: sessionds.SessionWrap(ids),
		groupConns:  make(map[string]map[peer.ID]struct{}),
	}

	h.SetStreamHandler(MSG_ID, msgsvc.messageHandler)
	h.SetStreamHandler(SYNC_ID, msgsvc.syncHandler)

	// 接收群消息
	if msgsvc.emitters.evtReceiveGroupMessage, err = eventBus.Emitter(&myevent.EvtReceiveGroupMessage{}); err != nil {
		return nil, fmt.Errorf("set receive group msg emitter error: %v", err)
	}

	if msgsvc.emitters.evtGetResourceData, err = eventBus.Emitter(&myevent.EvtGetResourceData{}); err != nil {
		return nil, fmt.Errorf("set get resource data emitter error: %v", err)
	}

	if msgsvc.emitters.evtSaveResourceData, err = eventBus.Emitter(&myevent.EvtSaveResourceData{}); err != nil {
		return nil, fmt.Errorf("set save resource data emitter error: %v", err)
	}

	if msgsvc.emitters.evtPullDepositGroupMessage, err = eventBus.Emitter(&myevent.EvtPullDepositGroupMessage{}); err != nil {
		return nil, fmt.Errorf("set save resource data emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(myevent.EvtSyncGroupMessage), new(myevent.EvtGroupConnectChange)}, eventbus.Name("syncmsg"))
	if err != nil {
		return nil, fmt.Errorf("subscription failed. group admin server error: %v", err)

	} else {
		go msgsvc.subscribeHandler(context.Background(), sub)
	}

	return msgsvc, nil
}

func (m *MessageProto) messageHandler(stream network.Stream) {

	remotePeerID := stream.Conn().RemotePeer()

	if err := stream.Scope().SetService(ServiceName); err != nil {
		log.Errorf("failed to attaching stream to identify service: %v", err)
		stream.Reset()
		return
	}

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

	// 检查本地是否存在
	ctx := context.Background()
	coreMsg := switchMsg.CoreMessage
	exists, err := m.data.HasMessage(ctx, coreMsg.GroupId, coreMsg.Id)
	if err != nil {
		log.Errorf("data.HasMessage error: %v", err)
		stream.Reset()
		return

	} else if exists {
		return
	}

	// 更新本地lamptime
	err = m.data.MergeLamportTime(ctx, coreMsg.GroupId, coreMsg.Lamportime)
	if err != nil {
		log.Errorf("update lamport time error: %v", err)
	}

	if coreMsg.AttachmentId != "" {
		if len(switchMsg.Attachment) <= 0 {
			log.Errorf("msg attachment empty")
			stream.Reset()
			return
		}

		// 触发接收消息
		resultCh := make(chan error, 1)
		sessionID := mytype.GroupSessionID(coreMsg.GroupId)
		if err = m.emitters.evtSaveResourceData.Emit(myevent.EvtSaveResourceData{
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

	// 保存消息
	err = m.saveCoreMessage(ctx, coreMsg.GroupId, coreMsg, false)
	if err != nil {
		log.Errorf("save group message error: %v", err)
	}

	// Ack确认
	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&pb.GroupMessageAck{}); err != nil {
		log.Errorf("pbio write ack msg error: %w", err)
		stream.Reset()
		return
	}

	// 触发接收消息
	if err = m.emitters.evtReceiveGroupMessage.Emit(myevent.EvtReceiveGroupMessage{
		MsgID:      coreMsg.Id,
		GroupID:    coreMsg.GroupId,
		FromPeerID: peer.ID(coreMsg.Member.Id),
		MsgType:    coreMsg.MsgType,
		MimeType:   coreMsg.MimeType,
		Payload:    coreMsg.Payload,
		Timestamp:  coreMsg.Timestamp,
	}); err != nil {
		log.Errorln("emit receive group msg error: %w", err)
	}

	// 转发消息
	err = m.broadcastMessage(ctx, coreMsg.GroupId, &switchMsg, remotePeerID)
	if err != nil {
		log.Errorf("emit forward group msg error: %v", err)
	}
}

func (m *MessageProto) syncHandler(stream network.Stream) {
	defer stream.Close()

	// 获取同步GroupID
	var syncmsg pb.GroupSyncMessage
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	if err := rd.ReadMsg(&syncmsg); err != nil {
		log.Errorf("pbio read sync init msg error: %v", err)
		return
	}

	groupID := syncmsg.GroupId

	// 发送同步摘要
	summary, err := m.getMessageSummary(groupID)
	if err != nil {
		log.Errorf("get msg summary error: %v", err)
		return
	}

	bs, err := proto.Marshal(summary)
	if err != nil {
		log.Errorf("proto marshal summary error: %v", err)
		return
	}

	wt := pbio.NewDelimitedWriter(stream)
	if err = wt.WriteMsg(&pb.GroupSyncMessage{
		Type:    pb.GroupSyncMessage_SUMMARY,
		Payload: bs,
	}); err != nil {
		log.Errorf("pbio write sync summary error: %v", err)
		return
	}

	// 后台同步处理
	err = m.loopSync(groupID, stream, rd, wt)
	if err != nil {
		log.Errorf("loop sync error: %v", err)
	}
}

func (m *MessageProto) subscribeHandler(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}

			switch evt := e.(type) {
			case myevent.EvtGroupConnectChange:
				if evt.IsConnected {
					if _, exists := m.groupConns[evt.GroupID]; !exists {
						m.groupConns[evt.GroupID] = make(map[peer.ID]struct{})
					}

					if _, exists := m.groupConns[evt.GroupID][evt.PeerID]; !exists {
						m.groupConns[evt.GroupID][evt.PeerID] = struct{}{}

						if m.host.ID() > evt.PeerID {
							// 大方主动发起同步
							go m.goSync(evt.GroupID, evt.PeerID)
						}
					}
					log.Debugln("event connect peer: ", evt.PeerID.String())

				} else {
					log.Debugln("event disconnect peer: ", evt.PeerID.String())
					delete(m.groupConns[evt.GroupID], evt.PeerID)
				}
			case myevent.EvtSyncGroupMessage:
				if err := m.emitters.evtPullDepositGroupMessage.Emit(myevent.EvtPullDepositGroupMessage{
					GroupID:        evt.GroupID,
					DepositAddress: evt.DepositAddress,
					MessageHandler: m.SaveDepositMessage,
				}); err != nil {
					log.Errorf("emit EvtPullDepositGroupMessage error: %w", err)
				} else {
					log.Debugln("emit EvtPullDepositGroupMessage ", evt.GroupID)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func (m *MessageProto) SaveDepositMessage(ctx context.Context, groupID string, msgID string, msgData []byte) error {

	// 检查本地是否存在
	exists, err := m.data.HasMessage(ctx, groupID, msgID)
	if err != nil {
		return fmt.Errorf("data.HasMessage error: %w", err)

	} else if exists {
		return nil
	}

	var switchMsg pb.MessageEnvelope
	if err := proto.Unmarshal(msgData, &switchMsg); err != nil {
		return fmt.Errorf("proto.Unmarshal error: %w", err)
	}

	coreMsg := switchMsg.CoreMessage

	// 更新本地lamptime
	err = m.data.MergeLamportTime(ctx, coreMsg.GroupId, coreMsg.Lamportime)
	if err != nil {
		return fmt.Errorf("data.MergeLamportTime error: %w", err)
	}

	// 保存消息的资源
	if coreMsg.AttachmentId != "" {
		if len(switchMsg.Attachment) <= 0 {
			return fmt.Errorf("attachment data is empty")
		}

		resultCh := make(chan error, 1)
		sessionID := mytype.GroupSessionID(coreMsg.GroupId)
		if err = m.emitters.evtSaveResourceData.Emit(myevent.EvtSaveResourceData{
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

	// 保存消息
	err = m.saveCoreMessage(ctx, coreMsg.GroupId, coreMsg, true)
	if err != nil {
		log.Errorf("save group message error: %v", err)
		return fmt.Errorf("save core msg error: %w", err)
	}

	return nil
}

func (m *MessageProto) GetMessage(ctx context.Context, groupID string, msgID string) (*pb.GroupMessage, error) {
	msgs, err := m.data.GetMessage(ctx, groupID, msgID)
	if err != nil {
		return nil, fmt.Errorf("m.data.ListMessages error: %w", err)
	}

	return msgs, nil
}

func (m *MessageProto) GetMessageData(ctx context.Context, groupID string, msgID string) ([]byte, error) {
	return m.data.GetMessageData(ctx, groupID, msgID)
}

func (m *MessageProto) DeleteMessage(ctx context.Context, groupID string, msgID string) error {
	return m.data.DeleteMessage(ctx, groupID, msgID)
}

func (m *MessageProto) GetMessageList(ctx context.Context, groupID string, offset int, limit int) ([]*pb.GroupMessage, error) {
	msgs, err := m.data.GetMessages(ctx, groupID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("m.data.ListMessages error: %w", err)
	}

	return msgs, nil
}

func (m *MessageProto) CreateMessage(ctx context.Context, account *mytype.Account, groupID string, msgType string, mimeType string,
	payload []byte, attachmentID string) (*pb.GroupMessage, error) {

	lamptime, err := m.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("ds tick lamptime error: %w", err)
	}

	id := msgID(lamptime, account.ID)
	nowtime := time.Now().Unix()

	msg := pb.GroupMessage{
		Id:      id,
		GroupId: groupID,
		CoreMessage: &pb.CoreMessage{
			Id:      id,
			GroupId: groupID,
			Member: &pb.CoreMessage_Member{
				Id:     []byte(account.ID),
				Name:   account.Name,
				Avatar: account.Avatar,
			},
			MsgType:      msgType,
			MimeType:     mimeType,
			Payload:      payload,
			AttachmentId: attachmentID,
			Lamportime:   lamptime,
			Timestamp:    nowtime,
			Signature:    []byte{},
		},
		SendState:  pb.GroupMessage_Sending,
		CreateTime: nowtime,
		UpdateTime: nowtime,
	}

	// 保存消息
	isLatest, err := m.data.SaveMessage(ctx, groupID, &msg)
	if err != nil {
		return nil, fmt.Errorf("data save message error: %w", err)
	}

	// 更新session时间
	sessionID := mytype.GroupSessionID(groupID)
	if err = m.sessionData.UpdateSessionTime(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("data.UpdateSessionTime error: %w", err)
	}

	// 重置session未读消息
	if err = m.sessionData.ResetUnreads(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("data.IncrUnreads error: %w", err)
	}

	if isLatest {
		var username string
		if msg.CoreMessage.Member != nil {
			username = msg.CoreMessage.Member.Name
		}

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

		if err = m.sessionData.SetLastMessage(ctx, sessionID.String(), &sessionpb.SessionLastMessage{
			Username: username,
			Content:  content,
		}); err != nil {
			return nil, fmt.Errorf("data.SetLastMessage error: %w", err)
		}
	}

	return &msg, nil
}

func (m *MessageProto) SendGroupMessage(ctx context.Context, groupID string, msgID string) ([]byte, error) {

	coreMsg, err := m.data.GetCoreMessage(ctx, groupID, msgID)
	if err != nil {
		return nil, fmt.Errorf("data.GetMessage error: %w", err)
	}

	var attachment []byte
	if coreMsg.AttachmentId != "" {
		// 获取附件数据
		resultCh := make(chan *myevent.GetResourceDataResult, 1)
		if err := m.emitters.evtGetResourceData.Emit(myevent.EvtGetResourceData{
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

	switchMsg := pb.MessageEnvelope{
		Id:          coreMsg.Id,
		CoreMessage: coreMsg,
		Attachment:  attachment,
	}

	if err1 := m.broadcastMessage(ctx, groupID, &switchMsg); err1 != nil {
		if bs, err2 := proto.Marshal(&switchMsg); err2 != nil {
			return nil, fmt.Errorf("proto.Marshal error2: %w, error1: %v", err2, err1)

		} else {
			return bs, err1
		}
	}

	return nil, nil
}

func (m *MessageProto) UpdateMessageState(ctx context.Context, groupID string, msgID string, isDeposit bool, isSucc bool) (*pb.GroupMessage, error) {
	return m.data.UpdateMessageSendState(ctx, groupID, msgID, isDeposit, isSucc)
}

func (m *MessageProto) ClearGroupMessage(ctx context.Context, groupID string) error {
	return m.data.ClearMessage(ctx, groupID)
}

// 发送消息
func (m *MessageProto) broadcastMessage(ctx context.Context, groupID string, msg *pb.MessageEnvelope, excludePeerIDs ...peer.ID) error {

	peerIDs := m.getConnectPeers(groupID)
	if len(peerIDs) == 0 {
		return fmt.Errorf("no connect peers: %w", myerror.ErrSendGroupMessageFailed)
	}

	var forwardPeerIDs []peer.ID
	for _, peerID := range peerIDs {
		if len(excludePeerIDs) > 0 {
			isExcluded := false
			for _, excludePeerID := range excludePeerIDs {
				if peerID == excludePeerID {
					isExcluded = true
				}
			}
			if isExcluded {
				continue
			}
		}
		forwardPeerIDs = append(forwardPeerIDs, peerID)
	}

	if len(forwardPeerIDs) > 0 {
		isSended := false
		for _, peerID := range forwardPeerIDs {
			if err := m.sendPeerMessage(ctx, groupID, peerID, msg); err != nil {
				log.Errorf("sendPeerMessage error: %w", err)

			} else {
				isSended = true
			}

		}
		if !isSended {
			return myerror.ErrSendGroupMessageFailed
		}
	}

	return nil
}

// 发送消息（指定peerID）
func (m *MessageProto) sendPeerMessage(ctx context.Context, groupID string, peerID peer.ID, msg *pb.MessageEnvelope) error {

	stream, err := m.host.NewStream(network.WithUseTransient(ctx, ""), peerID, MSG_ID)
	if err != nil {
		return err
	}

	stream.SetWriteDeadline(time.Now().Add(StreamTimeout))

	wt := pbio.NewDelimitedWriter(stream)
	defer wt.Close()

	if err := wt.WriteMsg(msg); err != nil {
		return fmt.Errorf("pbio write group msg error: %w", err)
	}

	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	var ackMsg pb.GroupMessageAck
	if err := rd.ReadMsg(&ackMsg); err != nil {
		return fmt.Errorf("pbio read ack msg error: %w", err)
	}

	return nil
}

func (m *MessageProto) getConnectPeers(groupID string) []peer.ID {
	var peerIDs []peer.ID
	for peerID := range m.groupConns[groupID] {
		peerIDs = append(peerIDs, peerID)
	}
	return peerIDs
}

func (m *MessageProto) saveCoreMessage(ctx context.Context, groupID string, coreMsg *pb.CoreMessage, isDeposit bool) error {

	exists, err := m.data.HasMessage(ctx, groupID, coreMsg.Id)
	if err != nil {
		return fmt.Errorf("data.HasMessage error: %w", err)

	} else if exists {
		return nil
	}

	isLatest, err := m.data.SaveMessage(ctx, groupID, &pb.GroupMessage{
		Id:          coreMsg.Id,
		GroupId:     coreMsg.GroupId,
		CoreMessage: coreMsg,
		IsDeposit:   isDeposit,
		SendState:   pb.GroupMessage_SendSucc,
		CreateTime:  time.Now().Unix(),
		UpdateTime:  time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("data save message error: %w", err)
	}

	sessionID := mytype.GroupSessionID(groupID)
	// 更新session时间
	if err = m.sessionData.UpdateSessionTime(ctx, sessionID.String()); err != nil {
		return fmt.Errorf("data.UpdateSessionTime error: %w", err)
	}

	// 更新session未读消息
	if coreMsg.Member != nil && peer.ID(coreMsg.Member.Id) == m.host.ID() {
		if err = m.sessionData.ResetUnreads(ctx, sessionID.String()); err != nil {
			return fmt.Errorf("data.ResetUnreads error: %w", err)
		}
	} else {
		if err = m.sessionData.IncrUnreads(ctx, sessionID.String()); err != nil {
			return fmt.Errorf("data.IncrUnreads error: %w", err)
		}
	}

	// 更新session最新消息
	if isLatest {
		var username string
		if coreMsg.Member != nil {
			username = coreMsg.Member.Name
		}

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

		if err = m.sessionData.SetLastMessage(ctx, sessionID.String(), &sessionpb.SessionLastMessage{
			Username: username,
			Content:  content,
		}); err != nil {
			return fmt.Errorf("data.SetLastMessage error: %w", err)
		}
	}

	return nil
}
