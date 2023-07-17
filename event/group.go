package event

import (
	admpb "github.com/jianbo-zh/dchat/service/group/protocol/admin/pb"
	msgpb "github.com/jianbo-zh/dchat/service/group/protocol/message/pb"
	networkpb "github.com/jianbo-zh/dchat/service/group/protocol/network/pb"
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

type EvtGroupPeerConnectChange struct {
	MsgData *networkpb.GroupConnect
}
