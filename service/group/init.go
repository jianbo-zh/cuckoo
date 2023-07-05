package group

import (
	"github.com/jianbo-zh/dchat/datastore"
	"github.com/jianbo-zh/dchat/service/group/protocol/admin"
	"github.com/jianbo-zh/dchat/service/group/protocol/network"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var groupsvc *GroupService

type GroupService struct {
	datastore datastore.GroupIface

	adminSvc *admin.GroupAdminService

	networkSvc *network.GroupNetworkService
}

func Get() GroupServiceIface {
	if groupsvc == nil {
		panic("group must init before use")
	}

	return groupsvc
}

func Init(h host.Host, rdiscvry *drouting.RoutingDiscovery, ds datastore.GroupIface, ev event.Bus, opts ...Option) (*GroupService, error) {
	groupsvc = &GroupService{
		datastore: ds,
	}

	if err := groupsvc.Apply(opts...); err != nil {
		return nil, err
	}

	groupsvc.adminSvc = admin.NewGroupAdminService(h, ds, ev)
	groupsvc.networkSvc = network.NewGroupNetworkService(h, rdiscvry, ds, ev)

	if err := groupsvc.networkSvc.Start(); err != nil {
		return nil, err
	}

	return groupsvc, nil
}
