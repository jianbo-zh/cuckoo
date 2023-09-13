package ds

import "github.com/libp2p/go-libp2p/core/peer"

type GroupID = string

type Group struct {
	ID     string
	Name   string
	Avatar string
}

type Member struct {
	PeerID peer.ID
}
