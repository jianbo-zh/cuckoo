package cuckoo

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/jianbo-zh/dchat/cuckoo/config"
	myevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/datastore"
	depositsvc "github.com/jianbo-zh/dchat/service/depositsvc"
	groupsvc "github.com/jianbo-zh/dchat/service/groupsvc"
	peersvc "github.com/jianbo-zh/dchat/service/peersvc"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
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

	peerSvc    peersvc.PeerServiceIface
	groupSvc   groupsvc.GroupServiceIface
	depositSvc depositsvc.DepositServiceIface

	ctx       context.Context
	ctxCancel context.CancelFunc
	closeSync sync.Once
}

func NewCuckoo(ctx context.Context, conf *config.Config) (*Cuckoo, error) {

	ebus := eventbus.NewBus()
	bootEmitter, err := ebus.Emitter(&myevent.EvtHostBootComplete{})
	if err != nil {
		return nil, fmt.Errorf("ebus.Emitter error: %s", err.Error())
	}

	ctx, ctxCancel := context.WithCancel(ctx)

	cuckoo := &Cuckoo{
		ctx:       ctx,
		ctxCancel: ctxCancel,
	}

	localhost, ddht, err := NewHost(ctx, bootEmitter, conf)
	if err != nil {
		return nil, fmt.Errorf("NewHost error: %s", err.Error())
	}

	cuckoo.host = localhost
	cuckoo.ddht = ddht

	routingDiscovery := drouting.NewRoutingDiscovery(ddht)

	ds, err := datastore.New(datastore.Config{
		Path: filepath.Join(conf.StorageDir, "leveldb2"),
	})
	if err != nil {
		return nil, fmt.Errorf("datastore.New error: %w", err)
	}

	cuckoo.peerSvc, err = peersvc.NewPeerService(conf.PeerMessage, cuckoo.host, ds, ebus, routingDiscovery, conf.AvatarDir)
	if err != nil {
		return nil, fmt.Errorf("peer.NewPeerService error: %s", err.Error())
	}

	cuckoo.groupSvc, err = groupsvc.NewGroupService(conf.GroupMessage, cuckoo.host, ds, ebus, routingDiscovery)
	if err != nil {
		return nil, fmt.Errorf("group.NewGroupService error: %s", err.Error())
	}

	cuckoo.depositSvc, err = depositsvc.NewDepositService(conf.DepositMessage, cuckoo.host, ds, ebus, routingDiscovery)
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

func (c *Cuckoo) GetPeerSvc() (peersvc.PeerServiceIface, error) {
	if c.peerSvc == nil {
		return nil, fmt.Errorf("peer service not start")
	}
	return c.peerSvc, nil
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

func (c *Cuckoo) GetLanPeerIDs() ([]peer.ID, error) {
	if c.ddht == nil {
		return nil, fmt.Errorf("ddht is nil")
	}

	return c.ddht.LAN.RoutingTable().ListPeers(), nil
}

func (c *Cuckoo) Close() {
	c.closeSync.Do(func() {
		c.ctxCancel()
		if c.host != nil {
			c.host.Close()
		}
		if c.peerSvc != nil {
			c.peerSvc.Close()
		}
		if c.groupSvc != nil {
			c.groupSvc.Close()
		}
		if c.depositSvc != nil {
			c.depositSvc.Close()
		}
	})
}
