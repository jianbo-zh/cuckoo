package contactsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/message/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (c *ContactSvc) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]Message, error) {
	var peerMsgs []Message

	msgs, err := c.msgSvc.GetMessages(ctx, peerID, offset, limit)
	if err != nil {
		return peerMsgs, err
	}

	for _, msg := range msgs {
		peerMsgs = append(peerMsgs, Message{
			ID:         msg.Id,
			Type:       convMsgType(msg.Type),
			SenderID:   peer.ID(msg.SenderId),
			ReceiverID: peer.ID(msg.ReceiverId),
			Payload:    msg.Payload,
			Timestamp:  msg.Timestamp,
			Lamportime: msg.Lamportime,
		})
	}

	return peerMsgs, nil
}

func (c *ContactSvc) SendTextMessage(ctx context.Context, peerID peer.ID, msg string) error {
	return c.msgSvc.SendTextMessage(ctx, peerID, msg)
}

func (c *ContactSvc) SendGroupInviteMessage(ctx context.Context, peerID peer.ID, groupID string) error {
	return c.msgSvc.SendGroupInviteMessage(ctx, peerID, groupID)
}

func convMsgType(mtype pb.Message_Type) MsgType {
	switch mtype {
	case pb.Message_TEXT:
		return MsgTypeText
	case pb.Message_AUDIO:
		return MsgTypeAudio
	case pb.Message_VIDEO:
		return MsgTypeVideo
	case pb.Message_INVITE:
		return MsgTypeInvite
	default:
		return MsgTypeText
	}
}
