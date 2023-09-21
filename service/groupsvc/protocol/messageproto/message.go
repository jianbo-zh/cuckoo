package messageproto

import (
	"context"
	"errors"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/protocol"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/messageproto/ds"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/messageproto/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
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
	host host.Host

	data ds.MessageIface

	groupConns map[string]map[peer.ID]struct{}

	emitters struct {
		evtReceiveGroupMessage event.Emitter
	}
}

func NewMessageProto(h host.Host, ids ipfsds.Batching, eventBus event.Bus) (*MessageProto, error) {
	var err error
	msgsvc := &MessageProto{
		host:       h,
		data:       ds.MessageWrap(ids),
		groupConns: make(map[string]map[peer.ID]struct{}),
	}

	h.SetStreamHandler(ID, msgsvc.messageHandler)
	h.SetStreamHandler(SYNC_ID, msgsvc.syncHandler)

	// 接收群消息
	if msgsvc.emitters.evtReceiveGroupMessage, err = eventBus.Emitter(&gevent.EvtReceiveGroupMessage{}); err != nil {
		return nil, fmt.Errorf("set receive group msg emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(gevent.EvtGroupConnectChange)}, eventbus.Name("syncmsg"))
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

	var msg pb.Message
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		s.Reset()
		return
	}

	s.SetReadDeadline(time.Time{})

	// 检查本地是否存在
	_, err := m.data.GetMessage(context.Background(), msg.GroupId, msg.Id)
	if err != nil {
		if errors.Is(err, ipfsds.ErrNotFound) {
			// 保存消息
			err = m.data.SaveMessage(context.Background(), msg.GroupId, &msg)
			if err != nil {
				log.Errorf("save group message error: %v", err)
			}

			// 更新本地lamptime
			err = m.data.MergeLamportTime(context.Background(), msg.GroupId, msg.Lamportime)
			if err != nil {
				log.Errorf("update lamport time error: %v", err)
			}

			// 触发接收消息
			if err = m.emitters.evtReceiveGroupMessage.Emit(gevent.EvtReceiveGroupMessage{
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
			err = m.broadcastMessage(context.Background(), msg.GroupId, &msg, remotePeerID)
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
			case gevent.EvtGroupConnectChange:
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

func (m *MessageProto) GetMessage(ctx context.Context, groupID string, msgID string) (*pb.Message, error) {
	msgs, err := m.data.GetMessage(ctx, groupID, msgID)
	if err != nil {
		return nil, fmt.Errorf("m.data.ListMessages error: %w", err)
	}

	return msgs, nil
}

func (m *MessageProto) GetMessageList(ctx context.Context, groupID string, offset int, limit int) ([]*pb.Message, error) {
	msgs, err := m.data.GetMessages(ctx, groupID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("m.data.ListMessages error: %w", err)
	}

	return msgs, nil
}

func (m *MessageProto) SendGroupMessage(ctx context.Context, account *types.Account, groupID string, msgType string, mimeType string, payload []byte) (*pb.Message, error) {

	lamportime, err := m.data.TickLamportTime(context.Background(), groupID)
	if err != nil {
		return nil, fmt.Errorf("ds tick lamptime error: %w", err)
	}

	msg := &pb.Message{
		Id:      msgID(lamportime, account.ID),
		GroupId: groupID,
		Member: &pb.Message_Member{
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

	// 保存消息
	err = m.data.SaveMessage(context.Background(), groupID, msg)
	if err != nil {
		return nil, fmt.Errorf("save group message error: %v", err)
	}

	err = m.broadcastMessage(ctx, groupID, msg)
	if err != nil {
		return nil, fmt.Errorf("m.broadcast message error: %w", err)
	}

	return msg, nil
}

func (m *MessageProto) ClearGroupMessage(ctx context.Context, groupID string) error {
	return m.data.ClearMessage(ctx, groupID)
}

// 发送消息
func (m *MessageProto) broadcastMessage(ctx context.Context, groupID string, msg *pb.Message, excludePeerIDs ...peer.ID) error {

	connectPeerIDs := m.getConnectPeers(groupID)
	fmt.Printf("get connect peers: %v\n", connectPeerIDs)

	for _, peerID := range connectPeerIDs {
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

		m.sendPeerMessage(ctx, groupID, peerID, msg)
	}

	return nil
}

// 发送消息（指定peerID）
func (m *MessageProto) sendPeerMessage(ctx context.Context, groupID string, peerID peer.ID, msg *pb.Message) error {
	stream, err := m.host.NewStream(ctx, peerID, ID)
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
