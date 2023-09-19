package event

import (
	admpb "github.com/jianbo-zh/dchat/service/groupsvc/protocol/adminproto/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type EvtSendAdminLog struct {
	MsgType admpb.Log_LogType
	MsgData *admpb.Log
}

type EvtRecvAdminLog struct {
	MsgType admpb.Log_LogType
	MsgData *admpb.Log
}

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
