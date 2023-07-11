package service

import (
	ipfsds "github.com/ipfs/go-datastore"
	groupsvc "github.com/jianbo-zh/dchat/service/group"
	peersvc "github.com/jianbo-zh/dchat/service/peer"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
)

var topsvc *TopService

type TopService struct {
	host      host.Host
	datastore ipfsds.Batching
	eventBus  event.Bus

	peerSvc  peersvc.PeerServiceIface
	groupSvc groupsvc.GroupServiceIface
}

func Init(localhost host.Host, rdiscvry *drouting.RoutingDiscovery, ids ipfsds.Batching) error {

	eventbus := eventbus.NewBus()

	topsvc = &TopService{
		host:      localhost,
		datastore: ids,
		eventBus:  eventbus,
	}

	// 初始化群相关服务
	gsvc, err := groupsvc.Init(localhost, rdiscvry, ids, eventbus)
	if err != nil {
		return err
	}
	topsvc.groupSvc = gsvc

	// 初始化Peer相关服务
	psvc, err := peersvc.Init(localhost, ids, eventbus)
	if err != nil {
		return err
	}
	topsvc.peerSvc = psvc

	return nil
}
