package groupmsgproto

import (
	"context"
	"errors"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	ds "github.com/jianbo-zh/dchat/datastore/ds/groupmsgds"
	"github.com/jianbo-zh/dchat/datastore/ds/sessionds"
	"github.com/jianbo-zh/dchat/internal/myerror"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/internal/protocol"
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
	ID      = protocol.GroupMessageID_v100
	SYNC_ID = protocol.GroupMessageSyncID_v100

	ServiceName = "group.message"
	maxMsgSize  = 4 * 1024 // 4K
)

type MessageProto struct {
	host myhost.Host

	data        ds.MessageIface
	sessionData sessionds.SessionIface

	groupConns map[string]map[peer.ID]struct{}

	emitters struct {
		evtReceiveGroupMessage event.Emitter
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

	h.SetStreamHandler(ID, msgsvc.messageHandler)
	h.SetStreamHandler(SYNC_ID, msgsvc.syncHandler)

	// 接收群消息
	if msgsvc.emitters.evtReceiveGroupMessage, err = eventBus.Emitter(&myevent.EvtReceiveGroupMessage{}); err != nil {
		return nil, fmt.Errorf("set receive group msg emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(myevent.EvtGroupConnectChange)}, eventbus.Name("syncmsg"))
	if err != nil {
		return nil, fmt.Errorf("subscription failed. group admin server error: %v", err)

	} else {
		go msgsvc.subscribeHandler(context.Background(), sub)
	}

	return msgsvc, nil
}

func (m *MessageProto) messageHandler(s network.Stream) {

	remotePeerID := s.Conn().RemotePeer()

	if err := s.Scope().SetService(ServiceName); err != nil {
		log.Errorf("failed to attaching stream to identify service: %v", err)
		s.Reset()
		return
	}

	rd := pbio.NewDelimitedReader(s, maxMsgSize)
	defer rd.Close()

	s.SetDeadline(time.Now().Add(StreamTimeout))

	var msg pb.GroupMessage
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		s.Reset()
		return
	}

	s.SetReadDeadline(time.Time{})

	// 检查本地是否存在
	ctx := context.Background()
	_, err := m.data.GetMessage(ctx, msg.GroupId, msg.Id)
	if err != nil {
		if errors.Is(err, ipfsds.ErrNotFound) {
			// 保存消息
			err = m.saveMessage(ctx, msg.GroupId, &msg)
			if err != nil {
				log.Errorf("save group message error: %v", err)
			}

			// 更新本地lamptime
			err = m.data.MergeLamportTime(ctx, msg.GroupId, msg.Lamportime)
			if err != nil {
				log.Errorf("update lamport time error: %v", err)
			}

			// 触发接收消息
			if err = m.emitters.evtReceiveGroupMessage.Emit(myevent.EvtReceiveGroupMessage{
				MsgID:      msg.Id,
				GroupID:    msg.GroupId,
				FromPeerID: peer.ID(msg.Member.Id),
				MsgType:    msg.MsgType,
				MimeType:   msg.MimeType,
				Payload:    msg.Payload,
				Timestamp:  msg.CreateTime,
			}); err != nil {
				log.Errorln("emit receive group msg error: %w", err)
			}

			// 转发消息
			err = m.broadcastMessage(ctx, msg.GroupId, &msg, remotePeerID)
			if err != nil {
				log.Errorf("emit forward group msg error: %v", err)
			}

		} else {
			log.Errorf("get group message error: %v", err)
		}

		return
	}
}

func (m *MessageProto) syncHandler(stream network.Stream) {
	defer stream.Close()

	// 获取同步GroupID
	var syncmsg pb.GroupSyncMessage
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
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
			}

		case <-ctx.Done():
			return
		}
	}
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

func (m *MessageProto) SendGroupMessage(ctx context.Context, account *mytype.Account, groupID string, msgType string, mimeType string, payload []byte) (string, error) {

	lamportime, err := m.data.TickLamportTime(context.Background(), groupID)
	if err != nil {
		return "", fmt.Errorf("ds tick lamptime error: %w", err)
	}

	msg := &pb.GroupMessage{
		Id:      msgID(lamportime, account.ID),
		GroupId: groupID,
		Member: &pb.GroupMessage_Member{
			Id:     []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MsgType:    msgType,
		MimeType:   mimeType,
		Payload:    payload,
		CreateTime: time.Now().Unix(),
		Lamportime: lamportime,
	}
	fmt.Println("12313w453")

	// 保存消息
	err = m.saveMessage(ctx, groupID, msg)
	if err != nil {
		return "", fmt.Errorf("save group message error: %v", err)
	}
	fmt.Println("56666")

	err = m.broadcastMessage(ctx, groupID, msg)
	if err != nil {
		return msg.Id, fmt.Errorf("broadcast message error: %w", err)
	}

	fmt.Println("9999")

	return msg.Id, nil
}

func (m *MessageProto) ClearGroupMessage(ctx context.Context, groupID string) error {
	return m.data.ClearMessage(ctx, groupID)
}

// 发送消息
func (m *MessageProto) broadcastMessage(ctx context.Context, groupID string, msg *pb.GroupMessage, excludePeerIDs ...peer.ID) error {

	peerIDs := m.getConnectPeers(groupID)
	if len(peerIDs) == 0 {
		return fmt.Errorf("no connect peers: %w", myerror.ErrSendGroupMessageFailed)
	}

	isSended := false
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

		if err := m.sendPeerMessage(ctx, groupID, peerID, msg); err == nil {
			isSended = true
		}
	}

	if !isSended {
		return myerror.ErrSendGroupMessageFailed
	}

	return nil
}

// 发送消息（指定peerID）
func (m *MessageProto) sendPeerMessage(ctx context.Context, groupID string, peerID peer.ID, msg *pb.GroupMessage) error {
	stream, err := m.host.NewStream(network.WithDialPeerTimeout(ctx, mytype.DialTimeout), peerID, ID)
	if err != nil {
		return err
	}

	stream.SetWriteDeadline(time.Now().Add(StreamTimeout))

	wt := pbio.NewDelimitedWriter(stream)
	defer wt.Close()

	if err := wt.WriteMsg(msg); err != nil {
		return err
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

func (m *MessageProto) saveMessage(ctx context.Context, groupID string, msg *pb.GroupMessage) error {
	isLatest, err := m.data.SaveMessage(ctx, groupID, msg)
	if err != nil {
		return fmt.Errorf("data save message error: %w", err)
	}

	sessionID := mytype.GroupSessionID(groupID)
	// 更新session时间
	if err = m.sessionData.UpdateSessionTime(ctx, sessionID.String()); err != nil {
		return fmt.Errorf("data.UpdateSessionTime error: %w", err)
	}

	// 更新session未读消息
	if msg.Member != nil && peer.ID(msg.Member.Id) == m.host.ID() {
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
		if msg.Member != nil {
			username = msg.Member.Name
		}

		var content string
		switch msg.MsgType {
		case mytype.MsgTypeText:
			content = string(msg.Payload)
		case mytype.MsgTypeImage:
			content = string("[image]")
		case mytype.MsgTypeAudio:
			content = string("[audio]")
		case mytype.MsgTypeVideo:
			content = string("[video]")
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
