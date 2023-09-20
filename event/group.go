package event

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

type EvtGroupConnectChange struct {
	GroupID     string
	PeerID      peer.ID
	IsConnected bool
}

type Groups struct {
	GroupID     string
	PeerIDs     []peer.ID
	AcptPeerIDs []peer.ID
}
type EvtGroupsInit struct {
	Groups []Groups
}

type EvtGroupsChange struct {
	DeleteGroups []string
	AddGroups    []Groups
}

type EvtGroupMemberChange struct {
	GroupID     string
	PeerIDs     []peer.ID
	AcptPeerIDs []peer.ID
}
