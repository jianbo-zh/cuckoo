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
			log.Error("send message error: %v", err)
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

func (c *ContactSvc) checkDepositAddr(ctx context.Context, contactID peer.ID) (peer.ID, error) {
	if account, err := c.accountGetter.GetAccount(ctx); err != nil {
		return peer.ID(""), fmt.Errorf("get account error: %w", err)

	} else if !account.AutoDepositMessage {
		return peer.ID(""), nil
	}

	contact, err := c.contactProto.GetContact(ctx, contactID)
	if err != nil {
		return peer.ID(""), fmt.Errorf("proto.GetContact error: %w", err)

	}
	return contact.DepositAddress, nil
}

func (c *ContactSvc) sendMessage(ctx context.Context, contactID peer.ID, msgID string) (isDeposit bool, err error) {

	onlineState := c.host.OnlineState(contactID)

	switch onlineState {
	case mytype.OnlineStateOnline, mytype.OnlineStateUnknown:
		// 在线消息
		msgData, err := c.msgProto.SendMessage(ctx, contactID, msgID)
		if err != nil {
			// 在线消息失败
			if errors.As(err, &myerror.StreamErr{}) && msgData != nil {
				streamErr := err
				// 流错误，可能是不在线
				depositAddr, err := c.checkDepositAddr(ctx, contactID)
				if err != nil {
					return false, fmt.Errorf("check deposit addr error: %w", err)
				}

				if depositAddr == peer.ID("") {
					return false, fmt.Errorf("send message failed: %w", streamErr)
				}

				// 发送寄存信息
				resultCh := make(chan error, 1)
				if err := c.emitters.evtPushDepositContactMessage.Emit(myevent.EvtPushDepositContactMessage{
					DepositAddress: depositAddr,
					ToPeerID:       contactID,
					MsgID:          msgID,
					MsgData:        msgData,
					Result:         resultCh,
				}); err != nil {
					return false, fmt.Errorf("emit EvtPushDepositContactMessage error: %w", err)
				}

				if err := <-resultCh; err != nil {
					// 发送寄存信息失败
					return false, fmt.Errorf("push deposit msg error: %w", err)
				}

				return true, nil

			} else {
				return false, fmt.Errorf("proto.SendMessage error: %w", err)
			}
		}

		return false, nil

	case mytype.OnlineStateOffline:
		depositAddr, err := c.checkDepositAddr(ctx, contactID)
		if err != nil {
			return false, fmt.Errorf("check deposit addr error: %w", err)
		}

		if depositAddr == peer.ID("") {
			return false, fmt.Errorf("peer offline")
		}

		msgData, err := c.msgProto.GetMessageEnvelope(ctx, contactID, msgID)
		if err != nil {
			return false, fmt.Errorf("proto.GetMessageEnvelope error: %w", err)
		}

		// 发送寄存信息
		resultCh := make(chan error, 1)
		if err := c.emitters.evtPushDepositContactMessage.Emit(myevent.EvtPushDepositContactMessage{
			DepositAddress: depositAddr,
			ToPeerID:       contactID,
			MsgID:          msgID,
			MsgData:        msgData,
			Result:         resultCh,
		}); err != nil {
			return false, fmt.Errorf("emit EvtPushDepositContactMessage error: %w", err)
		}

		if err := <-resultCh; err != nil {
			// 发送寄存信息失败
			return false, fmt.Errorf("push deposit msg error: %w", err)
		}

		return true, nil

	default:
		return false, fmt.Errorf("unsupport online state")
	}
}

func (c *ContactSvc) ClearMessage(ctx context.Context, contactID peer.ID) error {

	if err := c.msgProto.ClearMessage(ctx, contactID); err != nil {
		return fmt.Errorf("proto.ClearMessage error: %w", err)
	}

	sessionID := mytype.ContactSessionID(contactID)

	// 清空会话
	resultCh := make(chan error, 1)
	if err := c.emitters.evtClearSession.Emit(myevent.EvtClearSession{
		SessionID: sessionID.String(),
		Result:    resultCh,
	}); err != nil {
		return fmt.Errorf("emit clear session resources error: %w", err)
	}
	if err := <-resultCh; err != nil {
		return fmt.Errorf("clear session resources error: %w", err)
	}

	// 删除会话资源
	resultCh1 := make(chan error, 1)
	if err := c.emitters.evtClearSessionResources.Emit(myevent.EvtClearSessionResources{
		SessionID: sessionID.String(),
		Result:    resultCh1,
	}); err != nil {
		return fmt.Errorf("emit clear session resources error: %w", err)
	}
	if err := <-resultCh1; err != nil {
		return fmt.Errorf("clear session resources error: %w", err)
	}

	return nil
}
