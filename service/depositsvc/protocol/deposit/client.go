package deposit

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/myerror"
	"github.com/jianbo-zh/dchat/service/depositsvc/protocol/deposit/ds"
	"github.com/jianbo-zh/dchat/service/depositsvc/protocol/deposit/pb"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/libp2p/go-msgio/pbio"
)

type DepositClientProto struct {
	host host.Host

	datastore ds.DepositMessageIface
}

func NewDepositClientProto(ctx context.Context, h host.Host, ids ipfsds.Batching, eventBus event.Bus) (*DepositClientProto, error) {
	gcli := &DepositClientProto{
		host:      h,
		datastore: ds.DepositPeerWrap(ids),
	}

	// 订阅push、pull
	sub, err := eventBus.Subscribe([]any{
		new(gevent.PushDepositContactMessageEvt), new(gevent.PushDepositGroupMessageEvt),
		new(gevent.PullDepositContactMessageEvt), new(gevent.PullDepositGroupMessageEvt)}, eventbus.Name("deposit"))
	if err != nil {
		return nil, err

	} else {
		go gcli.subscribeHandler(ctx, sub)
	}

	return gcli, nil
}

func (d *DepositClientProto) subscribeHandler(ctx context.Context, sub event.Subscription) {

	defer func() {
		log.Error("peer client subscrib exit")
		sub.Close()
	}()

	for {
		select {
		case e, ok := <-sub.Out():
			log.Debugf("get subscribe: %v", ok)
			if !ok {
				return
			}
			switch evt := e.(type) {
			case gevent.PushDepositContactMessageEvt:
				go d.handlePushContactEvent(evt)

			case gevent.PushDepositGroupMessageEvt:
				go d.handlePushGroupEvent(evt)

			case gevent.PullDepositContactMessageEvt:
				go d.handlePullContactMessageEvent(evt)

			case gevent.PullDepositGroupMessageEvt:
				go d.handlePullGroupMessageEvent(evt)
			default:
				log.Warnf("undefined event type: %T", evt)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (d *DepositClientProto) handlePushContactEvent(ev gevent.PushDepositContactMessageEvt) {
	log.Debugf("handle push contact msg event: %s", ev.MsgID)
	err := d.PushContactMessage(ev.DepositAddress, ev.ToPeerID, ev.MsgID, ev.MsgData)
	if err != nil {
		log.Errorf("push contact message error: %w", err)

		if ev.ResultCallback != nil {
			ev.ResultCallback(ev.ToPeerID, ev.MsgID, fmt.Errorf("push contact message error: %w", err))
		}
		return
	}

	if ev.ResultCallback != nil {
		ev.ResultCallback(ev.ToPeerID, ev.MsgID, nil)
	}
}

func (d *DepositClientProto) handlePushGroupEvent(ev gevent.PushDepositGroupMessageEvt) {
	log.Debugf("handle push group msg event: %s", ev.MsgID)
	err := d.PushGroupMessage(ev.DepositAddress, ev.ToGroupID, ev.MsgID, ev.MsgData)
	if err != nil {
		log.Errorf("push group message error: %w", err)

		if ev.ResultCallback != nil {
			ev.ResultCallback(ev.ToGroupID, ev.MsgID, fmt.Errorf("push group message error: %w", err))
		}
		return
	}

	if ev.ResultCallback != nil {
		ev.ResultCallback(ev.ToGroupID, ev.MsgID, nil)
	}
}

func (d *DepositClientProto) handlePullContactMessageEvent(evt gevent.PullDepositContactMessageEvt) {
	log.Debugf("receive pull offline event")

	hostID := d.host.ID()

	lastDepositID, err := d.datastore.GetContactLastID(hostID)
	if err != nil {
		log.Errorf("data contact last id error: %v", err)
		return
	}

	depositPeerID := evt.DepositAddress
	stream, err := d.host.NewStream(network.WithUseTransient(context.Background(), ""), depositPeerID, CONTACT_GET_ID)
	if err != nil {
		log.Errorf("new stream to deposit error: %v", err)
		return
	}
	defer stream.Close()

	pr := pbio.NewDelimitedReader(stream, maxMsgSize)
	pw := pbio.NewDelimitedWriter(stream)

	if err = pw.WriteMsg(&pb.ContactMessagePull{StartId: lastDepositID}); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	var dmsg pb.ContactMessage
	for {
		dmsg.Reset()
		if err = pr.ReadMsg(&dmsg); err != nil {
			log.Errorf("receive deposit msg error: %v", err)
			stream.Reset()
			return
		}

		if evt.MessageHandler != nil {
			if err = evt.MessageHandler(peer.ID(dmsg.FromPeerId), dmsg.MsgId, dmsg.MsgData); err != nil {
				log.Errorf("handle deposit msg error: %v", err)
				stream.Reset()
				return
			}
		}

		if err = d.datastore.SetContactLastID(hostID, dmsg.Id); err != nil {
			log.Errorf("set contact last id error: %v", err)
			stream.Reset()
			return
		}
	}
}

func (d *DepositClientProto) handlePullGroupMessageEvent(evt gevent.PullDepositGroupMessageEvt) {
	log.Debugf("receive pull offline event")

	groupID := evt.GroupID
	depositPeerID := evt.DepositAddress

	lastDepositID, err := d.datastore.GetGroupLastID(evt.GroupID)
	if err != nil {
		log.Errorf("data contact last id error: %v", err)
		return
	}

	fmt.Println("deposit peerID: ", depositPeerID.String())

	stream, err := d.host.NewStream(network.WithUseTransient(context.Background(), ""), depositPeerID, GROUP_GET_ID)
	if err != nil {
		log.Errorf("new stream to deposit error: %v", err)
		return
	}
	defer stream.Close()

	pr := pbio.NewDelimitedReader(stream, maxMsgSize)
	pw := pbio.NewDelimitedWriter(stream)

	if err = pw.WriteMsg(&pb.GroupMessagePull{GroupId: groupID, StartId: lastDepositID}); err != nil {
		log.Errorf("pbio write msg error: %v", err)
		stream.Reset()
		return
	}

	var msg pb.GroupMessage
	for {
		msg.Reset()
		if err = pr.ReadMsg(&msg); err != nil {
			log.Errorf("pbio read msg error: %v", err)
			stream.Reset()
			return
		}

		if evt.MessageHandler != nil {
			if err = evt.MessageHandler(groupID, msg.MsgId, msg.MsgData); err != nil {
				log.Errorf("message handler error: %v", err)
				stream.Reset()
				return
			}
		}

		if err = d.datastore.SetGroupLastID(groupID, msg.Id); err != nil {
			log.Errorf("data set group last id error: %v", err)
			stream.Reset()
			return
		}
	}
}

func (d *DepositClientProto) PushContactMessage(depositPeerID peer.ID, toPeerID peer.ID, msgID string, msgData []byte) error {
	hostID := d.host.ID()

	log.Debugf("get deposit service peer: %s", depositPeerID.String())

	stream, err := d.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(context.Background(), time.Second), ""), depositPeerID, CONTACT_SAVE_ID)
	if err != nil {
		return myerror.WrapStreamError("new stream to deposit peer error", err)
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	wt := pbio.NewDelimitedWriter(stream)

	if err = wt.WriteMsg(&pb.ContactMessage{
		FromPeerId:  []byte(hostID),
		ToPeerId:    []byte(toPeerID),
		MsgId:       msgID,
		MsgData:     msgData,
		DepositTime: time.Now().Unix(),
	}); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("write msg error", err)
	}

	var msgAck pb.MessageAck
	if err = rd.ReadMsg(&msgAck); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("read ack msg error", err)

	} else if msgAck.MsgId != msgID {
		stream.Reset()
		return myerror.WrapStreamError("msg ack id error", nil)
	}

	return nil
}

func (d *DepositClientProto) PushGroupMessage(depositPeerID peer.ID, groupID string, msgID string, msgData []byte) error {
	hostID := d.host.ID()

	log.Debugf("get deposit service peer: %s", depositPeerID.String())

	stream, err := d.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(context.Background(), time.Second), ""), depositPeerID, GROUP_SAVE_ID)
	if err != nil {
		return myerror.WrapStreamError("new stream to deposit peer error", err)
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	pw := pbio.NewDelimitedWriter(stream)

	dmsg := pb.GroupMessage{
		FromPeerId:  []byte(hostID),
		GroupID:     groupID,
		MsgId:       msgID,
		MsgData:     msgData,
		DepositTime: time.Now().Unix(),
	}

	if err = pw.WriteMsg(&dmsg); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("write msg error", err)
	}

	var msgAck pb.MessageAck
	if err = rd.ReadMsg(&msgAck); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("pbio read ack msg error: %w", err)

	} else if msgAck.MsgId != msgID {
		stream.Reset()
		return myerror.WrapStreamError("msg ack id error", nil)
	}

	return nil
}
