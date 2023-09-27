package types

import "github.com/libp2p/go-libp2p/core/peer"

type Peer struct {
	ID     peer.ID
	Name   string
	Avatar string
}

type OnlineState int

const (
	OnlineStateOnline  OnlineState = 1  // 在线
	OnlineStateUnknown OnlineState = 0  // 未知
	OnlineStateOffline OnlineState = -1 // 离线
)
