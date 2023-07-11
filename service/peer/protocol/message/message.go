package message

import (
	"context"
	"fmt"
	"time"

	"github.com/jianbo-zh/dchat/service/peer/datastore"
	"github.com/jianbo-zh/dchat/service/peer/protocol/message/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

// 消息服务

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/message.proto=./pb pb/message.proto

var log = logging.Logger("message")

var StreamTimeout = 1 * time.Minute

const (
	ID = "/dchat/peer/msg/1.0.0"

	ServiceName = "peer.message"
	maxMsgSize  = 4 * 1024 // 4K
)

type PeerMessageService struct {
	host      host.Host
	datastore datastore.PeerIface
}

func NewPeerMessageService(h host.Host, ds datastore.PeerIface) *PeerMessageService {
	kas := &PeerMessageService{
		host:      h,
		datastore: ds,
	}
	h.SetStreamHandler(ID, kas.Handler)

	return kas
}

func (p *PeerMessageService) Handler(s network.Stream) {
	fmt.Println("handler....")
	if err := s.Scope().SetService(ServiceName); err != nil {
		log.Errorf("failed to attaching stream to identify service: %v", err)
		s.Reset()
		return
	}
	defer s.Close()

	rd := pbio.NewDelimitedReader(s, maxMsgSize)
	defer rd.Close()

	s.SetDeadline(time.Now().Add(StreamTimeout))

	var msg pb.Message
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		s.Reset()
		return
	}

	s.SetReadDeadline(time.Time{})

	lamptime, err := p.datastore.GetLamportTime(context.Background(), s.Conn().RemotePeer())
	if err != nil {
		log.Errorf("get lamport time error: %v", err)
		return
	}

	err = p.datastore.StoreMessage(context.Background(), s.Conn().RemotePeer(), &msg)
	if err != nil {
		log.Errorf("store message error %v", err)
		s.Reset()
		return
	}

	if msg.LamportTime > lamptime {
		err = p.datastore.SetLamportTime(context.Background(), s.Conn().RemotePeer(), msg.LamportTime)
		if err != nil {
			log.Errorf("update lamport time error: %v", err)
			s.Reset()
			return
		}
	}

	fmt.Printf("%s", msg.GetPayload())
}

func (p *PeerMessageService) SendTextMessage(ctx context.Context, peerID peer.ID, msg string) error {
	stream, err := p.host.NewStream(ctx, peerID, ID)
	if err != nil {
		return err
	}

	msgBytes := []byte(msg)

	signature, err := p.host.Peerstore().PrivKey(p.host.ID()).Sign(msgBytes)
	if err != nil {
		stream.Reset()
		return err
	}

	pw := pbio.NewDelimitedWriter(stream)

	lamportTime, err := p.datastore.TickLamportTime(ctx, peerID)
	if err != nil {
		stream.Reset()
		return err
	}

	pmsg := pb.Message{
		Id:            fmt.Sprintf("%d_%s", lamportTime, p.host.ID().String()),
		Type:          pb.Message_TEXT,
		Payload:       msgBytes,
		SenderId:      p.host.ID().String(),
		ReceiverId:    peerID.String(),
		LamportTime:   lamportTime,
		SendTimestamp: int32(time.Now().Unix()),
		Signature:     signature,
	}

	err = p.datastore.StoreMessage(ctx, peerID, &pmsg)
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

func (p *PeerMessageService) GetMessages(ctx context.Context, peerID peer.ID) ([]*pb.Message, error) {
	return p.datastore.GetMessages(ctx, peerID)
}

func (p *PeerMessageService) SendGroupInviteMessage(ctx context.Context, peerID peer.ID, groupID string) error {
	stream, err := p.host.NewStream(ctx, peerID, ID)
	if err != nil {
		return err
	}

	pw := pbio.NewDelimitedWriter(stream)

	lamportTime, err := p.datastore.TickLamportTime(ctx, peerID)
	if err != nil {
		stream.Reset()
		return err
	}

	pmsg := pb.Message{
		Id:            fmt.Sprintf("%d_%s", lamportTime, p.host.ID().String()),
		Type:          pb.Message_INVITE,
		Payload:       []byte(groupID),
		SenderId:      p.host.ID().String(),
		ReceiverId:    peerID.String(),
		LamportTime:   lamportTime,
		SendTimestamp: int32(time.Now().Unix()),
		Signature:     []byte{},
	}

	err = p.datastore.StoreMessage(ctx, peerID, &pmsg)
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
