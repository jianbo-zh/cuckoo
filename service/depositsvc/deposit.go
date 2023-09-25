package depositsvc

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/service/depositsvc/protocol/deposit"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

var depositsvc *DepositService

type DepositService struct {
	isEnableService bool

	service *deposit.DepositServiceProto
	client  *deposit.DepositClientProto
}

func NewDepositService(ctx context.Context, conf config.DepositServiceConfig, lhost host.Host, ids ipfsds.Batching, ebus event.Bus) (*DepositService, error) {

	depositsvc = &DepositService{}

	dcli, err := deposit.NewDepositClientProto(lhost, ids, ebus)
	if err != nil {
		return nil, err
	}
	depositsvc.client = dcli

	if conf.EnableService {
		dsvc, err := deposit.NewDepositServiceProto(lhost, ids)
		if err != nil {
			return nil, err
		}
		depositsvc.service = dsvc
	}

	return depositsvc, nil
}

func (d *DepositService) PushContactMessage(depositPeerID peer.ID, toPeerID peer.ID, msgID string, msgData []byte) error {
	return d.client.PushContactMessage(depositPeerID, toPeerID, msgID, msgData)
}

func (d *DepositService) PushGroupMessage(depositPeerID peer.ID, groupID string, msgID string, msgData []byte) error {
	return d.client.PushGroupMessage(depositPeerID, groupID, msgID, msgData)
}

func (d *DepositService) Close() {}
