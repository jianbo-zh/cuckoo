package myevent

import "github.com/libp2p/go-libp2p/core/peer"

type EvtConfigBootstrapsChange struct {
	Bootstraps []string
}

type EvtConfigPeeringPeersChange struct {
	PeeringPeers []peer.AddrInfo
}

type EvtConfigEnableMDNSChange struct {
	Enable bool
}

type EvtConfigEnableDepositServiceChange struct {
	Enable bool
}

type EvtConfigDownloadDirChange struct {
	DownloadDir string
}
