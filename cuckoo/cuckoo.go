package cuckoo

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/jianbo-zh/dchat/cuckoo/config"
	myevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/datastore"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/contactsvc"
	"github.com/jianbo-zh/dchat/service/depositsvc"
	"github.com/jianbo-zh/dchat/service/groupsvc"
	"github.com/jianbo-zh/dchat/service/systemsvc"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
)

type CuckooGetter interface {
	GetCuckoo() (*Cuckoo, error)
}

type Cuckoo struct {
	host host.Host
	ddht *dual.DHT
	ebus event.Bus

	accountSvc accountsvc.AccountServiceIface
	contactSvc contactsvc.ContactServiceIface
	groupSvc   groupsvc.GroupServiceIface
	depositSvc depositsvc.DepositServiceIface
	systemSvc  systemsvc.SystemServiceIface

	ctx       context.Context
	ctxCancel context.CancelFunc
	closeSync sync.Once
}

func NewCuckoo(ctx context.Context, conf *config.Config) (*Cuckoo, error) {

	ctx, ctxCancel := context.WithCancel(ctx)
	cuckoo := &Cuckoo{
		ctx:       ctx,
		ctxCancel: ctxCancel,
	}

	ebus := eventbus.NewBus()
	bootEmitter, err := ebus.Emitter(&myevent.EvtHostBootComplete{})
	if err != nil {
		return nil, fmt.Errorf("ebus.Emitter error: %s", err.Error())
	}
	cuckoo.ebus = ebus

	localhost, ddht, err := NewHost(ctx, bootEmitter, conf)
	if err != nil {
		return nil, fmt.Errorf("NewHost error: %s", err.Error())
	}

	cuckoo.host = localhost
	cuckoo.ddht = ddht

	routingDiscovery := drouting.NewRoutingDiscovery(ddht)

	ds, err := datastore.New(datastore.Config{
		Path: filepath.Join(conf.StorageDir, "leveldb"),
	})
	if err != nil {
		return nil, fmt.Errorf("datastore.New error: %w", err)
	}

	cuckoo.accountSvc, err = accountsvc.NewAccountService(ctx, conf.AccountService, cuckoo.host, ds, ebus, routingDiscovery)
	if err != nil {
		return nil, fmt.Errorf("accountsvc.NewAccountService error: %s", err.Error())
	}

	cuckoo.contactSvc, err = contactsvc.NewContactService(ctx, conf.ContactService, cuckoo.host, ds, ebus, routingDiscovery)
	if err != nil {
		return nil, fmt.Errorf("contactsvc.NewContactService error: %s", err.Error())
	}

	cuckoo.groupSvc, err = groupsvc.NewGroupService(ctx, conf.GroupService, cuckoo.host, ds, ebus, routingDiscovery, cuckoo.accountSvc)
	if err != nil {
		return nil, fmt.Errorf("group.NewGroupService error: %s", err.Error())
	}

	// cuckoo.depositSvc, err = depositsvc.NewDepositService(ctx, conf.DepositService, cuckoo.host, ds, ebus, routingDiscovery)
	// if err != nil {
	// 	return nil, fmt.Errorf("deposit.NewDepositService error: %s", err.Error())
	// }

	cuckoo.systemSvc, err = systemsvc.NewSystemService(ctx, cuckoo.host, ds, ebus, cuckoo.accountSvc, cuckoo.contactSvc, cuckoo.groupSvc)
	if err != nil {
		return nil, fmt.Errorf("deposit.NewDepositService error: %s", err.Error())
	}

	return cuckoo, nil
}

func (c *Cuckoo) GetHost() (host.Host, error) {
	if c.host == nil {
		return nil, fmt.Errorf("host not start")
	}
	return c.host, nil
}

func (c *Cuckoo) GetAccountSvc() (accountsvc.AccountServiceIface, error) {
	if c.accountSvc == nil {
		return nil, fmt.Errorf("peer service not start")
	}
	return c.accountSvc, nil
}

func (c *Cuckoo) GetContactSvc() (contactsvc.ContactServiceIface, error) {
	if c.contactSvc == nil {
		return nil, fmt.Errorf("peer service not start")
	}
	return c.contactSvc, nil
}

func (c *Cuckoo) GetGroupSvc() (groupsvc.GroupServiceIface, error) {
	if c.groupSvc == nil {
		return nil, fmt.Errorf("group service not start")
	}
	return c.groupSvc, nil
}

func (c *Cuckoo) GetDepositSvc() (depositsvc.DepositServiceIface, error) {
	if c.depositSvc == nil {
		return nil, fmt.Errorf("get deposit service error, server not start")
	}
	return c.depositSvc, nil
}

func (c *Cuckoo) GetSystemSvc() (systemsvc.SystemServiceIface, error) {
	if c.systemSvc == nil {
		return nil, fmt.Errorf("system service not start")
	}
	return c.systemSvc, nil
}

func (c *Cuckoo) GetLanPeerIDs() ([]peer.ID, error) {
	if c.ddht == nil {
		return nil, fmt.Errorf("ddht is nil")
	}

	return c.ddht.LAN.RoutingTable().ListPeers(), nil
}

func (c *Cuckoo) GetEbus() (event.Bus, error) {
	if c.ebus == nil {
		return nil, fmt.Errorf("ebus is nil")
	}

	return c.ebus, nil
}

func (c *Cuckoo) Close() {
	c.closeSync.Do(func() {
		c.ctxCancel()
		if c.host != nil {
			c.host.Close()
		}
		if c.accountSvc != nil {
			c.accountSvc.Close()
		}
		if c.groupSvc != nil {
			c.groupSvc.Close()
		}
		if c.depositSvc != nil {
			c.depositSvc.Close()
		}
	})
}
