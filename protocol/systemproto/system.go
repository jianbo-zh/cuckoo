package systemproto

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	ds "github.com/jianbo-zh/dchat/datastore/ds/systemds"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/internal/protocol"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/systempb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

var log = logging.Logger("system")

var StreamTimeout = 1 * time.Minute

const (
	ID         = protocol.SystemMessageID_v100
	maxMsgSize = 4 * 1024 // 4K
)

type SystemProto struct {
	host  myhost.Host
	data  ds.SystemIface
	msgCh chan<- *pb.SystemMessage
}

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/system.proto=./pb pb/system.proto

func NewSystemProto(lhost myhost.Host, ids ipfsds.Batching, msgCh chan<- *pb.SystemMessage) (*SystemProto, error) {
	systemProto := SystemProto{
		host:  lhost,
		data:  ds.Wrap(ids),
		msgCh: msgCh,
	}

	lhost.SetStreamHandler(ID, systemProto.Handler)

	return &systemProto, nil
}

func (s *SystemProto) Handler(stream network.Stream) {
	rb := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rb.Close()

	var msg pb.SystemMessage
	if err := rb.ReadMsg(&msg); err != nil {
		log.Errorf("rb.ReadMsg error: %w", err)
		return
	}

	// 交给上级处理
	s.msgCh <- &msg
}

func (s *SystemProto) SaveMessage(ctx context.Context, msg *pb.SystemMessage) error {
	if err := s.data.AddSystemMessage(ctx, msg); err != nil {
		return fmt.Errorf("s.data.AddSystemMessage error: %w", err)
	}
	return nil
}

func (s *SystemProto) GetMessage(ctx context.Context, msgID string) (*pb.SystemMessage, error) {
	msg, err := s.data.GetSystemMessage(ctx, msgID)
	if err != nil {
		return nil, fmt.Errorf("s.data.GetSystemMessage error: %w", err)
	}

	return msg, nil
}

func (s *SystemProto) GetMessageList(ctx context.Context, offset int, limit int) ([]*pb.SystemMessage, error) {
	msgs, err := s.data.GetSystemMessageList(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("data get system message list error: %w", err)
	}

	return msgs, nil
}

func (s *SystemProto) UpdateMessageState(ctx context.Context, msgID string, state string) error {
	err := s.data.UpdateSystemMessageState(ctx, msgID, state)
	if err != nil {
		return fmt.Errorf("update system message state error: %w", err)
	}

	return nil
}

func (s *SystemProto) DeleteSystemMessage(ctx context.Context, msgIDs []string) error {
	return s.data.DeleteSystemMessage(ctx, msgIDs)
}

func (s *SystemProto) SendMessage(ctx context.Context, msg *pb.SystemMessage) error {

	stream, err := s.host.NewStream(network.WithDialPeerTimeout(ctx, mytype.DialTimeout), peer.ID(msg.ToPeerId), ID)
	if err != nil {
		return fmt.Errorf("a.host.NewStream error: %w,%s", err, peer.ID(msg.ToPeerId).String())
	}

	wt := pbio.NewDelimitedWriter(stream)
	defer wt.Close()

	if err = wt.WriteMsg(msg); err != nil {
		return fmt.Errorf("wt.WriteMsg error: %w", err)
	}

	return nil
}
