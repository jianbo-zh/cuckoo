package peer

import (
	"context"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/deposit/protocol/peer/ds"
	"github.com/jianbo-zh/dchat/service/deposit/protocol/peer/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-msgio/pbio"
)

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/offline_msg.proto=./pb pb/offline_msg.proto

var log = logging.Logger("deposit")

var StreamTimeout = 1 * time.Minute

const (
	PUSH_ID = "/dchat/deposit/peer/push/1.0.0" // 寄存服务
	PULL_ID = "/dchat/deposit/peer/pull/1.0.0" // 获取服务

	ServiceName  = "deposit.peer"
	maxMsgSize   = 4 * 1024  // 4K
	msgCacheDays = 3 * 86400 // 3Day
	msgCacheNums = 5000
	rendezvous   = "/dchat/deposit/peer"
)

type PeerDepositService struct {
	host host.Host

	datastore ds.DepositMessageIface
	discv     *drouting.RoutingDiscovery
}

func NewPeerDepositService(h host.Host, rdiscvry *drouting.RoutingDiscovery, ids ipfsds.Batching) (*PeerDepositService, error) {
	gsvc := &PeerDepositService{
		host:      h,
		datastore: ds.DepositPeerWrap(ids),
		discv:     rdiscvry,
	}

	h.SetStreamHandler(PUSH_ID, gsvc.PushHandler)
	h.SetStreamHandler(PULL_ID, gsvc.PullHandler)

	ctx := context.Background()
	dutil.Advertise(ctx, gsvc.discv, rendezvous)

	peerChan, err := gsvc.discv.FindPeers(ctx, rendezvous, discovery.Limit(100))
	if err != nil {
		return nil, err
	}

	go gsvc.findPeers(ctx, peerChan)

	return gsvc, nil
}

func (pds *PeerDepositService) PushHandler(stream network.Stream) {
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

	var dmsg pb.OfflineMessage
	if err = rb.ReadMsg(&dmsg); err != nil {
		stream.Reset()
		log.Errorf("read deposit peer message error: %v", err)
		return
	}

	if err = pds.datastore.SaveDepositMessage(&dmsg); err != nil {
		stream.Reset()
		log.Errorf("save deposit message error: %v", err)
		return
	}

	log.Debugf("save deposit message success")
}

func (pds *PeerDepositService) PullHandler(stream network.Stream) {

	remotePeerID := stream.Conn().RemotePeer()
	log.Debugf("receive pull request, %s", remotePeerID.String())

	defer stream.Close()

	pr := pbio.NewDelimitedReader(stream, maxMsgSize)
	pw := pbio.NewDelimitedWriter(stream)

	lastID, err := pds.datastore.GetLastAckID(remotePeerID)
	if err != nil {
		log.Errorf("get last ack id error: %v", err)
		return
	}

	offset := 0
	startime := time.Now().Unix() - msgCacheDays
	var ackmsg pb.AckMessage
	for {
		msgs, err := pds.datastore.GetDepositMessages(remotePeerID, offset, msgCacheNums, startime, lastID)
		if err != nil {
			log.Errorf("get deposit message ids error: %v", err)
			return
		}

		if len(msgs) == 0 {
			break
		}

		offset = offset + len(msgs)

		for _, msg := range msgs {
			if err = pw.WriteMsg(msg); err != nil {
				log.Errorf("io write message error: %v", err)
				return
			}

			log.Debugf("send offline msg: %s", msg.MsgId)

			if err = pr.ReadMsg(&ackmsg); err != nil {
				log.Errorf("io read ack message error: %v", err)
				return
			}

			if ackmsg.Id == msg.Id {
				if err = pds.datastore.SetLastAckID(remotePeerID, ackmsg.Id); err != nil {
					log.Errorf("set last ack id error: %v", err)
					return
				}
			}
		}
	}
}

func (pds *PeerDepositService) findPeers(ctx context.Context, peerCh <-chan peer.AddrInfo) {

}
