package contactsvc

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (c *ContactSvc) GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*mytype.ContactMessage, error) {
	msg, err := c.msgProto.GetMessage(ctx, peerID, msgID)
	if err != nil {
		return nil, fmt.Errorf("msgSvc.GetMessage error: %w", err)
	}

	return &mytype.ContactMessage{
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

func (c *ContactSvc) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]mytype.ContactMessage, error) {
	var peerMsgs []mytype.ContactMessage

	msgs, err := c.msgProto.GetMessages(ctx, peerID, offset, limit)
	if err != nil {
		return peerMsgs, err
	}

	for _, msg := range msgs {
		peerMsgs = append(peerMsgs, mytype.ContactMessage{
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

func (c *ContactSvc) SendMessage(ctx context.Context, peerID peer.ID, msgType string, mimeType string, payload []byte, attachments []string) (string, error) {

	if len(attachments) > 0 {
		fmt.Println("attachments: ", len(attachments))
		for _, fileID := range attachments {
			fmt.Println("attachment: ", fileID)
			resultCh := make(chan error, 1)

			if err := c.emitters.evtUploadResource.Emit(myevent.EvtUploadResource{
				ToPeerID: peerID,
				GroupID:  "",
				FileID:   fileID,
				Result:   resultCh,
			}); err != nil {
				return "", fmt.Errorf("emit evtUploadResource error: %w", err)
			}

			if err := <-resultCh; err != nil {
				close(resultCh)
				return "", fmt.Errorf("send attahment %s, error: %w", fileID, err)
			}
			close(resultCh)
		}
		fmt.Println("attachment send finish")
	}

	msgID, err := c.msgProto.SendMessage(ctx, peerID, msgType, mimeType, payload)
	if err != nil {
		fmt.Println("proto.SendMessage error: %w", err)
		return "", fmt.Errorf("msgProto.SendMessage error: %w", err)
	}

	fmt.Println("sendMessage succ: ", msgID)

	return msgID, nil
}

func (c *ContactSvc) ClearMessage(ctx context.Context, peerID peer.ID) error {
	return c.msgProto.ClearMessage(ctx, peerID)
}
