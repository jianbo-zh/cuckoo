package mytype

import "github.com/libp2p/go-libp2p/core/peer"

type Peer struct {
	ID             peer.ID
	Name           string
	Avatar         string
	DepositAddress peer.ID
}

type OnlineState int

const (
	OnlineStateUnknown OnlineState = 0  // 未知
	OnlineStateOnline  OnlineState = 1  // 在线
	OnlineStateOffline OnlineState = -1 // 离线
)
