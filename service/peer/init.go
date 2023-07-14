package peer

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/peer/protocol/message"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
)

var peersvc *PeerSvc

func Init(lhost host.Host, ids ipfsds.Batching, ebus event.Bus, opts ...Option) (*PeerSvc, error) {

	peersvc = &PeerSvc{}

	if err := peersvc.Apply(opts...); err != nil {
		return nil, err
	}

	peersvc.msgSvc = message.NewPeerMessageSvc(lhost, ids, ebus)

	return peersvc, nil
}
