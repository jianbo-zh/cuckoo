package message

import (
	"context"
	"errors"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/service/group/protocol/message/ds"
	"github.com/jianbo-zh/dchat/service/group/protocol/message/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/libp2p/go-msgio/pbio"
	"google.golang.org/protobuf/proto"
)

var log = logging.Logger("message")

var StreamTimeout = 1 * time.Minute

const (
	ID      = "/dchat/group/msg/1.0.0"
	SYNC_ID = "/dchat/group/syncmsg/1.0.0"

	ServiceName = "group.message"
	maxMsgSize  = 4 * 1024 // 4K
)

type MessageService struct {
	host host.Host

	data ds.MessageIface

	emitters struct {
		evtForwardGroupMsg event.Emitter
	}

	groupConns map[string]map[peer.ID]struct{}
}

func NewGroupMessageService(h host.Host, ids ipfsds.Batching, eventBus event.Bus) (*MessageService, error) {
	var err error
	msgsvc := &MessageService{
		host:       h,
		data:       ds.MessageWrap(ids),
		groupConns: make(map[string]map[peer.ID]struct{}),
	}

	h.SetStreamHandler(ID, msgsvc.messageHandler)
	h.SetStreamHandler(SYNC_ID, msgsvc.syncHandler)

	if msgsvc.emitters.evtForwardGroupMsg, err = eventBus.Emitter(&gevent.EvtForwardGroupMsg{}); err != nil {
		return nil, fmt.Errorf("set group msg emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(gevent.EvtGroupConnectChange)}, eventbus.Name("syncmsg"))
	if err != nil {
		return nil, fmt.Errorf("subscription failed. group admin server error: %v", err)

	} else {
		go msgsvc.subscribeHandler(context.Background(), sub)
	}

	return msgsvc, nil
}

func (m *MessageService) messageHandler(s network.Stream) {

	remotePeerID := s.Conn().RemotePeer()

	if err := s.Scope().SetService(ServiceName); err != nil {
		log.Errorf("failed to attaching stream to identify service: %v", err)
		s.Reset()
		return
	}

	rd := pbio.NewDelimitedReader(s, maxMsgSize)
	defer rd.Close()

	s.SetDeadline(time.Now().Add(StreamTimeout))

	var msg pb.GroupMsg
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		s.Reset()
		return
	}

	s.SetReadDeadline(time.Time{})

	// 检查本地是否存在
	_, err := m.data.GetMessage(context.Background(), ds.GroupID(msg.GroupId), msg.Id)
	if err != nil {
		if errors.Is(err, ipfsds.ErrNotFound) {
			// 保存消息
			err = m.data.SaveMessage(context.Background(), ds.GroupID(msg.GroupId), &msg)
			if err != nil {
				log.Errorf("save group message error: %v", err)
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

func (m *MessageService) syncHandler(stream network.Stream) {
	defer stream.Close()

	// 获取同步GroupID
	var syncmsg pb.GroupSyncMessage
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	if err := rd.ReadMsg(&syncmsg); err != nil {
		log.Errorf("read msg error: %v", err)
		return
	}

	groupID := syncmsg.GroupId

	// 发送同步摘要
	summary, err := m.getMessageSummary(groupID)
	if err != nil {
		return
	}

	bs, err := proto.Marshal(summary)
	if err != nil {
		return
	}

	wt := pbio.NewDelimitedWriter(stream)
	if err = wt.WriteMsg(&pb.GroupSyncMessage{
		Type:    pb.GroupSyncMessage_SUMMARY,
		Payload: bs,
	}); err != nil {
		return
	}

	// 后台同步处理
	err = m.loopSync(groupID, stream, rd, wt)
	if err != nil {
		log.Errorf("loop sync error: %v", err)
	}
}

func (m *MessageService) subscribeHandler(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}

			evt := e.(gevent.EvtGroupConnectChange)

			if !evt.IsConnected { // 断开连接
				delete(m.groupConns[evt.GroupID], evt.PeerID)
			}

			// 新建连接
			if _, exists := m.groupConns[evt.GroupID]; !exists {
				m.groupConns[evt.GroupID] = make(map[peer.ID]struct{})
			}
			m.groupConns[evt.GroupID][evt.PeerID] = struct{}{}

			// 启动同步
			m.RunSync(evt.GroupID, evt.PeerID)

		case <-ctx.Done():
			return
		}
	}
}

func (m *MessageService) SendTextMessage(ctx context.Context, groupID string, msg string) error {

	peerID := m.host.ID().String()
	lamportime, err := m.data.TickLamportTime(context.Background(), ds.GroupID(groupID))
	if err != nil {
		return err
	}
	msg1 := pb.GroupMsg{
		Id:         fmt.Sprintf("%d_%s", lamportime, peerID),
		PeerId:     peerID,
		Type:       pb.GroupMsg_TEXT,
		Payload:    []byte(msg),
		Timestamp:  time.Now().Unix(),
		Lamportime: lamportime,
	}

	err = m.broadcastMessage(ctx, groupID, &msg1)
	if err != nil {
		return err
	}

	return nil
}

// 发送消息
func (m *MessageService) broadcastMessage(ctx context.Context, groupID string, msg *pb.GroupMsg, excludePeerIDs ...peer.ID) error {

	for _, peerID := range m.getConnectPeers(groupID) {
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
func (m *MessageService) sendPeerMessage(ctx context.Context, groupID string, peerID peer.ID, msg *pb.GroupMsg) error {
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

func (m *MessageService) getConnectPeers(groupID string) []peer.ID {
	var peerIDs []peer.ID
	for peerID := range m.groupConns[groupID] {
		peerIDs = append(peerIDs, peerID)
	}
	return peerIDs
}
