package service

import (
	ipfsds "github.com/ipfs/go-datastore"
	peersvc "github.com/jianbo-zh/dchat/service/peer"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

func Init(lhost host.Host, rdiscvry *drouting.RoutingDiscovery, ebus event.Bus, ids ipfsds.Batching) error {

	// // 初始化群相关服务
	// gsvc, err := groupsvc.Init(lhost, rdiscvry, ids, eventBus)
	// if err != nil {
	// 	return err
	// }
	// topsvc.groupSvc = gsvc

	// 初始化Peer相关服务
	_, err := peersvc.Init(lhost, ids, ebus)
	if err != nil {
		return err
	}

	return nil
}
