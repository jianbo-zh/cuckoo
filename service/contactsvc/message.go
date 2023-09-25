package contactsvc

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (c *ContactSvc) GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*types.ContactMessage, error) {
	msg, err := c.msgProto.GetMessage(ctx, peerID, msgID)
	if err != nil {
		return nil, fmt.Errorf("msgSvc.GetMessage error: %w", err)
	}

	return &types.ContactMessage{
		ID:         msg.Id,
		MsgType:    msg.MsgType,
		MimeType:   msg.MimeType,
		FromPeerID: peer.ID(msg.FromPeerId),
		ToPeerID:   peer.ID(msg.ToPeerId),
		Payload:    msg.Payload,
		Timestamp:  msg.Timestamp,
		Lamportime: msg.Lamportime,
	}, nil
}

func (c *ContactSvc) DeleteMessage(ctx context.Context, peerID peer.ID, msgID string) error {
	return c.msgProto.DeleteMessage(ctx, peerID, msgID)
}

func (c *ContactSvc) GetMessageData(ctx context.Context, peerID peer.ID, msgID string) ([]byte, error) {
	bs, err := c.msgProto.GetMessageData(ctx, peerID, msgID)
	if err != nil {
		return nil, fmt.Errorf("proto get msg data error: %w", err)
	}

	return bs, nil
}

func (c *ContactSvc) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]types.ContactMessage, error) {
	var peerMsgs []types.ContactMessage

	msgs, err := c.msgProto.GetMessages(ctx, peerID, offset, limit)
	if err != nil {
		return peerMsgs, err
	}

	for _, msg := range msgs {
		peerMsgs = append(peerMsgs, types.ContactMessage{
			ID:         msg.Id,
			MsgType:    msg.MsgType,
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

func (c *ContactSvc) SendMessage(ctx context.Context, peerID peer.ID, msgType string, mimeType string, payload []byte) (string, error) {
	return c.msgProto.SendMessage(ctx, peerID, msgType, mimeType, payload)
}

func (c *ContactSvc) ClearMessage(ctx context.Context, peerID peer.ID) error {
	return c.msgProto.ClearMessage(ctx, peerID)
}
