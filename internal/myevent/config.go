package myevent

import "github.com/libp2p/go-libp2p/core/peer"

// EvtConfigBootstrapsChange 引导地址更新
type EvtConfigBootstrapsChange struct {
	Bootstraps []string
}

// EvtConfigPeeringPeersChange 对等连接Peers更新
type EvtConfigPeeringPeersChange struct {
	PeeringPeers []peer.AddrInfo
}

// EvtConfigEnableMDNSChange MDNS配置变更
type EvtConfigEnableMDNSChange struct {
	Enable bool
}

// EvtConfigEnableDepositServiceChange 寄存服务配置更新
type EvtConfigEnableDepositServiceChange struct {
	Enable bool
}

// EvtConfigDownloadDirChange 下载目录配置更新
type EvtConfigDownloadDirChange struct {
	DownloadDir string
}
