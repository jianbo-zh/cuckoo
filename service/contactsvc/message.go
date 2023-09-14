package contactsvc

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (c *ContactSvc) GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*Message, error) {
	msg, err := c.msgProto.GetMessage(ctx, peerID, msgID)
	if err != nil {
		return nil, fmt.Errorf("msgSvc.GetMessage error: %w", err)
	}

	return &Message{
		ID:         msg.Id,
		MsgType:    decodeMsgType(msg.MsgType),
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

	msgs, err := c.msgProto.GetMessages(ctx, peerID, offset, limit)
	if err != nil {
		return peerMsgs, err
	}

	for _, msg := range msgs {
		peerMsgs = append(peerMsgs, Message{
			ID:         msg.Id,
			MsgType:    decodeMsgType(msg.MsgType),
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

func (c *ContactSvc) SendMessage(ctx context.Context, peerID peer.ID, msgType types.MsgType, mimeType string, payload []byte) error {
	return c.msgProto.SendMessage(ctx, peerID, encodeMsgType(msgType), mimeType, payload)
}

func (c *ContactSvc) ClearMessage(ctx context.Context, peerID peer.ID) error {
	return c.msgProto.ClearMessage(ctx, peerID)
}
