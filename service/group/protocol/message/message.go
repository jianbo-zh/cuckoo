package message

import (
	"context"
	"errors"
	"fmt"
	"time"

	ds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/service/group/datastore"
	"github.com/jianbo-zh/dchat/service/group/protocol/message/pb"
	networkpb "github.com/jianbo-zh/dchat/service/group/protocol/network/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

// 群消息相关协议

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/message.proto=./pb pb/message.proto

var log = logging.Logger("message")

var StreamTimeout = 1 * time.Minute

const (
	ID = "/dchat/group/msg/1.0.0"

	ServiceName = "group.message"
	maxMsgSize  = 4 * 1024 // 4K
)

type GroupMessageService struct {
	host host.Host

	datastore datastore.GroupIface

	emitters struct {
		evtForwardGroupMsg event.Emitter
	}
}

func NewGroupMessageService(h host.Host, ds datastore.GroupIface, eventBus event.Bus) *GroupMessageService {
	msgsvc := &GroupMessageService{
		host:      h,
		datastore: ds,
	}

	h.SetStreamHandler(ID, msgsvc.Handler)

	var err error
	if msgsvc.emitters.evtForwardGroupMsg, err = eventBus.Emitter(&gevent.EvtForwardGroupMsg{}); err != nil {
		log.Errorf("set group msg emitter error: %v", err)
	}

	return msgsvc
}

func (msgsvc *GroupMessageService) Handler(s network.Stream) {

	remotePeerID := s.Conn().RemotePeer()

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

	var msg pb.GroupMsg
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		s.Reset()
		return
	}

	s.SetReadDeadline(time.Time{})

	// 检查本地是否存在
	_, err := msgsvc.datastore.GetMessage(context.Background(), datastore.GroupID(msg.GroupId), msg.Id)
	if err != nil && errors.Is(err, ds.ErrNotFound) {
		// 保存消息
		err = msgsvc.datastore.PutMessage(context.Background(), datastore.GroupID(msg.GroupId), &msg)
		if err != nil {
			log.Errorf("save group message error: %v", err)
		}

		// 转发消息
		err = msgsvc.ForwardMessage(context.Background(), msg.GroupId, remotePeerID, &msg)
		if err != nil {
			log.Errorf("emit forward group msg error: %v", err)
		}
		return

	} else if err != nil {
		log.Errorf("get group message error: %v", err)
		return
	}
}

func (msgsvc *GroupMessageService) SendTextMessage(ctx context.Context, groupID string, msg string) error {

	peerID := msgsvc.host.ID().String()
	lamportime, err := msgsvc.datastore.TickMessageLamportTime(context.Background(), datastore.GroupID(groupID))
	if err != nil {
		return err
	}
	msg1 := pb.GroupMsg{
		Id:         fmt.Sprintf("%d_%s", lamportime, peerID),
		PeerId:     peerID,
		Type:       pb.GroupMsg_TEXT,
		Payload:    []byte(msg),
		Timestamp:  time.Now().Unix(),
		Lamportime: lamportime,
	}

	err = msgsvc.SendMessage(ctx, groupID, &msg1)
	if err != nil {
		return err
	}

	return nil
}

// 转发消息
func (msgsvc *GroupMessageService) ForwardMessage(ctx context.Context, groupID string, peerID0 peer.ID, msg *pb.GroupMsg) error {

	peerIDs, err := msgsvc.getConnectPeers(groupID)
	if err != nil {
		return err
	}

	for _, peerID := range peerIDs {
		if peerID.String() != peerID0.String() {
			msgsvc.sendPeerMessage(ctx, groupID, peerID, msg)
		}
	}

	return nil
}

// 发送消息
func (msgsvc *GroupMessageService) SendMessage(ctx context.Context, groupID string, msg *pb.GroupMsg) error {

	peerIDs, err := msgsvc.getConnectPeers(groupID)
	if err != nil {
		return err
	}

	for _, peerID := range peerIDs {
		msgsvc.sendPeerMessage(ctx, groupID, peerID, msg)
	}

	return nil
}

// 发送消息（指定peerID）
func (msgsvc *GroupMessageService) sendPeerMessage(ctx context.Context, groupID string, peerID peer.ID, msg *pb.GroupMsg) error {
	stream, err := msgsvc.host.NewStream(ctx, peerID, ID)
	if err != nil {
		return err
	}

	stream.SetWriteDeadline(time.Now().Add(StreamTimeout))
	wb := pbio.NewDelimitedWriter(stream)
	if err := wb.WriteMsg(msg); err != nil {
		return err
	}

	stream.SetWriteDeadline(time.Time{})

	return nil
}

func (msgsvc *GroupMessageService) getConnectPeers(groupID string) ([]peer.ID, error) {
	peerID := msgsvc.host.ID().String()
	connects, err := msgsvc.datastore.GetGroupConnects(context.Background(), datastore.GroupID(groupID))
	if err != nil {
		return nil, err
	}
	onlinesMap := make(map[string]struct{})
	for _, connect := range connects {
		if connect.State == networkpb.GroupConnect_CONNECT {
			if connect.PeerIdA == peerID {
				onlinesMap[connect.PeerIdB] = struct{}{}

			} else if connect.PeerIdB == peerID {
				onlinesMap[connect.PeerIdA] = struct{}{}
			}
		}
	}

	var peerIDs []peer.ID
	for pid := range onlinesMap {
		peerID, _ := peer.Decode(pid)
		peerIDs = append(peerIDs, peerID)
	}

	return peerIDs, nil
}
