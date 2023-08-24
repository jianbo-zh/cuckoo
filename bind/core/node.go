package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/jianbo-zh/dchat/bind/utils"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"go.uber.org/zap"

	"github.com/jianbo-zh/dchat/cmd/chat/config"
	libp2p_config "github.com/jianbo-zh/dchat/cmd/chat/config"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	libp2p_mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	manet "github.com/multiformats/go-multiaddr/net"
)

type Node struct {
	mdnsLocker  sync.Locker
	mdnsLocked  bool
	mdnsService mdns.Service

	libp2phost host.Host
}

func (n *Node) Host() host.Host {
	return n.libp2phost
}

func NewNode(config *NodeConfig) (*Node, error) {
	if config == nil {
		config = NewNodeConfig()
	}

	ctx := context.Background()

	// Set up netdriver.
	if config.netDriver != nil {
		logger, _ := zap.NewDevelopment()
		inet := &inet{
			net:    config.netDriver,
			logger: logger,
		}
		utils.SetNetDriver(inet)
		manet.SetNetInterface(inet)
	}

	libconf, err := libp2p_config.Init()
	if err != nil {
		panic(err)
	}

	mdnsLocked := false
	if libconf.Discovery.MDNS.Enabled && config.mdnsLockerDriver != nil {
		config.mdnsLockerDriver.Lock()
		mdnsLocked = true

		// set false
		libconf.Discovery.MDNS.Enabled = false
	}

	libp2phost, err := startLibp2pHost(ctx, *libconf)
	if err != nil {
		if mdnsLocked {
			config.mdnsLockerDriver.Unlock()
		}
		panic(err)
	}

	var mdnsService libp2p_mdns.Service = nil
	if mdnsLocked {
		// Restore.
		libconf.Discovery.MDNS.Enabled = true

		mdnslogger, _ := zap.NewDevelopment()
		dh := utils.DiscoveryHandler(ctx, mdnslogger, libp2phost)
		mdnsService = utils.NewMdnsService(mdnslogger, libp2phost, utils.MDNSServiceName, dh)

		// Start the mDNS service.
		// Get multicast interfaces.
		ifaces, err := utils.GetMulticastInterfaces()
		if err != nil {
			if mdnsLocked {
				config.mdnsLockerDriver.Unlock()
			}
			return nil, err
		}

		// If multicast interfaces are found, start the mDNS service.
		if len(ifaces) > 0 {
			mdnslogger.Info("starting mdns")
			if err := mdnsService.Start(); err != nil {
				if mdnsLocked {
					config.mdnsLockerDriver.Unlock()
				}
				return nil, fmt.Errorf("unable to start mdns service: %w", err)
			}
		} else {
			mdnslogger.Error("unable to start mdns service, no multicast interfaces found")
		}
	}

	// 连接引导节点
	go func() {
		for _, addr := range libconf.Bootstrap {
			pi, _ := peer.AddrInfoFromString(addr)
			if err := libp2phost.Connect(ctx, *pi); err == nil {
				fmt.Println("connect", pi.ID.String())
			}
		}
	}()

	return &Node{
		mdnsLocker:  config.mdnsLockerDriver,
		mdnsLocked:  mdnsLocked,
		mdnsService: mdnsService,

		libp2phost: libp2phost,
	}, nil
}

// func StartNode() {
// 	go startNode()
// }

// func startNode() {

// 	conf, err := config.Init()
// 	if err != nil {
// 		panic(err)
// 	}

// 	localhost, err := startLocalNode(context.Background(), *conf)
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println("PeerID: ", localhost.ID().String())

// 	ticker := time.NewTicker(10 * time.Second)
// 	defer ticker.Stop()

// 	for t := range ticker.C {
// 		fmt.Println("time: ", t)
// 		fmt.Println("address: ", localhost.Addrs())
// 	}
// }

func startLibp2pHost(ctx context.Context, conf libp2p_config.Config) (host.Host, error) {

	privKey, err := conf.Identity.DecodePrivateKey("")
	if err != nil {
		return nil, err
	}

	connmgr, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		panic(err)
	}

	var dualDHT *ddht.DHT
	// peerChan := make(chan peer.AddrInfo)

	// 创建host节点
	localnode, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(conf.Addresses.Swarm...),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		libp2p.ConnectionManager(connmgr),
		libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			var err error
			dualDHT, err = ddht.New(ctx, h)
			return dualDHT, err
		}),
		libp2p.EnableNATService(),
		// libp2p.EnableRelay(),
		// libp2p.EnableAutoRelayWithPeerSource(
		// 	func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
		// 		r := make(chan peer.AddrInfo)
		// 		go func() {
		// 			defer close(r)
		// 			for ; numPeers != 0; numPeers-- {
		// 				select {
		// 				case v, ok := <-peerChan:
		// 					if !ok {
		// 						return
		// 					}
		// 					select {
		// 					case r <- v:
		// 					case <-ctx.Done():
		// 						return
		// 					}
		// 				case <-ctx.Done():
		// 					return
		// 				}
		// 			}
		// 		}()
		// 		return r
		// 	},
		// 	autorelay.WithMinInterval(0),
		// ),
	)
	if err != nil {
		panic(err)
	}

	// // 查找中继节点
	// go autoRelayFeeder(ctx, localnode, dualDHT, conf.Peering, peerChan)

	// // 连接引导节点
	// go func() {
	// 	isAnySucc := false
	// 	for _, addr := range conf.Bootstrap {
	// 		pi, _ := peer.AddrInfoFromString(addr)
	// 		if err := localnode.Connect(ctx, *pi); err == nil {
	// 			isAnySucc = true
	// 		}
	// 	}

	// 	fmt.Println("boot complete")
	// 	if isAnySucc {
	// 		fmt.Println("boot success")
	// 	}
	// }()

	return localnode, nil
}

func autoRelayFeeder(ctx context.Context, h host.Host, dualDHT *ddht.DHT, cfgPeering config.Peering, peerChan chan<- peer.AddrInfo) {
	// Feed peers more often right after the bootstrap, then backoff
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 15 * time.Second
	bo.Multiplier = 3
	bo.MaxInterval = 1 * time.Hour
	bo.MaxElapsedTime = 0 // never stop
	t := backoff.NewTicker(bo)
	defer t.Stop()

	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			return
		}

		// Always feed trusted IDs (Peering.Peers in the config)
		for _, trustedPeer := range cfgPeering.Peers {
			if len(trustedPeer.Addrs) == 0 {
				continue
			}
			select {
			case peerChan <- trustedPeer:
			case <-ctx.Done():
				return
			}
		}

		// Additionally, feed closest peers discovered via DHT
		if dualDHT == nil {
			/* noop due to missing dht.WAN. happens in some unit tests,
			   not worth fixing as we will refactor this after go-libp2p 0.20 */
			continue
		}

		closestPeers, err := dualDHT.WAN.GetClosestPeers(ctx, h.ID().String())
		if err != nil {
			// no-op: usually 'failed to find any peer in table' during startup
			continue
		}

		for _, p := range closestPeers {
			addrs := h.Peerstore().Addrs(p)
			if len(addrs) == 0 {
				continue
			}
			dhtPeer := peer.AddrInfo{ID: p, Addrs: addrs}
			select {
			case peerChan <- dhtPeer:
			case <-ctx.Done():
				return
			}
		}
	}
}

type discoveryNotifee struct {
	h   host.Host
	ctx context.Context
}

func (m *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if m.h.Network().Connectedness(pi.ID) != network.Connected {
		fmt.Printf("Found %s!\n", pi.ID.ShortString())
		m.h.Connect(m.ctx, pi)
	}
}

// Initialize the MDNS service
func initMDNS(ctx context.Context, peerhost host.Host) {
	// register with service so that we get notified about peer discovery
	n := &discoveryNotifee{h: peerhost, ctx: ctx}

	// An hour might be a long long period in practical applications. But this is fine for us
	ser := mdns.NewMdnsService(peerhost, "", n)
	if err := ser.Start(); err != nil {
		panic(err)
	}
}
