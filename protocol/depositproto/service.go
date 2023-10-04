package deposit

import (
	"context"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	ds "github.com/jianbo-zh/dchat/datastore/ds/depositds"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/protocol"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/depositpb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-msgio/pbio"
)

var log = logging.Logger("deposit")

var StreamTimeout = 1 * time.Minute

const (
	CONTACT_SAVE_ID = protocol.DepositContactSaveID_v100
	CONTACT_GET_ID  = protocol.DepositContactGetID_v100
	GROUP_SAVE_ID   = protocol.DepositGroupSaveID_v100
	GROUP_GET_ID    = protocol.DepositGroupGetID_v100

	ServiceName  = "deposit.peer"
	maxMsgSize   = 4 * 1024  // 4K
	msgCacheDays = 3 * 86400 // 3Day
	msgCacheNums = 5000
)

type DepositServiceProto struct {
	host myhost.Host

	datastore ds.DepositMessageIface
}

func NewDepositServiceProto(ctx context.Context, h myhost.Host, ids ipfsds.Batching) (*DepositServiceProto, error) {
	gsvc := &DepositServiceProto{
		host:      h,
		datastore: ds.DepositPeerWrap(ids),
	}

	h.SetStreamHandler(CONTACT_SAVE_ID, gsvc.SaveContactMessageHandler)
	h.SetStreamHandler(CONTACT_GET_ID, gsvc.GetContactMessageHandler)

	h.SetStreamHandler(GROUP_SAVE_ID, gsvc.SaveGroupMessageHandler)
	h.SetStreamHandler(GROUP_GET_ID, gsvc.GetGroupMessageHandler)

	return gsvc, nil
}

func (d *DepositServiceProto) Close() {
	d.host.RemoveStreamHandler(CONTACT_SAVE_ID)
	d.host.RemoveStreamHandler(CONTACT_GET_ID)
	d.host.RemoveStreamHandler(GROUP_SAVE_ID)
	d.host.RemoveStreamHandler(GROUP_GET_ID)
}

func (d *DepositServiceProto) SaveContactMessageHandler(stream network.Stream) {
	log.Debugln("save contact message start")

	err := stream.SetReadDeadline(time.Now().Add(StreamTimeout))
	if err != nil {
		stream.Reset()
		log.Errorf("set read deadline error: %v", err)
		return
	}
	defer stream.Close()

	rb := pbio.NewDelimitedReader(stream, maxMsgSize)
	wt := pbio.NewDelimitedWriter(stream)

	var msg pb.DepositContactMessage
	if err = rb.ReadMsg(&msg); err != nil {
		stream.Reset()
		log.Errorf("read deposit peer message error: %v", err)
		return
	}

	if err = d.datastore.SaveContactMessage(&msg); err != nil {
		stream.Reset()
		log.Errorf("save deposit message error: %v", err)
		return
	}

	if err = wt.WriteMsg(&pb.DepositMessageAck{MsgId: msg.MsgId}); err != nil {
		stream.Reset()
		log.Errorf("pbio write ack msg error: %w", err)
		return
	}

	log.Debugln("save contact message success")
}

func (d *DepositServiceProto) SaveGroupMessageHandler(stream network.Stream) {
	log.Debugf("save group message start")

	err := stream.SetReadDeadline(time.Now().Add(StreamTimeout))
	if err != nil {
		stream.Reset()
		log.Errorf("set read deadline error: %v", err)
		return
	}
	defer stream.Close()

	rb := pbio.NewDelimitedReader(stream, maxMsgSize)
	wt := pbio.NewDelimitedWriter(stream)

	var msg pb.DepositGroupMessage
	if err = rb.ReadMsg(&msg); err != nil {
		stream.Reset()
		log.Errorf("read deposit peer message error: %v", err)
		return
	}

	if err = d.datastore.SaveGroupMessage(&msg); err != nil {
		stream.Reset()
		log.Errorf("save deposit message error: %v", err)
		return
	}

	if err = wt.WriteMsg(&pb.DepositMessageAck{MsgId: msg.MsgId}); err != nil {
		stream.Reset()
		log.Errorf("pbio write ack msg error: %w", err)
		return
	}

	log.Debugf("save group message success")
}

func (d *DepositServiceProto) GetContactMessageHandler(stream network.Stream) {

	log.Debugln("get contact message start")

	defer stream.Close()

	remotePeerID := stream.Conn().RemotePeer()

	pr := pbio.NewDelimitedReader(stream, maxMsgSize)
	pw := pbio.NewDelimitedWriter(stream)

	var msg pb.DepositContactMessagePull
	if err := pr.ReadMsg(&msg); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	startID := msg.StartId

	for {
		msgs, err := d.datastore.GetContactMessages(remotePeerID, startID, 50)
		if err != nil {
			log.Errorf("get deposit message ids error: %v", err)
			stream.Reset()
			return
		}

		if len(msgs) == 0 {
			break
		}

		for _, msg := range msgs {
			if err = pw.WriteMsg(msg); err != nil {
				log.Errorf("io write message error: %v", err)
				stream.Reset()
				return
			}

			log.Debugln("send msg: ", msg.String())

			startID = msg.Id
		}
	}

	log.Debugln("get contact message success")
}

func (d *DepositServiceProto) GetGroupMessageHandler(stream network.Stream) {

	defer stream.Close()

	pr := pbio.NewDelimitedReader(stream, maxMsgSize)
	pw := pbio.NewDelimitedWriter(stream)

	var msg pb.DepositGroupMessagePull
	if err := pr.ReadMsg(&msg); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	groupID := msg.GroupId
	startID := msg.StartId

	for {
		msgs, err := d.datastore.GetGroupMessages(groupID, startID, 50)
		if err != nil {
			log.Errorf("get deposit message ids error: %v", err)
			stream.Reset()
			return
		}

		if len(msgs) == 0 {
			break
		}

		for _, msg := range msgs {
			if err = pw.WriteMsg(msg); err != nil {
				log.Errorf("io write message error: %v", err)
				stream.Reset()
				return
			}

			startID = msg.Id
		}
	}
}
