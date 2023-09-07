package depositsvc

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/depositsvc/protocol/peer"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

func Setup(lhost host.Host, rdiscvry *drouting.RoutingDiscovery, ids ipfsds.Batching, ebus event.Bus, options ...Option) (*DepositService, error) {

	depositsvc = &DepositService{}

	if err := depositsvc.Apply(options...); err != nil {
		return nil, err
	}

	dcli, err := peer.NewPeerDepositClient(lhost, rdiscvry, ids, ebus)
	if err != nil {
		return nil, err
	}
	depositsvc.client = dcli

	if depositsvc.isEnableService {
		dsvc, err := peer.NewPeerDepositService(lhost, rdiscvry, ids)
		if err != nil {
			return nil, err
		}
		depositsvc.service = dsvc
	}

	return depositsvc, nil
}
