package deposit

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myerror"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	ds "github.com/jianbo-zh/dchat/service/depositsvc/datastore/ds/depositds"
	pb "github.com/jianbo-zh/dchat/service/depositsvc/protobuf/pb/depositpb"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/libp2p/go-msgio/pbio"
)

type DepositClientProto struct {
	host myhost.Host

	datastore ds.DepositMessageIface
}

func NewDepositClientProto(ctx context.Context, h myhost.Host, ids ipfsds.Batching, eventBus event.Bus) (*DepositClientProto, error) {
	gcli := &DepositClientProto{
		host:      h,
		datastore: ds.DepositPeerWrap(ids),
	}

	// 订阅push、pull
	sub, err := eventBus.Subscribe([]any{
		new(myevent.EvtPushDepositContactMessage), new(myevent.EvtPushDepositGroupMessage),
		new(myevent.EvtPullDepositContactMessage), new(myevent.EvtPullDepositGroupMessage),
		new(myevent.EvtPushDepositSystemMessage), new(myevent.EvtPullDepositSystemMessage)}, eventbus.Name("deposit"))
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
			case myevent.EvtPushDepositContactMessage:
				go d.handlePushContactEvent(evt)

			case myevent.EvtPushDepositGroupMessage:
				go d.handlePushGroupEvent(evt)

			case myevent.EvtPullDepositContactMessage:
				go d.handlePullContactMessageEvent(evt)

			case myevent.EvtPullDepositGroupMessage:
				go d.handlePullGroupMessageEvent(evt)

			case myevent.EvtPushDepositSystemMessage:
				go d.handlePushSystemEvent(evt)

			case myevent.EvtPullDepositSystemMessage:
				go d.handlePullSystemMessageEvent(evt)

			default:
				log.Warnf("undefined event type: %T", evt)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (d *DepositClientProto) handlePushContactEvent(evt myevent.EvtPushDepositContactMessage) {
	var resultErr error

	defer func() {
		evt.Result <- resultErr
		close(evt.Result)
	}()

	log.Debugf("handle push contact msg event: %s", evt.MsgID)
	err := d.PushContactMessage(evt.DepositAddress, evt.ToPeerID, evt.MsgID, evt.MsgData)
	if err != nil {
		resultErr = fmt.Errorf("push contact msg error: %w", err)
	}
}

func (d *DepositClientProto) handlePushGroupEvent(evt myevent.EvtPushDepositGroupMessage) {
	var resultErr error

	defer func() {
		evt.Result <- resultErr
		close(evt.Result)
	}()

	log.Debugf("handle push group msg event: %s", evt.MsgID)
	err := d.PushGroupMessage(evt.DepositAddress, evt.ToGroupID, evt.MsgID, evt.MsgData)
	if err != nil {
		resultErr = fmt.Errorf("push group msg error: %w", err)
	}
}

func (d *DepositClientProto) handlePushSystemEvent(evt myevent.EvtPushDepositSystemMessage) {
	var resultErr error

	defer func() {
		evt.Result <- resultErr
		close(evt.Result)
	}()

	log.Debugf("handle push system msg event: %s", evt.MsgID)
	err := d.PushSystemMessage(evt.DepositAddress, evt.ToPeerID, evt.MsgID, evt.MsgData)
	if err != nil {
		resultErr = fmt.Errorf("push contact msg error: %w", err)
	}
}

func (d *DepositClientProto) handlePullContactMessageEvent(evt myevent.EvtPullDepositContactMessage) {
	log.Debugf("receive pull offline event")

	hostID := d.host.ID()

	lastDepositID, err := d.datastore.GetContactLastID(hostID)
	if err != nil {
		log.Errorf("data contact last id error: %v", err)
		return
	}

	depositAddr := evt.DepositAddress
	ctx := context.Background()
	stream, err := d.host.NewStream(network.WithUseTransient(ctx, ""), depositAddr, PULL_CONTACT_MSG_ID)
	if err != nil {
		log.Errorf("new stream to deposit error: %v", err)
		return
	}
	defer stream.Close()

	pr := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeMessage)
	pw := pbio.NewDelimitedWriter(stream)

	if err = pw.WriteMsg(&pb.DepositContactMessagePull{StartId: lastDepositID}); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	var dmsg pb.DepositContactMessage
	for {
		dmsg.Reset()
		if err = pr.ReadMsg(&dmsg); err != nil {
			log.Errorf("receive deposit msg error: %v", err)
			stream.Reset()
			return
		}

		if evt.MessageHandler != nil {
			if err = evt.MessageHandler(context.Background(), peer.ID(dmsg.FromPeerId), dmsg.MsgId, dmsg.MsgData); err != nil {
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

func (d *DepositClientProto) handlePullGroupMessageEvent(evt myevent.EvtPullDepositGroupMessage) {
	log.Debugf("receive pull offline event")

	groupID := evt.GroupID
	depositAddr := evt.DepositAddress

	lastDepositID, err := d.datastore.GetGroupLastID(evt.GroupID)
	if err != nil {
		log.Errorf("data contact last id error: %v", err)
		return
	}

	ctx := context.Background()
	stream, err := d.host.NewStream(network.WithUseTransient(ctx, ""), depositAddr, PULL_GROUP_MSG_ID)
	if err != nil {
		log.Errorf("new stream to deposit error: %v", err)
		return
	}
	defer stream.Close()

	pr := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeMessage)
	pw := pbio.NewDelimitedWriter(stream)

	if err = pw.WriteMsg(&pb.DepositGroupMessagePull{GroupId: groupID, StartId: lastDepositID}); err != nil {
		log.Errorf("pbio write msg error: %v", err)
		stream.Reset()
		return
	}

	var msg pb.DepositGroupMessage
	for {
		msg.Reset()
		if err = pr.ReadMsg(&msg); err != nil {
			log.Errorf("pbio read msg error: %v", err)
			stream.Reset()
			return
		}

		if evt.MessageHandler != nil {
			if err = evt.MessageHandler(ctx, groupID, msg.MsgId, msg.MsgData); err != nil {
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

func (d *DepositClientProto) handlePullSystemMessageEvent(evt myevent.EvtPullDepositSystemMessage) {
	log.Debugf("receive pull offline event")

	hostID := d.host.ID()

	lastDepositID, err := d.datastore.GetSystemLastID(hostID)
	if err != nil {
		log.Errorf("data contact last id error: %v", err)
		return
	}

	depositAddr := evt.DepositAddress
	ctx := context.Background()
	stream, err := d.host.NewStream(network.WithUseTransient(ctx, ""), depositAddr, PULL_SYSTEM_MSG_ID)
	if err != nil {
		log.Errorf("new stream to deposit error: %v", err)
		return
	}
	defer stream.Close()

	pr := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	pw := pbio.NewDelimitedWriter(stream)

	if err = pw.WriteMsg(&pb.DepositSystemMessagePull{StartId: lastDepositID}); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	var dmsg pb.DepositSystemMessage
	for {
		dmsg.Reset()
		if err = pr.ReadMsg(&dmsg); err != nil {
			log.Errorf("receive deposit msg error: %v", err)
			stream.Reset()
			return
		}

		if evt.MessageHandler != nil {
			if err = evt.MessageHandler(context.Background(), peer.ID(dmsg.FromPeerId), dmsg.MsgId, dmsg.MsgData); err != nil {
				log.Errorf("handle deposit msg error: %v", err)
				stream.Reset()
				return
			}
		}

		if err = d.datastore.SetSystemLastID(hostID, dmsg.Id); err != nil {
			log.Errorf("set system last id error: %v", err)
			stream.Reset()
			return
		}
	}
}

func (d *DepositClientProto) PushContactMessage(depositAddr peer.ID, toPeerID peer.ID, msgID string, msgData []byte) error {
	hostID := d.host.ID()

	log.Debugf("get deposit service peer: %s", depositAddr.String())

	ctx := context.Background()
	stream, err := d.host.NewStream(network.WithUseTransient(ctx, ""), depositAddr, PUSH_CONTACT_MSG_ID)
	if err != nil {
		return myerror.WrapStreamError("new stream to deposit peer error", err)
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	wt := pbio.NewDelimitedWriter(stream)

	if err = wt.WriteMsg(&pb.DepositContactMessage{
		FromPeerId:  []byte(hostID),
		ToPeerId:    []byte(toPeerID),
		MsgId:       msgID,
		MsgData:     msgData,
		DepositTime: time.Now().Unix(),
	}); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("write msg error", err)
	}

	var msgAck pb.DepositMessageAck
	if err = rd.ReadMsg(&msgAck); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("read ack msg error", err)

	} else if msgAck.MsgId != msgID {
		stream.Reset()
		return myerror.WrapStreamError("msg ack id error", nil)
	}

	return nil
}

func (d *DepositClientProto) PushGroupMessage(depositAddr peer.ID, groupID string, msgID string, msgData []byte) error {
	hostID := d.host.ID()

	log.Debugf("get deposit service peer: %s", depositAddr.String())

	ctx := context.Background()
	stream, err := d.host.NewStream(network.WithUseTransient(ctx, ""), depositAddr, PUSH_GROUP_MSG_ID)
	if err != nil {
		return myerror.WrapStreamError("new stream to deposit peer error", err)
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	pw := pbio.NewDelimitedWriter(stream)

	dmsg := pb.DepositGroupMessage{
		FromPeerId:  []byte(hostID),
		GroupId:     groupID,
		MsgId:       msgID,
		MsgData:     msgData,
		DepositTime: time.Now().Unix(),
	}

	if err = pw.WriteMsg(&dmsg); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("write msg error", err)
	}

	var msgAck pb.DepositMessageAck
	if err = rd.ReadMsg(&msgAck); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("pbio read ack msg error: %w", err)

	} else if msgAck.MsgId != msgID {
		stream.Reset()
		return myerror.WrapStreamError("msg ack id error", nil)
	}

	return nil
}

func (d *DepositClientProto) PushSystemMessage(depositAddr peer.ID, toPeerID peer.ID, msgID string, msgData []byte) error {
	hostID := d.host.ID()

	log.Debugf("get deposit service peer: %s", depositAddr.String())

	ctx := context.Background()
	stream, err := d.host.NewStream(network.WithUseTransient(ctx, ""), depositAddr, PUSH_SYSTEM_MSG_ID)
	if err != nil {
		return myerror.WrapStreamError("new stream to deposit peer error", err)
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	wt := pbio.NewDelimitedWriter(stream)

	if err = wt.WriteMsg(&pb.DepositSystemMessage{
		FromPeerId:  []byte(hostID),
		ToPeerId:    []byte(toPeerID),
		MsgId:       msgID,
		MsgData:     msgData,
		DepositTime: time.Now().Unix(),
	}); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("write msg error", err)
	}

	var msgAck pb.DepositMessageAck
	if err = rd.ReadMsg(&msgAck); err != nil {
		stream.Reset()
		return myerror.WrapStreamError("read ack msg error", err)

	} else if msgAck.MsgId != msgID {
		stream.Reset()
		return myerror.WrapStreamError("msg ack id error", nil)
	}

	return nil
}
