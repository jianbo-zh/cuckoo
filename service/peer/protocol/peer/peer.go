package peer

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/service/peer/protocol/peer/ds"
	"github.com/jianbo-zh/dchat/service/peer/protocol/peer/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

var log = logging.Logger("message")

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/peer.proto=./pb pb/peer.proto

type PeerPeerSvc struct {
	host host.Host
	data ds.PeerIface

	emitters struct {
		evtSyncPeers event.Emitter
	}
}

func NewPeerPeerSvc(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus) (*PeerPeerSvc, error) {
	var err error
	peersvc := PeerPeerSvc{
		host: lhost,
		data: ds.Wrap(ids),
	}

	if peersvc.emitters.evtSyncPeers, err = eventBus.Emitter(&gevent.EvtSyncPeers{}); err != nil {
		return nil, fmt.Errorf("set sync peers emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(gevent.EvtHostBootComplete)})
	if err != nil {
		return nil, fmt.Errorf("subscribe boot complete error: %v", err)

	} else {
		go peersvc.handleSubscribe(context.Background(), sub)
	}

	return &peersvc, nil
}

func (p *PeerPeerSvc) AddPeer(peerID peer.ID, peerInfo PeerInfo) error {
	return p.data.SavePeer(context.Background(), &pb.PeerInfo{
		PeerId:   []byte(peerInfo.PeerID),
		Nickname: peerInfo.Nickname,
		AddTs:    time.Now().Unix(),
		AccessTs: time.Now().Unix(),
	})
}

func (p *PeerPeerSvc) handleSubscribe(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}
			switch ev := e.(type) {
			case gevent.EvtHostBootComplete:
				if !ev.IsSucc {
					log.Warnf("host boot complete but not succ")
					return
				}

				if peerIDs, err := p.data.GetPeerIDs(ctx); err != nil {
					log.Warnf("get peer ids error: %v", err)
					return

				} else if len(peerIDs) > 0 {
					p.emitters.evtSyncPeers.Emit(gevent.EvtSyncPeers{
						PeerIDs: peerIDs,
					})
				}

			default:
				log.Warnf("undefined event type: %T", ev)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (p *PeerPeerSvc) GetPeer(peerID peer.ID) (*PeerInfo, error) {
	peeri, err := p.data.GetPeer(context.Background(), peerID)
	if err != nil {
		return nil, err
	}

	return &PeerInfo{
		PeerID:   peer.ID(peeri.PeerId),
		Nickname: peeri.Nickname,
		AddTs:    peeri.AddTs,
		AccessTs: peeri.AccessTs,
	}, nil
}

func (p *PeerPeerSvc) DeletePeer(peerID peer.ID) error {
	return p.data.DeletePeer(context.Background(), peerID)
}

func (p *PeerPeerSvc) GetPeers() ([]*PeerInfo, error) {
	peeris, err := p.data.GetPeers(context.Background())
	if err != nil {
		return nil, err
	}

	var peers []*PeerInfo
	for _, peeri := range peeris {
		peers = append(peers, &PeerInfo{
			PeerID:   peer.ID(peeri.PeerId),
			Nickname: peeri.Nickname,
			AddTs:    peeri.AddTs,
			AccessTs: peeri.AccessTs,
		})
	}

	return peers, nil
}
