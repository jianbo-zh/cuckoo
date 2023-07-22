package message

import (
	"context"
	"fmt"
	"time"

	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/service/peer/protocol/message/ds"
	"github.com/jianbo-zh/dchat/service/peer/protocol/message/pb"
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
	ID      = "/dchat/peer/msg/1.0.0"
	SYNC_ID = "/dchat/peer/msg/sync/1.0.0"

	ServiceName = "peer.message"
	maxMsgSize  = 4 * 1024 // 4K
)

type PeerMessageSvc struct {
	host host.Host
	data ds.PeerMessageIface

	emitters struct {
		evtPushOfflineMessage event.Emitter
		evtPullOfflineMessage event.Emitter
	}
}

func NewMessageSvc(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus) (*PeerMessageSvc, error) {
	var err error
	msgsvc := PeerMessageSvc{
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

func (p *PeerMessageSvc) handleAppSubs(ctx context.Context, sub event.Subscription) {
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

func (p *PeerMessageSvc) handleHostSubs(ctx context.Context, sub event.Subscription) {
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

func (p *PeerMessageSvc) Handler(stream network.Stream) {

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

	err := p.data.SaveMessage(context.Background(), stream.Conn().RemotePeer(), &msg)
	if err != nil {
		log.Errorf("store message error %v", err)
		stream.Reset()
		return
	}

	err = p.data.MergeLamportTime(context.Background(), stream.Conn().RemotePeer(), msg.Lamportime)
	if err != nil {
		log.Errorf("update lamport time error: %v", err)
		stream.Reset()
		return
	}

	fmt.Println("receive ->", string(msg.Payload))
}

func (p *PeerMessageSvc) HasMessage(peerID peer.ID, msgID string) (bool, error) {
	return p.data.HasMessage(context.Background(), peerID, msgID)
}

func (p *PeerMessageSvc) SaveMessage(peerID peer.ID, msgID string, msgData []byte) error {
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

func (p *PeerMessageSvc) SendTextMessage(ctx context.Context, peerID peer.ID, msg string) error {

	hostID := p.host.ID()

	lamportTime, err := p.data.TickLamportTime(ctx, peerID)
	if err != nil {
		return err
	}

	pmsg := pb.Message{
		Id:         msgID(lamportTime, hostID),
		Type:       pb.Message_TEXT,
		Payload:    []byte(msg),
		SenderId:   []byte(hostID),
		ReceiverId: []byte(peerID),
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

func (p *PeerMessageSvc) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]*pb.Message, error) {
	return p.data.GetMessages(ctx, peerID, offset, limit)
}

func (p *PeerMessageSvc) SendGroupInviteMessage(ctx context.Context, peerID peer.ID, groupID string) error {

	hostID := p.host.ID()

	stream, err := p.host.NewStream(ctx, peerID, ID)
	if err != nil {
		return err
	}

	pw := pbio.NewDelimitedWriter(stream)

	lamportTime, err := p.data.TickLamportTime(ctx, peerID)
	if err != nil {
		stream.Reset()
		return err
	}

	pmsg := pb.Message{
		Id:         msgID(lamportTime, hostID),
		Type:       pb.Message_INVITE,
		Payload:    []byte(groupID),
		SenderId:   []byte(hostID),
		ReceiverId: []byte(peerID),
		Lamportime: lamportTime,
		Timestamp:  time.Now().Unix(),
	}

	err = p.data.SaveMessage(ctx, peerID, &pmsg)
	if err != nil {
		stream.Reset()
		return err
	}

	err = pw.WriteMsg(&pmsg)
	if err != nil {
		stream.Reset()
		return err
	}

	return nil
}
