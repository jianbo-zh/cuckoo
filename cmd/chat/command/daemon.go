package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/jianbo-zh/dchat/cmd/chat/config"
	"github.com/jianbo-zh/dchat/cmd/chat/httpapi"
	"github.com/jianbo-zh/dchat/internal/datastore"
	"github.com/jianbo-zh/dchat/service"
	"github.com/libp2p/go-libp2p"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/urfave/cli/v2"
)

var DaemonCmd cli.Command

func init() {
	DaemonCmd = cli.Command{
		Name:  "daemon",
		Usage: "run daemon progress",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Value:   "",
				Usage:   "set listen host",
				Aliases: []string{"H"},
			},
			&cli.UintFlag{
				Name:    "port",
				Value:   9000,
				Usage:   "set listen port",
				Aliases: []string{"P"},
			},
			&cli.StringFlag{
				Name:  "root",
				Value: "",
				Usage: "set root dir",
			},
		},
		Action: func(cCtx *cli.Context) error {

			rootpath := cCtx.String("root")
			cCtx.App.Run([]string{"dchat", "init", "--noprint", "--root=" + rootpath})

			// 启动节点主机
			confpath, err := config.Filename(rootpath, config.DefaultConfigFile)
			if err != nil {
				return err
			}

			bs, err := os.ReadFile(confpath)
			if err != nil {
				return err
			}

			var conf config.Config
			if err = json.Unmarshal(bs, &conf); err != nil {
				return err
			}

			// 创建host节点
			localnode, rdiscvry, err := startLocalNode(context.Background(), conf)
			if err != nil {
				return err
			}
			defer localnode.Close()

			lvpath, err := config.Filename(rootpath, "leveldb")
			if err != nil {
				return err
			}
			ds, err := datastore.New(datastore.Config{Path: lvpath})
			if err != nil {
				return err
			}

			// 注入运行服务
			if err = service.Init(localnode, rdiscvry, ds); err != nil {
				return err
			}

			fmt.Println("PeerID: ", localnode.ID())

			// ticker := time.NewTicker(5 * time.Second)
			// defer ticker.Stop()

			// for t := range ticker.C {
			// 	fmt.Println()
			// 	fmt.Printf("PeerInfo: %s", t)
			// 	for _, addr := range localnode.Addrs() {
			// 		fmt.Println(addr)
			// 	}
			// }

			// 启动HTTP服务
			go func() {
				httpapi.Daemon(httpapi.Config{
					Host: cCtx.String("host"),
					Port: cCtx.Uint("port"),
				})
			}()

			select {}
		},
	}
}

func startLocalNode(ctx context.Context, conf config.Config) (host.Host, *drouting.RoutingDiscovery, error) {

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

	// 初始化mdns服务
	initMDNS(ctx, localnode)

	// 查找中继节点
	go autoRelayFeeder(ctx, localnode, dualDHT, conf.Peering, peerChan)

	// 连接引导节点
	for _, addr := range conf.Bootstrap {
		pi, _ := peer.AddrInfoFromString(addr)
		localnode.Connect(ctx, *pi)
	}

	routingDiscovery := drouting.NewRoutingDiscovery(dualDHT)

	return localnode, routingDiscovery, nil
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
