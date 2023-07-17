package peer

// 发布及获取离线消息

import (
	"context"
	"fmt"
	"sync"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/service/deposit/protocol/peer/ds"
	"github.com/jianbo-zh/dchat/service/deposit/protocol/peer/pb"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/libp2p/go-msgio/pbio"
	"github.com/whyrusleeping/go-keyspace"
)

type PeerDepositClient struct {
	host host.Host

	datastore ds.DepositMessageIface
	discv     *drouting.RoutingDiscovery

	depositPeerMutex sync.Mutex
	depositPeers     map[string]struct{}
}

func NewPeerDepositClient(h host.Host, rdiscvry *drouting.RoutingDiscovery, ids ipfsds.Batching, eventBus event.Bus) *PeerDepositClient {
	gcli := &PeerDepositClient{
		host:      h,
		datastore: ds.DepositPeerWrap(ids),
		discv:     rdiscvry,
	}

	// 订阅push、pull
	sub, err := eventBus.Subscribe([]any{new(gevent.PushOfflineMessageEvt), new(gevent.PullOfflineMessageEvt)}, eventbus.Name("deposit"))
	if err != nil {
		log.Warnf("eventbus subscribe deposit event error: %v", err)

	} else {
		gcli.handleSubscribe(context.Background(), sub)
	}

	return gcli
}

func (pcli *PeerDepositClient) handleSubscribe(ctx context.Context, sub event.Subscription) {

	// 处理寄存Peer发现
	peerChan, err := pcli.discv.FindPeers(ctx, rendezvous, discovery.Limit(10))
	if err != nil {
		log.Errorf("peer client find peers error: %v", err)
		return
	}

	go pcli.findPeers(ctx, peerChan)

	// 处理
	defer func() {
		log.Error("peer client subscrib exit")
		sub.Close()
	}()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}
			switch ev := e.(type) {
			case gevent.PushOfflineMessageEvt:
				pcli.handPushEvent(ev)
			case gevent.PullOfflineMessageEvt:
				pcli.handPullEvent(ev)
			default:
				log.Warnf("undefined event type: %T", ev)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (pcli *PeerDepositClient) handPushEvent(ev gevent.PushOfflineMessageEvt) error {
	return pcli.Push(ev.ToPeerID, ev.MsgID, ev.MsgData)
}

func (pcli *PeerDepositClient) handPullEvent(ev gevent.PullOfflineMessageEvt) error {

	ctx := context.Background()
	peerID := pcli.host.ID().String()

	depositPeerID, err := pcli.getDepositPeer(context.Background(), peerID)
	if err != nil {
		return fmt.Errorf("get deposit peer error: %v", err)
	}

	stream, err := pcli.host.NewStream(ctx, depositPeerID, PULL_ID)
	if err != nil {
		return err
	}
	defer stream.Close()

	pr := pbio.NewDelimitedReader(stream, maxMsgSize)
	pw := pbio.NewDelimitedWriter(stream)

	var dmsg pb.DepositMessage
	for {
		dmsg.Reset()
		if err = pr.ReadMsg(&dmsg); err != nil {
			return err
		}

		exists, err := ev.HasMessage(dmsg.FromPeerId, dmsg.MsgId)
		if err != nil {
			return err
		}

		if !exists {
			if err = ev.SaveMessage(dmsg.FromPeerId, dmsg.MsgId, dmsg.MsgData); err != nil {
				return err
			}
		}

		if err = pw.WriteMsg(&pb.AckMessage{Id: dmsg.Id}); err != nil {
			return err
		}
	}
}

func (pcli *PeerDepositClient) Push(toPeerID string, msgID string, msgData []byte) error {
	fromPeerID := pcli.host.ID().String()

	depositPeerID, err := pcli.getDepositPeer(context.Background(), toPeerID)
	if err != nil {
		return fmt.Errorf("get deposit peer error: %v", err)
	}

	stream, err := pcli.host.NewStream(network.WithUseTransient(context.Background(), ""), depositPeerID, PUSH_ID)
	if err != nil {
		return fmt.Errorf("new stream to deposit peer error: %v", err)
	}

	pw := pbio.NewDelimitedWriter(stream)
	defer pw.Close()

	dmsg := pb.DepositMessage{
		FromPeerId:  fromPeerID,
		ToPeerId:    toPeerID,
		MsgId:       msgID,
		MsgData:     msgData,
		DepositTime: time.Now().Unix(),
	}

	if err = pw.WriteMsg(&dmsg); err != nil {
		return fmt.Errorf("write msg error: %v", err)
	}

	return nil
}

func (pcli *PeerDepositClient) findPeers(ctx context.Context, peerCh <-chan peer.AddrInfo) {
	if pcli.depositPeers == nil {
		pcli.depositPeerMutex.Lock()
		pcli.depositPeers = make(map[string]struct{})
		pcli.depositPeerMutex.Unlock()
	}

	for peer := range peerCh {
		pcli.depositPeerMutex.Lock()
		pcli.depositPeers[peer.ID.String()] = struct{}{}
		pcli.depositPeerMutex.Unlock()
	}
}

func (pcli *PeerDepositClient) getDepositPeer(ctx context.Context, toPeerID string) (peer.ID, error) {
	if len(pcli.depositPeers) == 0 {
		return "", fmt.Errorf("not found deposit service peer")
	}

	peerID0, err := peer.Decode(toPeerID)
	if err != nil {
		return "", fmt.Errorf("peer decode peerid error: %v", err)
	}
	peerKey0 := keyspace.XORKeySpace.Key([]byte(peerID0))

	var findPeer peer.ID
	var minDistance int64

	for peerIDs := range pcli.depositPeers {
		peerID, _ := peer.Decode(peerIDs)
		peerKey := keyspace.XORKeySpace.Key([]byte(peerID))
		distance := peerKey0.Distance(peerKey).Int64()

		if minDistance == 0 || distance < minDistance {
			findPeer = peerID
			minDistance = distance
		}
	}

	return findPeer, nil
}
