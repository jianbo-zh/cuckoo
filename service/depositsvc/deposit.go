package depositsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	deposit "github.com/jianbo-zh/dchat/service/depositsvc/protocols/depositproto"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
)

var log = logging.Logger("cuckoo/depositsvc")

var _ DepositServiceIface = (*DepositService)(nil)

type DepositService struct {
	cuckooCtx context.Context

	service *deposit.DepositServiceProto
	client  *deposit.DepositClientProto

	host myhost.Host
	ids  ipfsds.Batching
}

func NewDepositService(ctx context.Context, conf config.DepositServiceConfig, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus) (*DepositService, error) {

	var err error
	depositsvc := &DepositService{
		cuckooCtx: ctx,
		host:      lhost,
		ids:       ids,
	}

	depositsvc.client, err = deposit.NewDepositClientProto(ctx, lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("new deposit client proto error: %w", err)
	}

	if conf.EnableDepositService {
		depositsvc.service, err = deposit.NewDepositServiceProto(ctx, lhost, ids)
		if err != nil {
			return nil, fmt.Errorf("new deposit service proto error: %w", err)
		}
	}

	// 监听配置变化
	sub, err := ebus.Subscribe([]any{new(myevent.EvtConfigEnableDepositServiceChange)}, eventbus.Name("deposit"))
	if err != nil {
		return nil, err
	}
	go depositsvc.subscribeHandler(ctx, sub)

	return depositsvc, nil
}

func (d *DepositService) subscribeHandler(ctx context.Context, sub event.Subscription) {

	defer func() {
		log.Error("deposit subscribe exit")
		sub.Close()
	}()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}
			switch evt := e.(type) {
			case myevent.EvtConfigEnableDepositServiceChange:
				if err := d.handleEnableDepositService(evt.Enable); err != nil {
					log.Errorf("handle enable deposit service error: %w", err)
				}

			default:
				log.Warnf("undefined event type: %T", evt)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (d *DepositService) handleEnableDepositService(isEnable bool) error {
	log.Debugln("handleEnableDepositService: ", isEnable)

	var err error
	if isEnable && d.service == nil {
		// start service
		d.service, err = deposit.NewDepositServiceProto(d.cuckooCtx, d.host, d.ids)
		if err != nil {
			return fmt.Errorf("start deposit service error: %w", err)
		}

		log.Debugln("start deposit service")

	} else if !isEnable && d.service != nil {
		// close service
		d.service.Close()
		d.service = nil

		log.Debugln("close deposit service")
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
