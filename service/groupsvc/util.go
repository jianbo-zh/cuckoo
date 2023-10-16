package groupsvc

import (
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/grouppb"
	"github.com/libp2p/go-libp2p/core/peer"
)

func convertMessage(msg *pb.GroupMessage) *mytype.GroupMessage {
	msgState := mytype.MessageStateSending
	switch msg.SendState {
	case pb.GroupMessage_SendSucc:
		msgState = mytype.MessageStateSuccess

	case pb.GroupMessage_SendFail:
		msgState = mytype.MessageStateFail

	default:
		msgState = mytype.MessageStateSending
	}

	return &mytype.GroupMessage{
		ID:      msg.Id,
		GroupID: msg.GroupId,
		FromPeer: mytype.GroupMember{
			ID:     peer.ID(msg.CoreMessage.Member.Id),
			Name:   msg.CoreMessage.Member.Name,
			Avatar: msg.CoreMessage.Member.Avatar,
		},
		MsgType:   msg.CoreMessage.MsgType,
		MimeType:  msg.CoreMessage.MimeType,
		Payload:   msg.CoreMessage.Payload,
		IsDeposit: msg.IsDeposit,
		State:     msgState,
		Timestamp: msg.CreateTime,
	}
}
