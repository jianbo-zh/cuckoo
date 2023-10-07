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

	return convertMessage(msg), nil
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
		peerMsgs = append(peerMsgs, *convertMessage(msg))
	}

	return peerMsgs, nil
}

func (c *ContactSvc) SendMessage(ctx context.Context, peerID peer.ID, msgType string, mimeType string, payload []byte, resources []string) (<-chan mytype.ContactMessage, error) {

	resultCh := make(chan mytype.ContactMessage, 1)

	msg, err := c.msgProto.GenerateMessage(ctx, peerID, msgType, mimeType, payload)
	if err != nil {
		return nil, fmt.Errorf("generate message error: %w", err)
	}
	fmt.Println("msg: ", msg.String())

	resultCh <- *convertMessage(msg)

	go func(msgID string) {
		defer func() {
			close(resultCh)
		}()

		isSucc := true
		if err := c.sendMessage(ctx, peerID, msgID, resources); err != nil {
			isSucc = false
			// log error
			log.Error("send message error: %w", err)
		}

		pbmsg, err := c.msgProto.UpdateMessageState(ctx, peerID, msgID, isSucc)
		if err != nil {
			// log error
			log.Errorf("msgProto.UpdateMessageState error: %w", err)
			return
		}
		resultCh <- *convertMessage(pbmsg)

	}(msg.Id)

	return resultCh, nil
}

func (c *ContactSvc) sendMessage(ctx context.Context, peerID peer.ID, msgID string, resources []string) error {
	if len(resources) > 0 {
		for _, fileID := range resources {
			resultCh := make(chan error, 1)
			if err := c.emitters.evtUploadResource.Emit(myevent.EvtUploadResource{
				ToPeerID: peerID,
				GroupID:  "",
				FileID:   fileID,
				Result:   resultCh,
			}); err != nil {
				return fmt.Errorf("emit evtUploadResource error: %w", err)
			}

			if err := <-resultCh; err != nil {
				return fmt.Errorf("send attahment %s, error: %w", fileID, err)
			}
		}
		fmt.Println("attachment send finish")
	}

	if err := c.msgProto.SendMessage(ctx, peerID, msgID); err != nil {
		fmt.Println("proto.SendMessage error: %w", err)
		return fmt.Errorf("msgProto.SendMessage error: %w", err)
	}

	fmt.Println("sendMessage succ: ", msgID)
	return nil
}

func (c *ContactSvc) UpdateMessageState(ctx context.Context, peerID peer.ID, msgID string, isSucc bool) (*mytype.ContactMessage, error) {
	msg, err := c.msgProto.UpdateMessageState(ctx, peerID, msgID, isSucc)
	if err != nil {
		return nil, fmt.Errorf("msgProto.UpdateMessageState error: %w", err)
	}

	return convertMessage(msg), nil
}

func (c *ContactSvc) ClearMessage(ctx context.Context, peerID peer.ID) error {
	return c.msgProto.ClearMessage(ctx, peerID)
}
