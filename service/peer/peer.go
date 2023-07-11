package peer

import (
	"context"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/peer/datastore"
	"github.com/jianbo-zh/dchat/service/peer/protocol/message"
	"github.com/jianbo-zh/dchat/service/peer/protocol/message/pb"
	"github.com/jianbo-zh/dchat/service/peer/protocol/msgsync"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

var peersvc *PeerService

type PeerService struct {
	datastore datastore.PeerIface

	pingSvc    *ping.PingService
	msgSvc     *message.PeerMessageService
	msgSyncSvc *msgsync.PeerMsgSyncService
}

func Get() PeerServiceIface {
	if peersvc == nil {
		panic("peer service must init before use")
	}

	return peersvc
}

func Init(h host.Host, ids ipfsds.Batching, ev event.Bus, opts ...Option) (*PeerService, error) {

	peerDs := datastore.PeerWrap(ids)

	peersvc = &PeerService{
		datastore: peerDs,
	}

	if err := peersvc.Apply(opts...); err != nil {
		return nil, err
	}

	// 注册协议
	peersvc.pingSvc = ping.NewPingService(h)
	peersvc.msgSvc = message.NewPeerMessageService(h, peerDs)
	peersvc.msgSyncSvc = msgsync.NewPeerMsgSyncService(h, peerDs)

	return peersvc, nil
}

func (peer *PeerService) Ping(ctx context.Context, p peer.ID) (time.Duration, error) {
	ctx1, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := <-peer.pingSvc.Ping(ctx1, p)
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RTT, nil
}

func (peer *PeerService) SendTextMessage(ctx context.Context, peerID peer.ID, msg string) error {
	return peer.msgSvc.SendTextMessage(ctx, peerID, msg)
}

func (peer *PeerService) GetMessages(ctx context.Context, peerID peer.ID) ([]*pb.Message, error) {
	return peer.msgSvc.GetMessages(ctx, peerID)
}

func (peer *PeerService) SendGroupInviteMessage(ctx context.Context, peerID peer.ID, groupID string) error {
	return peer.msgSvc.SendGroupInviteMessage(ctx, peerID, groupID)
}
