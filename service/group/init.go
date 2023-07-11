package group

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/group/datastore"
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

func Init(h host.Host, rdiscvry *drouting.RoutingDiscovery, ids ipfsds.Batching, ev event.Bus, opts ...Option) (*GroupService, error) {
	groupDs := datastore.GroupWrap(ids)

	groupsvc = &GroupService{
		datastore: groupDs,
	}

	if err := groupsvc.Apply(opts...); err != nil {
		return nil, err
	}

	groupsvc.adminSvc = admin.NewGroupAdminService(h, groupDs, ev)
	groupsvc.networkSvc = network.NewGroupNetworkService(h, rdiscvry, groupDs, ev)

	if err := groupsvc.networkSvc.Start(); err != nil {
		return nil, err
	}

	return groupsvc, nil
}
