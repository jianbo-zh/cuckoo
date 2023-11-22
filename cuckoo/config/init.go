package config

import (
	"time"
)

func DefaultConfig() (*Config, error) {

	identity, err := CreateIdentity()
	if err != nil {
		return nil, err
	}

	bootstrapPeers, err := DefaultBootstrapPeers()
	if err != nil {
		return nil, err
	}

	conf := &Config{

		// setup the node's default addresses.
		// NOTE: two swarm listen addrs, one tcp, one utp.
		Addresses: Addresses{
			Swarm: []string{
				"/ip4/0.0.0.0/tcp/4001",
				"/ip6/::/tcp/4001",
				"/ip4/0.0.0.0/udp/4001/quic",
				"/ip4/0.0.0.0/udp/4001/quic-v1",
				"/ip4/0.0.0.0/udp/4001/quic-v1/webtransport",
				"/ip6/::/udp/4001/quic",
				"/ip6/::/udp/4001/quic-v1",
				"/ip6/::/udp/4001/quic-v1/webtransport",
			},
			Announce:       []string{},
			AppendAnnounce: []string{},
			NoAnnounce:     []string{},
		},
		Bootstrap: BootstrapPeerStrings(bootstrapPeers),
		Identity:  identity,
		Discovery: Discovery{
			MDNS: MDNS{
				Enabled: true,
			},
		},

		Ipns: Ipns{
			ResolveCacheSize: 128,
		},

		DNS: DNS{
			Resolvers: map[string]string{},
		},

		DepositService: DepositServiceConfig{
			EnableDepositService: false,
		},

		FileService: FileServiceConfig{
			DownloadDir: "./download",
		},
	}

	return conf, nil
}

// DefaultConnMgrHighWater is the default value for the connection managers
// 'high water' mark
const DefaultConnMgrHighWater = 96

// DefaultConnMgrLowWater is the default value for the connection managers 'low
// water' mark
const DefaultConnMgrLowWater = 32

// DefaultConnMgrGracePeriod is the default value for the connection managers
// grace period
const DefaultConnMgrGracePeriod = time.Second * 20

// DefaultConnMgrType is the default value for the connection managers
// type.
const DefaultConnMgrType = "basic"

// DefaultResourceMgrMinInboundConns is a MAGIC number that probably a good
// enough number of inbound conns to be a good network citizen.
const DefaultResourceMgrMinInboundConns = 800
