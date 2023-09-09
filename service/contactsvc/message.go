package contactsvc

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/message/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (c *ContactSvc) GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*Message, error) {
	msg, err := c.msgSvc.GetMessage(ctx, peerID, msgID)
	if err != nil {
		return nil, fmt.Errorf("msgSvc.GetMessage error: %w", err)
	}

	return &Message{
		ID:         msg.Id,
		MsgType:    convMsgType(msg.MsgType),
		MimeType:   msg.MimeType,
		FromPeerID: peer.ID(msg.FromPeerId),
		ToPeerID:   peer.ID(msg.ToPeerId),
		Payload:    msg.Payload,
		Timestamp:  msg.Timestamp,
		Lamportime: msg.Lamportime,
	}, nil
}

func (c *ContactSvc) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]Message, error) {
	var peerMsgs []Message

	msgs, err := c.msgSvc.GetMessages(ctx, peerID, offset, limit)
	if err != nil {
		return peerMsgs, err
	}

	for _, msg := range msgs {
		peerMsgs = append(peerMsgs, Message{
			ID:         msg.Id,
			MsgType:    convMsgType(msg.MsgType),
			MimeType:   msg.MimeType,
			FromPeerID: peer.ID(msg.FromPeerId),
			ToPeerID:   peer.ID(msg.ToPeerId),
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

func convMsgType(mtype pb.Message_MsgType) MsgType {
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
