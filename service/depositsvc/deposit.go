package depositsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/depositsvc/protocol/deposit"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

var depositsvc *DepositService

type DepositService struct {
	cuckooCtx context.Context

	service *deposit.DepositServiceProto
	client  *deposit.DepositClientProto

	accountGetter types.AccountGetter
	host          host.Host
	ids           ipfsds.Batching
}

func NewDepositService(ctx context.Context, conf config.DepositServiceConfig, lhost host.Host, ids ipfsds.Batching, ebus event.Bus, accountGetter types.AccountGetter) (*DepositService, error) {

	var err error
	depositsvc = &DepositService{
		cuckooCtx:     ctx,
		accountGetter: accountGetter,
		host:          lhost,
		ids:           ids,
	}

	depositsvc.client, err = deposit.NewDepositClientProto(ctx, lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("new deposit client proto error: %w", err)
	}

	account, err := depositsvc.accountGetter.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account error: %w", err)
	}

	if account.EnableDepositService {
		depositsvc.service, err = deposit.NewDepositServiceProto(ctx, lhost, ids)
		if err != nil {
			return nil, fmt.Errorf("new deposit service proto error: %w", err)
		}
	}

	return depositsvc, nil
}

func (d *DepositService) PushContactMessage(depositPeerID peer.ID, toPeerID peer.ID, msgID string, msgData []byte) error {
	return d.client.PushContactMessage(depositPeerID, toPeerID, msgID, msgData)
}

func (d *DepositService) PushGroupMessage(depositPeerID peer.ID, groupID string, msgID string, msgData []byte) error {
	return d.client.PushGroupMessage(depositPeerID, groupID, msgID, msgData)
}

func (d *DepositService) MaintainDepositService() error {
	account, err := d.accountGetter.GetAccount(context.Background())
	if err != nil {
		return fmt.Errorf("get account error: %w", err)
	}

	if account.EnableDepositService && d.service == nil {
		// start service
		d.service, err = deposit.NewDepositServiceProto(d.cuckooCtx, d.host, d.ids)
		if err != nil {
			return fmt.Errorf("start deposit service error: %w", err)
		}
		fmt.Println("deposit service started")

	} else if !account.EnableDepositService && d.service != nil {
		// close service
		d.service.Close()
		d.service = nil
		fmt.Println("deposit service closed")
	}

	return nil
}

func (d *DepositService) Close() {
	if d.service != nil {
		d.service.Close()
		d.service = nil
	}
	d.client = nil
}
