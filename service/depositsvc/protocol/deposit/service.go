package deposit

import (
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/protocol"
	"github.com/jianbo-zh/dchat/service/depositsvc/protocol/deposit/ds"
	"github.com/jianbo-zh/dchat/service/depositsvc/protocol/deposit/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-msgio/pbio"
)

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/offline_msg.proto=./pb pb/offline_msg.proto

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
	host host.Host

	datastore ds.DepositMessageIface
}

func NewDepositServiceProto(h host.Host, ids ipfsds.Batching) (*DepositServiceProto, error) {
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

func (d *DepositServiceProto) SaveContactMessageHandler(stream network.Stream) {
	log.Debugf("handle deposit message start")

	err := stream.SetReadDeadline(time.Now().Add(StreamTimeout))
	if err != nil {
		stream.Reset()
		log.Errorf("set read deadline error: %v", err)
		return
	}

	log.Debugf("handle deposit message start")

	rb := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rb.Close()

	var dmsg pb.ContactMessage
	if err = rb.ReadMsg(&dmsg); err != nil {
		stream.Reset()
		log.Errorf("read deposit peer message error: %v", err)
		return
	}

	if err = d.datastore.SaveContactMessage(&dmsg); err != nil {
		stream.Reset()
		log.Errorf("save deposit message error: %v", err)
		return
	}

	log.Debugf("save deposit message success")
}

func (d *DepositServiceProto) SaveGroupMessageHandler(stream network.Stream) {
	log.Debugf("handle deposit message start")

	err := stream.SetReadDeadline(time.Now().Add(StreamTimeout))
	if err != nil {
		stream.Reset()
		log.Errorf("set read deadline error: %v", err)
		return
	}

	log.Debugf("handle deposit message start")

	rb := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rb.Close()

	var dmsg pb.GroupMessage
	if err = rb.ReadMsg(&dmsg); err != nil {
		stream.Reset()
		log.Errorf("read deposit peer message error: %v", err)
		return
	}

	if err = d.datastore.SaveGroupMessage(&dmsg); err != nil {
		stream.Reset()
		log.Errorf("save deposit message error: %v", err)
		return
	}

	log.Debugf("save deposit message success")
}

func (d *DepositServiceProto) GetContactMessageHandler(stream network.Stream) {

	defer stream.Close()

	remotePeerID := stream.Conn().RemotePeer()

	pr := pbio.NewDelimitedReader(stream, maxMsgSize)
	pw := pbio.NewDelimitedWriter(stream)

	var msg pb.ContactMessagePull
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

			startID = msg.Id
		}
	}
}

func (d *DepositServiceProto) GetGroupMessageHandler(stream network.Stream) {

	defer stream.Close()

	pr := pbio.NewDelimitedReader(stream, maxMsgSize)
	pw := pbio.NewDelimitedWriter(stream)

	var msg pb.GroupMessagePull
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
