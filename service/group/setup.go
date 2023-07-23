package group

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/group/protocol/admin"
	"github.com/jianbo-zh/dchat/service/group/protocol/network"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var groupsvc *GroupService

type GroupService struct {
	adminSvc *admin.AdminService

	networkSvc *network.NetworkService
}

func Get() GroupServiceIface {
	if groupsvc == nil {
		panic("group must init before use")
	}

	return groupsvc
}

func Setup(lhost host.Host, rdiscvry *drouting.RoutingDiscovery, ids ipfsds.Batching, ebus event.Bus, opts ...Option) (*GroupService, error) {
	var err error

	groupsvc = &GroupService{}

	if err := groupsvc.Apply(opts...); err != nil {
		return nil, err
	}

	groupsvc.adminSvc, err = admin.NewAdminService(lhost, ids, ebus)
	if err != nil {
		return nil, err
	}

	groupsvc.networkSvc, err = network.NewNetworkService(lhost, rdiscvry, ids, ebus)
	if err != nil {
		return nil, err
	}

	return groupsvc, nil
}
