package peer

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/peer/protocol/message"
	peerpeer "github.com/jianbo-zh/dchat/service/peer/protocol/peer"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
)

var peersvc *PeerSvc

func Init(lhost host.Host, ids ipfsds.Batching, ebus event.Bus, opts ...Option) (*PeerSvc, error) {

	var err error

	peersvc = &PeerSvc{}

	if err := peersvc.Apply(opts...); err != nil {
		return nil, err
	}

	peersvc.msgSvc, err = message.NewPeerMessageSvc(lhost, ids, ebus)
	if err != nil {
		return nil, err
	}

	peersvc.peerSvc, err = peerpeer.NewPeerPeerSvc(lhost, ids, ebus)
	if err != nil {
		return nil, err
	}

	return peersvc, nil
}
