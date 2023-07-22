package service

import (
	ipfsds "github.com/ipfs/go-datastore"
	depositsvc "github.com/jianbo-zh/dchat/service/deposit"
	peersvc "github.com/jianbo-zh/dchat/service/peer"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

func Setup(lhost host.Host, rdiscvry *drouting.RoutingDiscovery, ebus event.Bus, ids ipfsds.Batching) error {

	var err error
	// // 初始化群相关服务
	// gsvc, err := groupsvc.Setup(lhost, rdiscvry, ids, eventBus)
	// if err != nil {
	// 	return err
	// }
	// topsvc.groupSvc = gsvc

	// 初始化Peer相关服务
	_, err = peersvc.Setup(lhost, ids, ebus)
	if err != nil {
		return err
	}

	// 寄存服务
	_, err = depositsvc.Setup(lhost, rdiscvry, ids, ebus, depositsvc.WithService())
	if err != nil {
		return err
	}

	return nil
}
