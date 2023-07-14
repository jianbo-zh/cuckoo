package service

import (
	ipfsds "github.com/ipfs/go-datastore"
	peersvc "github.com/jianbo-zh/dchat/service/peer"
	"github.com/libp2p/go-libp2p/core/host"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
)

func Init(localhost host.Host, rdiscvry *drouting.RoutingDiscovery, ids ipfsds.Batching) error {

	eventBus := eventbus.NewBus()

	// // 初始化群相关服务
	// gsvc, err := groupsvc.Init(localhost, rdiscvry, ids, eventBus)
	// if err != nil {
	// 	return err
	// }
	// topsvc.groupSvc = gsvc

	// 初始化Peer相关服务
	_, err := peersvc.Init(localhost, ids, eventBus)
	if err != nil {
		return err
	}

	return nil
}
