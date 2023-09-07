package peer

import "github.com/libp2p/go-libp2p/core/peer"

type PeerInfo struct {
	PeerID peer.ID
	Name   string
	Avatar string
}

type ContactEntiy struct {
	PeerID   peer.ID
	Name     string
	Avatar   string
	AddTs    int64
	AccessTs int64
}

type Account struct {
	PeerID         peer.ID
	Name           string
	Avatar         string
	AutoAddContact bool
	AutoJoinGroup  bool
}
