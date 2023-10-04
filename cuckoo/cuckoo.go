package cuckoo

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/internal/datastore"
	myevent "github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/configsvc"
	"github.com/jianbo-zh/dchat/service/contactsvc"
	"github.com/jianbo-zh/dchat/service/depositsvc"
	"github.com/jianbo-zh/dchat/service/filesvc"
	"github.com/jianbo-zh/dchat/service/groupsvc"
	"github.com/jianbo-zh/dchat/service/sessionsvc"
	"github.com/jianbo-zh/dchat/service/systemsvc"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
)

type CuckooGetter interface {
	GetCuckoo() (*Cuckoo, error)
}

type Cuckoo struct {
	host myhost.Host

	ddht *dual.DHT
	ebus event.Bus

	sessionSvc sessionsvc.SessionServiceIface
	configSvc  configsvc.ConfigServiceIface
	accountSvc accountsvc.AccountServiceIface
	contactSvc contactsvc.ContactServiceIface
	groupSvc   groupsvc.GroupServiceIface
	depositSvc depositsvc.DepositServiceIface
	systemSvc  systemsvc.SystemServiceIface
	fileSvc    filesvc.FileServiceIface

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
		Path: filepath.Join(conf.DataDir, config.DefaultLevelDBDir),
	})
	if err != nil {
		return nil, fmt.Errorf("datastore.New error: %w", err)
	}

	cuckoo.configSvc, err = configsvc.NewConfigService(ctx, conf, ebus)
	if err != nil {
		return nil, fmt.Errorf("config svc new config service error: %w", err)
	}

	cuckoo.fileSvc, err = filesvc.NewFileService(ctx, conf.FileService, cuckoo.host, ds, ebus)
	if err != nil {
		return nil, fmt.Errorf("deposit.NewFileService error: %s", err.Error())
	}

	cuckoo.sessionSvc, err = sessionsvc.NewSessionService(ctx, cuckoo.host, ds, ebus)
	if err != nil {
		return nil, fmt.Errorf("deposit.NewSessionService error: %s", err.Error())
	}

	cuckoo.accountSvc, err = accountsvc.NewAccountService(ctx, "avatardir", cuckoo.host, ds, ebus, routingDiscovery)
	if err != nil {
		return nil, fmt.Errorf("accountsvc.NewAccountService error: %s", err.Error())
	}

	cuckoo.contactSvc, err = contactsvc.NewContactService(ctx, cuckoo.host, ds, ebus, cuckoo.accountSvc)
	if err != nil {
		return nil, fmt.Errorf("contactsvc.NewContactService error: %s", err.Error())
	}

	cuckoo.groupSvc, err = groupsvc.NewGroupService(ctx, cuckoo.host, ds, ebus, routingDiscovery, cuckoo.accountSvc, cuckoo.contactSvc, cuckoo.sessionSvc)
	if err != nil {
		return nil, fmt.Errorf("group.NewGroupService error: %s", err.Error())
	}

	cuckoo.depositSvc, err = depositsvc.NewDepositService(ctx, conf.DepositService, cuckoo.host, ds, ebus)
	if err != nil {
		return nil, fmt.Errorf("deposit.NewDepositService error: %s", err.Error())
	}

	cuckoo.systemSvc, err = systemsvc.NewSystemService(ctx, cuckoo.host, ds, ebus, cuckoo.accountSvc, cuckoo.contactSvc, cuckoo.groupSvc)
	if err != nil {
		return nil, fmt.Errorf("deposit.NewSystemService error: %s", err.Error())
	}

	return cuckoo, nil
}

func (c *Cuckoo) GetHost() (myhost.Host, error) {
	if c.host == nil {
		return nil, fmt.Errorf("host not start")
	}
	return c.host, nil
}

func (c *Cuckoo) GetConfigSvc() (configsvc.ConfigServiceIface, error) {
	if c.configSvc == nil {
		return nil, fmt.Errorf("config service not start")
	}
	return c.configSvc, nil
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

func (c *Cuckoo) GetFileSvc() (filesvc.FileServiceIface, error) {
	if c.fileSvc == nil {
		return nil, fmt.Errorf("config service not start")
	}
	return c.fileSvc, nil
}

func (c *Cuckoo) GetSessionSvc() (sessionsvc.SessionServiceIface, error) {
	if c.sessionSvc == nil {
		return nil, fmt.Errorf("session service not start")
	}
	return c.sessionSvc, nil
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
