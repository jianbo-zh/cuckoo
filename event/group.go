package event

import (
	admpb "github.com/jianbo-zh/dchat/service/group/protocol/admin/pb"
	msgpb "github.com/jianbo-zh/dchat/service/group/protocol/message/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type EvtSendAdminLog struct {
	MsgType admpb.AdminLog_Type
	MsgData *admpb.AdminLog
}

type EvtRecvAdminLog struct {
	MsgType admpb.AdminLog_Type
	MsgData *admpb.AdminLog
}

type EvtForwardGroupMsg struct {
	Exclude *peer.ID
	MsgData *msgpb.GroupMsg
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
