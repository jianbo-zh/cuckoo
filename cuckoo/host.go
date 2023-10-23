package cuckoo

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/libp2p/go-libp2p"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
)

func NewHost(ctx context.Context, bootEmitter event.Emitter, conf *config.Config, ebus event.Bus) (myhost.Host, *ddht.DHT, error) {

	privKey, err := conf.Identity.DecodePrivateKey("")
	if err != nil {
		return nil, nil, err
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
	peerChan := make(chan peer.AddrInfo)

	// 创建host节点
	localhost, err := libp2p.New(
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
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelayWithPeerSource(
			func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
				r := make(chan peer.AddrInfo)
				go func() {
					defer close(r)
					for ; numPeers != 0; numPeers-- {
						select {
						case v, ok := <-peerChan:
							if !ok {
								return
							}
							select {
							case r <- v:
							case <-ctx.Done():
								return
							}
						case <-ctx.Done():
							return
						}
					}
				}()
				return r
			},
			autorelay.WithMinInterval(0),
		),
	)
	if err != nil {
		panic(err)
	}

	if conf.Discovery.MDNS.Enabled {
		// 初始化mdns服务
		initMDNS(ctx, localhost)
	}

	// 查找中继节点
	go autoRelayFeeder(ctx, localhost, dualDHT, conf.Peering, peerChan)

	// 连接引导节点
	go func() {
		isAnySucc := false
		for _, addr := range conf.Bootstrap {
			pi, _ := peer.AddrInfoFromString(addr)
			if err := localhost.Connect(ctx, *pi); err == nil {
				isAnySucc = true
				fmt.Println("connect peer: " + pi.String())
			}
		}

		fmt.Println("boot complete")

		bootEmitter.Emit(myevent.EvtHostBootComplete{IsSucc: isAnySucc})
	}()

	lhost, err := myhost.NewHost(localhost, ebus)
	if err != nil {
		return nil, nil, fmt.Errorf("new host error: %w", err)
	}

	return lhost, dualDHT, nil
}

func autoRelayFeeder(ctx context.Context, h host.Host, dualDHT *ddht.DHT, peering config.Peering, peerChan chan<- peer.AddrInfo) {
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
		for _, trustedPeer := range peering.Peers {
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
