package contactsvc

import (
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/service/contactsvc/protobuf/pb/contactpb"
	"github.com/libp2p/go-libp2p/core/peer"
)

func convertMessage(msg *pb.ContactMessage) *mytype.ContactMessage {
	msgState := mytype.MessageStateSending
	switch msg.SendState {
	case pb.ContactMessage_SendSucc:
		msgState = mytype.MessageStateSuccess

	case pb.ContactMessage_SendFail:
		msgState = mytype.MessageStateFail

	default:
		msgState = mytype.MessageStateSending
	}

	return &mytype.ContactMessage{
		ID:         msg.Id,
		MsgType:    msg.CoreMessage.MsgType,
		MimeType:   msg.CoreMessage.MimeType,
		FromPeerID: peer.ID(msg.CoreMessage.FromPeerId),
		ToPeerID:   peer.ID(msg.CoreMessage.ToPeerId),
		Payload:    msg.CoreMessage.Payload,
		IsDeposit:  msg.IsDeposit,
		State:      msgState,
		Timestamp:  msg.CreateTime,
	}
}
