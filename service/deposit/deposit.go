package deposit

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/host"
	"github.com/jianbo-zh/dchat/service/deposit/datastore"
	"github.com/jianbo-zh/dchat/service/deposit/protocol/peer"
	"github.com/libp2p/go-libp2p/core/event"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var depositsvc *DepositService

type DepositService struct {
	datastore datastore.DepositIface

	service *peer.PeerDepositService

	client *peer.PeerDepositClient
}

func Init(h host.Host, rdiscvry *drouting.RoutingDiscovery, ids ipfsds.Batching, ev event.Bus) (*DepositService, error) {
	depositDs := datastore.DepositWrap(ids)

	dsvc := peer.NewPeerDepositService(h, rdiscvry, ids)
	if err := dsvc.Run(); err != nil {
		return nil, err
	}

	depositsvc = &DepositService{
		datastore: depositDs,
		service:   dsvc,
		client:    peer.NewPeerDepositClient(h, rdiscvry, ids, ev),
	}

	return depositsvc, nil
}
