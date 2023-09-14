package event

import "github.com/libp2p/go-libp2p/core/peer"

type EvtInviteJoinGroup struct {
	PeerIDs       []peer.ID
	GroupID       string
	GroupName     string
	GroupAvatar   string
	GroupLamptime uint64
}