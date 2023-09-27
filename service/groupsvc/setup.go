package groupsvc

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myhost"
	admin "github.com/jianbo-zh/dchat/service/groupsvc/protocol/adminproto"
	network "github.com/jianbo-zh/dchat/service/groupsvc/protocol/networkproto"
	"github.com/libp2p/go-libp2p/core/event"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var groupsvc *GroupService

func Setup(lhost myhost.Host, rdiscvry *drouting.RoutingDiscovery, ids ipfsds.Batching, ebus event.Bus) (*GroupService, error) {
	var err error

	groupsvc = &GroupService{}

	groupsvc.adminProto, err = admin.NewAdminProto(lhost, ids, ebus)
	if err != nil {
		return nil, err
	}

	groupsvc.networkProto, err = network.NewNetworkProto(lhost, rdiscvry, ids, ebus)
	if err != nil {
		return nil, err
	}

	return groupsvc, nil
}
