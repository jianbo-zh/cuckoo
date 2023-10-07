package contactsvc

import (
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/contactpb"
	"github.com/libp2p/go-libp2p/core/peer"
)

func convertMessage(msg *pb.ContactMessage) *mytype.ContactMessage {
	msgState := mytype.MessageStateSending
	switch msg.State {
	case pb.ContactMessage_Success:
		msgState = mytype.MessageStateSuccess

	case pb.ContactMessage_Fail:
		msgState = mytype.MessageStateFail

	default:
		msgState = mytype.MessageStateSending
	}

	return &mytype.ContactMessage{
		ID:         msg.Id,
		MsgType:    msg.MsgType,
		MimeType:   msg.MimeType,
		FromPeerID: peer.ID(msg.FromPeerId),
		ToPeerID:   peer.ID(msg.ToPeerId),
		Payload:    msg.Payload,
		State:      msgState,
		Timestamp:  msg.Timestamp,
		Lamportime: msg.Lamportime,
	}
}
