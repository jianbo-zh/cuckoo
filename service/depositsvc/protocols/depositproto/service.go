package deposit

import (
	"context"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/myprotocol"
	"github.com/jianbo-zh/dchat/internal/mytype"
	ds "github.com/jianbo-zh/dchat/service/depositsvc/datastore/ds/depositds"
	pb "github.com/jianbo-zh/dchat/service/depositsvc/protobuf/pb/depositpb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-msgio/pbio"
)

var log = logging.Logger("cuckoo/depositproto")

var StreamTimeout = 1 * time.Minute

const (
	PUSH_CONTACT_MSG_ID = myprotocol.DepositSaveContactMsgID_v100
	PULL_CONTACT_MSG_ID = myprotocol.DepositGetContactMsgID_v100

	PUSH_GROUP_MSG_ID = myprotocol.DepositSaveGroupMsgID_v100
	PULL_GROUP_MSG_ID = myprotocol.DepositGetGroupMsgID_v100

	PUSH_SYSTEM_MSG_ID = myprotocol.DepositSaveSystemMsgID_v100
	PULL_SYSTEM_MSG_ID = myprotocol.DepositGetSystemMsgID_v100

	ServiceName  = "deposit.peer"
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

	h.SetStreamHandler(PUSH_CONTACT_MSG_ID, gsvc.SaveContactMessageHandler)
	h.SetStreamHandler(PULL_CONTACT_MSG_ID, gsvc.GetContactMessageHandler)

	h.SetStreamHandler(PUSH_GROUP_MSG_ID, gsvc.SaveGroupMessageHandler)
	h.SetStreamHandler(PULL_GROUP_MSG_ID, gsvc.GetGroupMessageHandler)

	h.SetStreamHandler(PUSH_SYSTEM_MSG_ID, gsvc.SaveSystemMessageHandler)
	h.SetStreamHandler(PULL_SYSTEM_MSG_ID, gsvc.GetSystemMessageHandler)

	return gsvc, nil
}

func (d *DepositServiceProto) Close() {
	d.host.RemoveStreamHandler(PUSH_CONTACT_MSG_ID)
	d.host.RemoveStreamHandler(PULL_CONTACT_MSG_ID)
	d.host.RemoveStreamHandler(PUSH_GROUP_MSG_ID)
	d.host.RemoveStreamHandler(PULL_GROUP_MSG_ID)
	d.host.RemoveStreamHandler(PUSH_SYSTEM_MSG_ID)
	d.host.RemoveStreamHandler(PULL_SYSTEM_MSG_ID)
}

// SaveContactMessageHandler 保存联系人消息
func (d *DepositServiceProto) SaveContactMessageHandler(stream network.Stream) {
	log.Debugln("SaveContactMessageHandler: ")

	err := stream.SetReadDeadline(time.Now().Add(StreamTimeout))
	if err != nil {
		stream.Reset()
		log.Errorf("set read deadline error: %v", err)
		return
	}
	defer stream.Close()

	rb := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeMessage)
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
	log.Debugln("datastore.SaveContactMessage ", msg.Id)

	if err = wt.WriteMsg(&pb.DepositMessageAck{MsgId: msg.MsgId}); err != nil {
		stream.Reset()
		log.Errorf("pbio write ack msg error: %w", err)
		return
	}
}

// SaveGroupMessageHandler 保存群组消息
func (d *DepositServiceProto) SaveGroupMessageHandler(stream network.Stream) {
	log.Debugf("save group message start")

	err := stream.SetReadDeadline(time.Now().Add(StreamTimeout))
	if err != nil {
		stream.Reset()
		log.Errorf("set read deadline error: %v", err)
		return
	}
	defer stream.Close()

	rb := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeMessage)
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

// SaveSystemMessageHandler 保存系统消息
func (d *DepositServiceProto) SaveSystemMessageHandler(stream network.Stream) {
	log.Debugln("SaveSystemMessageHandler: ")

	err := stream.SetReadDeadline(time.Now().Add(StreamTimeout))
	if err != nil {
		stream.Reset()
		log.Errorf("set read deadline error: %v", err)
		return
	}
	defer stream.Close()

	rb := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	wt := pbio.NewDelimitedWriter(stream)

	log.Debugln("pbio read deposit system msg")
	var msg pb.DepositSystemMessage
	if err = rb.ReadMsg(&msg); err != nil {
		stream.Reset()
		log.Errorf("read deposit system message error: %v", err)
		return
	}

	if err = d.datastore.SaveSystemMessage(&msg); err != nil {
		stream.Reset()
		log.Errorf("save deposit message error: %v", err)
		return
	}

	log.Debugln("pbio write deposit system msg ack")
	if err = wt.WriteMsg(&pb.DepositMessageAck{MsgId: msg.MsgId}); err != nil {
		stream.Reset()
		log.Errorf("pbio write ack msg error: %w", err)
		return
	}
}

// GetContactMessageHandler 获取联系人消息
func (d *DepositServiceProto) GetContactMessageHandler(stream network.Stream) {

	log.Debugln("GetContactMessageHandler: ")

	defer stream.Close()

	remotePeerID := stream.Conn().RemotePeer()

	pr := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	pw := pbio.NewDelimitedWriter(stream)

	var msg pb.DepositContactMessagePull
	if err := pr.ReadMsg(&msg); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	log.Debugln("recevie pull deposit contact msg request: ", msg.String())

	startID := msg.StartId

	for {
		msgs, err := d.datastore.GetContactMessages(remotePeerID, startID, 50)
		if err != nil {
			log.Errorf("get deposit contact msg ids error: %v", err)
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

			log.Debugln("send deposit contact msg: ", msg.Id)
			startID = msg.Id
		}
	}
}

// GetGroupMessageHandler 获取群组消息
func (d *DepositServiceProto) GetGroupMessageHandler(stream network.Stream) {
	log.Debugln("GetGroupMessageHandler: ")

	defer stream.Close()

	pr := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
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

// GetSystemMessageHandler 获取系统消息
func (d *DepositServiceProto) GetSystemMessageHandler(stream network.Stream) {
	log.Debugln("deposit GetSystemMessageHandler")

	defer stream.Close()

	remotePeerID := stream.Conn().RemotePeer()

	pr := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	pw := pbio.NewDelimitedWriter(stream)

	log.Debugln("pbio read pull system request")
	var msg pb.DepositSystemMessagePull
	if err := pr.ReadMsg(&msg); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	startID := msg.StartId

	for {
		msgs, err := d.datastore.GetSystemMessages(remotePeerID, startID, 50)
		if err != nil {
			log.Errorf("get deposit message ids error: %v", err)
			stream.Reset()
			return
		}

		if len(msgs) == 0 {
			break
		}

		for _, msg := range msgs {
			log.Debugln("pbio write system msg ", msg.MsgId)
			if err = pw.WriteMsg(msg); err != nil {
				log.Errorf("io write message error: %v", err)
				stream.Reset()
				return
			}

			startID = msg.Id
		}
	}
}
