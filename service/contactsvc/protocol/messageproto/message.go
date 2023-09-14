package message

import (
	"context"
	"fmt"
	"time"

	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/protocol"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/messageproto/ds"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/messageproto/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
	"google.golang.org/protobuf/proto"

	ipfsds "github.com/ipfs/go-datastore"
)

// 消息服务

var log = logging.Logger("message")

var StreamTimeout = 1 * time.Minute

const (
	ID      = protocol.ContactMessageID_v100
	SYNC_ID = protocol.ContactMessageSyncID_v100

	ServiceName = "peer.message"
	maxMsgSize  = 4 * 1024 // 4K
)

type PeerMessageProto struct {
	host host.Host
	data ds.PeerMessageIface

	emitters struct {
		evtPushOfflineMessage event.Emitter
		evtPullOfflineMessage event.Emitter
		evtReceiveMessage     event.Emitter
	}
}

func NewMessageSvc(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus) (*PeerMessageProto, error) {
	var err error
	msgsvc := PeerMessageProto{
		host: lhost,
		data: ds.Wrap(ids),
	}

	lhost.SetStreamHandler(ID, msgsvc.Handler)
	lhost.SetStreamHandler(SYNC_ID, msgsvc.SyncHandler)

	// 触发器：发送离线消息
	if msgsvc.emitters.evtPushOfflineMessage, err = eventBus.Emitter(&gevent.PushOfflineMessageEvt{}); err != nil {
		return nil, fmt.Errorf("set pull deposit msg emitter error: %v", err)
	}

	// 触发器：获取离线消息
	if msgsvc.emitters.evtPullOfflineMessage, err = eventBus.Emitter(&gevent.PullOfflineMessageEvt{}); err != nil {
		return nil, fmt.Errorf("set pull deposit msg emitter error: %v", err)
	}

	// 触发器：接收到消息
	if msgsvc.emitters.evtReceiveMessage, err = eventBus.Emitter(&gevent.EvtReceivePeerMessage{}); err != nil {
		return nil, fmt.Errorf("set receive msg emitter error: %v", err)
	}

	// 订阅：同步Peer的消息
	gsubs, err := eventBus.Subscribe([]any{new(gevent.EvtSyncPeers)})
	if err != nil {
		return nil, fmt.Errorf("subscribe boot complete event error: %v", err)

	} else {
		go msgsvc.handleAppSubs(context.Background(), gsubs)
	}

	// 订阅：开始拉取离线消息事件
	hsubs, err := lhost.EventBus().Subscribe([]any{new(event.EvtLocalAddressesUpdated)})
	if err != nil {
		return nil, fmt.Errorf("subscribe error: %v", err)

	} else {
		go msgsvc.handleHostSubs(context.Background(), hsubs)
	}

	return &msgsvc, nil
}

func (p *PeerMessageProto) handleAppSubs(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				log.Warnf("subscribe out not ok")
				return
			}

			fmt.Println("start sync peers message")

			if ev, ok := e.(gevent.EvtSyncPeers); ok {
				fmt.Printf("peerIDs: %v", ev.PeerIDs)
				for _, peerID := range ev.PeerIDs {
					p.RunSync(peerID)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func (p *PeerMessageProto) handleHostSubs(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				log.Warnf("subscribe out not ok")
				return
			}
			ev, ok := e.(event.EvtLocalAddressesUpdated)
			if !ok || !ev.Diffs {
				log.Warnf("event local address no ok or no diff")
				return
			}

			for _, curr := range ev.Current {
				if isPublicAddr(curr.Address) {
					p.emitters.evtPullOfflineMessage.Emit(gevent.PullOfflineMessageEvt{
						HasMessage:  p.HasMessage,
						SaveMessage: p.SaveMessage,
					})
					break
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func (p *PeerMessageProto) Handler(stream network.Stream) {

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rd.Close()

	stream.SetDeadline(time.Now().Add(StreamTimeout))

	var msg pb.Message
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		stream.Reset()
		return
	}

	stream.SetReadDeadline(time.Time{})

	remotePeerID := stream.Conn().RemotePeer()

	err := p.data.SaveMessage(context.Background(), remotePeerID, &msg)
	if err != nil {
		log.Errorf("store message error %v", err)
		stream.Reset()
		return
	}

	err = p.data.MergeLamportTime(context.Background(), remotePeerID, msg.Lamportime)
	if err != nil {
		log.Errorf("update lamport time error: %v", err)
		stream.Reset()
		return
	}

	p.emitters.evtReceiveMessage.Emit(gevent.EvtReceivePeerMessage{
		MsgID:      msg.Id,
		FromPeerID: peer.ID(msg.FromPeerId),
		MsgType:    gevent.MsgTypeText,
		MimeType:   "text/plain",
		Payload:    msg.Payload,
		Timestamp:  msg.Timestamp,
	})

	fmt.Println("receive ->", string(msg.Payload))
}

func (p *PeerMessageProto) HasMessage(peerID peer.ID, msgID string) (bool, error) {
	return p.data.HasMessage(context.Background(), peerID, msgID)
}

func (p *PeerMessageProto) SaveMessage(peerID peer.ID, msgID string, msgData []byte) error {
	var pmsg pb.Message
	if err := proto.Unmarshal(msgData, &pmsg); err != nil {
		return err

	} else if pmsg.Id != msgID {
		return fmt.Errorf("msg id not equal")
	}

	if err := p.data.SaveMessage(context.Background(), peerID, &pmsg); err != nil {
		return err
	}

	return nil
}

func (p *PeerMessageProto) SendMessage(ctx context.Context, peerID peer.ID, msgType pb.Message_MsgType, mimeType string, payload []byte) error {

	hostID := p.host.ID()

	lamportTime, err := p.data.TickLamportTime(ctx, peerID)
	if err != nil {
		return err
	}

	pmsg := pb.Message{
		Id:         msgID(lamportTime, hostID),
		FromPeerId: []byte(hostID),
		ToPeerId:   []byte(peerID),
		MsgType:    msgType,
		MimeType:   mimeType,
		Payload:    payload,
		Timestamp:  time.Now().Unix(),
		Lamportime: lamportTime,
	}

	err = p.data.SaveMessage(ctx, peerID, &pmsg)
	if err != nil {
		return err
	}

	stream, err := p.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, time.Second), ""), peerID, ID)
	if err != nil {
		// 连接失败，则走离线服务
		msgdata, _ := proto.Marshal(&pmsg)
		err = p.emitters.evtPushOfflineMessage.Emit(gevent.PushOfflineMessageEvt{
			ToPeerID: peerID,
			MsgID:    pmsg.Id,
			MsgData:  msgdata,
		})
		if err != nil {
			return fmt.Errorf("push offline error: %v", err)
		}

		fmt.Println("emit offline msg ->", string(pmsg.Payload))

		return nil
	}

	pw := pbio.NewDelimitedWriter(stream)
	defer pw.Close()

	if err = pw.WriteMsg(&pmsg); err != nil {
		stream.Reset()
		return err
	}

	fmt.Println("send ->", string(pmsg.Payload))

	return nil
}

func (p *PeerMessageProto) GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*pb.Message, error) {
	return p.data.GetMessage(ctx, peerID, msgID)
}

func (p *PeerMessageProto) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]*pb.Message, error) {
	return p.data.GetMessages(ctx, peerID, offset, limit)
}

func (p *PeerMessageProto) ClearMessage(ctx context.Context, peerID peer.ID) error {
	return p.data.ClearMessage(ctx, peerID)
}