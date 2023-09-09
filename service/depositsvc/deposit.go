package depositsvc

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/service/depositsvc/protocol/peer"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var depositsvc *DepositService

type DepositService struct {
	isEnableService bool

	service *peer.PeerDepositService
	client  *peer.PeerDepositClient
}

func NewDepositService(conf config.DepositServiceConfig, lhost host.Host, ids ipfsds.Batching, ebus event.Bus, rdiscvry *drouting.RoutingDiscovery) (*DepositService, error) {

	depositsvc = &DepositService{}

	dcli, err := peer.NewPeerDepositClient(lhost, rdiscvry, ids, ebus)
	if err != nil {
		return nil, err
	}
	depositsvc.client = dcli

	if conf.EnableService {
		dsvc, err := peer.NewPeerDepositService(lhost, rdiscvry, ids)
		if err != nil {
			return nil, err
		}
		depositsvc.service = dsvc
	}

	return depositsvc, nil
}

func (d *DepositService) Close() {}
