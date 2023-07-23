package event

import (
	admpb "github.com/jianbo-zh/dchat/service/group/protocol/admin/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type EvtSendAdminLog struct {
	MsgType admpb.Log_Type
	MsgData *admpb.Log
}

type EvtRecvAdminLog struct {
	MsgType admpb.Log_Type
	MsgData *admpb.Log
}

type EvtGroupConnectChange struct {
	GroupID     string
	PeerID      peer.ID
	IsConnected bool
}

type EvtGroupsInit struct {
	Groups []struct {
		GroupID string
		PeerIDs []peer.ID
	}
}

type EvtGroupsChange struct {
	DeleteGroups []string
	AddGroups    []struct {
		GroupID string
		PeerIDs []peer.ID
	}
}

type EvtGroupMemberChange struct {
	GroupID string
	PeerIDs []peer.ID
}
