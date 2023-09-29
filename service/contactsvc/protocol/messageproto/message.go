package message

import (
	"context"
	"fmt"
	"time"

	"github.com/jianbo-zh/dchat/internal/myerror"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/protocol"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/messageproto/ds"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/messageproto/pb"
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
	ID      = protocol.ContactMessageID_v100
	SYNC_ID = protocol.ContactMessageSyncID_v100

	ServiceName = "peer.message"
	maxMsgSize  = 4 * 1024 // 4K
)

type PeerMessageProto struct {
	host myhost.Host
	data ds.PeerMessageIface

	emitters struct {
		evtPushOfflineMessage event.Emitter
		evtPullOfflineMessage event.Emitter
		evtReceiveMessage     event.Emitter
	}
}

func NewMessageSvc(lhost myhost.Host, ids ipfsds.Batching, eventBus event.Bus) (*PeerMessageProto, error) {
	var err error
	msgsvc := PeerMessageProto{
		host: lhost,
		data: ds.Wrap(ids),
	}

	lhost.SetStreamHandler(ID, msgsvc.Handler)
	lhost.SetStreamHandler(SYNC_ID, msgsvc.SyncHandler)

	// 触发器：发送离线消息
	if msgsvc.emitters.evtPushOfflineMessage, err = eventBus.Emitter(&myevent.EvtPushDepositContactMessage{}); err != nil {
		return nil, fmt.Errorf("set pull deposit msg emitter error: %v", err)
	}

	// 触发器：获取离线消息
	if msgsvc.emitters.evtPullOfflineMessage, err = eventBus.Emitter(&myevent.EvtPullDepositContactMessage{}); err != nil {
		return nil, fmt.Errorf("set pull deposit msg emitter error: %v", err)
	}

	// 触发器：接收到消息
	if msgsvc.emitters.evtReceiveMessage, err = eventBus.Emitter(&myevent.EvtReceivePeerMessage{}); err != nil {
		return nil, fmt.Errorf("set receive msg emitter error: %v", err)
	}

	// 订阅：同步Peer的消息，拉取离线消息事件
	sub, err := eventBus.Subscribe([]any{new(myevent.EvtSyncPeerMessage), new(event.EvtLocalAddressesUpdated)})
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
			case myevent.EvtSyncPeerMessage:
				log.Debugln("event sync peer")
				go p.goSync(evt.ContactID)

			case event.EvtLocalAddressesUpdated:
				if !evt.Diffs {
					log.Warnf("event local address no ok or no diff")
					return
				}

				for _, curr := range evt.Current {
					if isPublicAddr(curr.Address) {
						p.emitters.evtPullOfflineMessage.Emit(myevent.EvtPullDepositContactMessage{
							DepositAddress: peer.ID(""),
							MessageHandler: p.SaveMessage,
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

func (p *PeerMessageProto) Handler(stream network.Stream) {

	// 开始处理流
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

	wt := pbio.NewDelimitedWriter(stream)
	if err = wt.WriteMsg(&pb.MessageAck{Id: msg.Id}); err != nil {
		log.Errorf("ack msg error: %w", err)
		stream.Reset()
		return
	}

	err = p.data.MergeLamportTime(context.Background(), remotePeerID, msg.Lamportime)
	if err != nil {
		log.Errorf("update lamport time error: %v", err)
		stream.Reset()
		return
	}

	p.emitters.evtReceiveMessage.Emit(myevent.EvtReceivePeerMessage{
		MsgID:      msg.Id,
		FromPeerID: peer.ID(msg.FromPeerId),
		MsgType:    msg.MsgType,
		MimeType:   msg.MimeType,
		Payload:    msg.Payload,
		Timestamp:  msg.Timestamp,
	})

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

func (p *PeerMessageProto) SendMessage(ctx context.Context, peerID peer.ID, msgType string, mimeType string, payload []byte) (string, error) {

	hostID := p.host.ID()

	lamportTime, err := p.data.TickLamportTime(ctx, peerID)
	if err != nil {
		return "", fmt.Errorf("data tick lamptime error: %w", err)
	}

	msg := pb.Message{
		Id:         msgID(lamportTime, hostID),
		FromPeerId: []byte(hostID),
		ToPeerId:   []byte(peerID),
		MsgType:    msgType,
		MimeType:   mimeType,
		Payload:    payload,
		Timestamp:  time.Now().Unix(),
		Lamportime: lamportTime,
	}

	err = p.data.SaveMessage(ctx, peerID, &msg)
	if err != nil {
		return "", fmt.Errorf("data save msg error: %w", err)
	}

	stream, err := p.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, time.Second), ""), peerID, ID)
	if err != nil {
		return msg.Id, myerror.WrapStreamError("host new stream error", err)
	}

	pw := pbio.NewDelimitedWriter(stream)
	defer pw.Close()

	if err = pw.WriteMsg(&msg); err != nil {
		stream.Reset()
		return msg.Id, myerror.WrapStreamError("pbio write msg error", err)
	}

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	var msgAck pb.MessageAck
	if err = rd.ReadMsg(&msgAck); err != nil {
		return msg.Id, myerror.WrapStreamError("pbio read ack msg error", err)

	} else if msgAck.Id != msg.Id {
		return msg.Id, myerror.WrapStreamError("msg ack id error", nil)
	}

	return msg.Id, nil
}

func (p *PeerMessageProto) GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*pb.Message, error) {
	return p.data.GetMessage(ctx, peerID, msgID)
}

func (p *PeerMessageProto) DeleteMessage(ctx context.Context, peerID peer.ID, msgID string) error {
	return p.data.DeleteMessage(ctx, peerID, msgID)
}

func (p *PeerMessageProto) GetMessageData(ctx context.Context, peerID peer.ID, msgID string) ([]byte, error) {
	return p.data.GetMessageData(ctx, peerID, msgID)
}

func (p *PeerMessageProto) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]*pb.Message, error) {
	return p.data.GetMessages(ctx, peerID, offset, limit)
}

func (p *PeerMessageProto) ClearMessage(ctx context.Context, peerID peer.ID) error {
	return p.data.ClearMessage(ctx, peerID)
}
