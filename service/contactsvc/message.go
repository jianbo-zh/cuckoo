package contactsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/jianbo-zh/dchat/internal/myerror"
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

func (c *ContactSvc) SendMessage(ctx context.Context, contactID peer.ID, msgType string, mimeType string, payload []byte,
	attachmentID string, file *mytype.FileInfo) (<-chan mytype.ContactMessage, error) {

	// 处理资源文件
	if attachmentID != "" || file != nil {
		resultCh := make(chan error)
		sessionID := mytype.ContactSessionID(contactID)
		if err := c.emitters.evtLogSessionAttachment.Emit(myevent.EvtLogSessionAttachment{
			SessionID:  sessionID.String(),
			ResourceID: attachmentID,
			File:       file,
			Result:     resultCh,
		}); err != nil {
			return nil, fmt.Errorf("emit record session attachment error: %w", err)
		}
		if err := <-resultCh; err != nil {
			return nil, fmt.Errorf("record session attachment error: %w", err)
		}
	}

	// 创建消息
	msg, err := c.msgProto.CreateMessage(ctx, contactID, msgType, mimeType, payload, attachmentID)
	if err != nil {
		return nil, fmt.Errorf("generate message error: %w", err)
	}

	// 发送消息
	msgCh := make(chan mytype.ContactMessage, 1)
	msgCh <- *convertMessage(msg)

	go func(msgID string) {
		defer func() {
			close(msgCh)
		}()

		isSucc := true
		isDeposit, err := c.sendMessage(ctx, contactID, msgID)
		if err != nil {
			isSucc = false
			log.Error("send message error: %w", err)
		}

		msg, err := c.msgProto.UpdateMessageState(ctx, contactID, msgID, isDeposit, isSucc)
		if err != nil {
			// log error
			log.Errorf("msgProto.UpdateMessageState error: %w", err)
			return
		}
		msgCh <- *convertMessage(msg)

	}(msg.Id)

	return msgCh, nil
}

func (c *ContactSvc) sendMessage(ctx context.Context, contactID peer.ID, msgID string) (isDeposit bool, err error) {

	if msgData, err1 := c.msgProto.SendMessage(ctx, contactID, msgID); err1 != nil {
		// 发送失败
		if errors.As(err1, &myerror.StreamErr{}) && len(msgData) > 0 {
			// 可能对方不在线
			if account, err2 := c.accountGetter.GetAccount(ctx); err2 != nil {
				return false, fmt.Errorf("get account error: %w", err2)

			} else if account.AutoDepositMessage {
				// 开启了消息自动寄存
				if contact, err3 := c.contactProto.GetContact(ctx, contactID); err3 != nil {
					return false, fmt.Errorf("proto.GetContact error: %w", err3)

				} else if contact.DepositAddress != peer.ID("") {
					// 对方设置了寄存地址
					resultCh := make(chan error, 1)
					if err4 := c.emitters.evtPushDepositContactMessage.Emit(myevent.EvtPushDepositContactMessage{
						DepositAddress: contact.DepositAddress,
						ToPeerID:       contact.ID,
						MsgID:          msgID,
						MsgData:        msgData,
						Result:         resultCh,
					}); err4 != nil {
						return false, fmt.Errorf("emit EvtPushDepositContactMessage error: %w", err4)
					}

					if err5 := <-resultCh; err5 != nil {
						// 发送寄存信息失败
						return false, fmt.Errorf("push deposit msg error: %w", err5)

					} else {
						return true, nil
					}
				}
			}
		}

		return false, fmt.Errorf("msgProto.SendMessage error: %w", err1)
	}

	return false, nil
}

func (c *ContactSvc) ClearMessage(ctx context.Context, peerID peer.ID) error {
	return c.msgProto.ClearMessage(ctx, peerID)
}
