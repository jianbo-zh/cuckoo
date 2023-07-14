package message

import (
	"context"
	"fmt"
	"time"

	sevent "github.com/jianbo-zh/dchat/service/event"
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

func NewPeerMessageSvc(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus) *PeerMessageSvc {
	var err error
	msgsvc := PeerMessageSvc{
		host: lhost,
		data: ds.Wrap(ids),
	}

	lhost.SetStreamHandler(ID, msgsvc.Handler)
	lhost.SetStreamHandler(SYNC_ID, msgsvc.SyncHandler)

	// 发送：离线消息
	if msgsvc.emitters.evtPushOfflineMessage, err = eventBus.Emitter(&sevent.PushOfflineMessageEvt{}); err != nil {
		log.Errorf("set pull deposit msg emitter error: %v", err)
	}

	// 获取：离线消息
	if msgsvc.emitters.evtPullOfflineMessage, err = eventBus.Emitter(&sevent.PullOfflineMessageEvt{}); err != nil {
		log.Errorf("set pull deposit msg emitter error: %v", err)
	}

	subs, err := lhost.EventBus().Subscribe([]any{new(event.EvtLocalAddressesUpdated)})
	if err != nil {
		log.Errorf("subscribe error: %v", err)

	} else {
		go msgsvc.background(context.Background(), subs)
	}

	return &msgsvc
}

func (p *PeerMessageSvc) background(ctx context.Context, sub event.Subscription) {
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
					p.emitters.evtPullOfflineMessage.Emit(sevent.PullOfflineMessageEvt{
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

func (p *PeerMessageSvc) HasMessage(peerID string, msgID string) (bool, error) {
	peerID1, err := peer.Decode(peerID)
	if err != nil {
		return false, err
	}

	return p.data.HasMessage(context.Background(), peerID1, msgID)
}

func (p *PeerMessageSvc) SaveMessage(peerID string, msgID string, msgData []byte) error {
	var pmsg pb.Message
	if err := proto.Unmarshal(msgData, &pmsg); err != nil {
		return err

	} else if pmsg.Id != msgID {
		return fmt.Errorf("msg id not equal")
	}

	peerID1, err := peer.Decode(peerID)
	if err != nil {
		return err
	}

	if err = p.data.SaveMessage(context.Background(), peerID1, &pmsg); err != nil {
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
		Id:         fmt.Sprintf("%d_%s", lamportTime, hostID.String()),
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

	stream, err := p.host.NewStream(network.WithUseTransient(ctx, ""), peerID, ID)
	if err != nil {
		// // if not connect
		// if errors.Is(err, network.ErrNoConn) {
		// 	msgdata, _ := proto.Marshal(&pmsg)
		// 	p.emitters.evtPushOfflineMessage.Emit(sevent.PushOfflineMessageEvt{
		// 		ToPeerID: peerID.String(),
		// 		MsgID:    pmsg.Id,
		// 		MsgData:  msgdata,
		// 	})

		// 	return nil
		// }

		return err
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
		Id:         fmt.Sprintf("%d_%s", lamportTime, hostID.String()),
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
