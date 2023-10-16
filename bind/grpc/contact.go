package service

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/contactsvc"
	"github.com/jianbo-zh/dchat/service/depositsvc"
	"github.com/jianbo-zh/dchat/service/filesvc"
	"github.com/libp2p/go-libp2p/core/peer"
	goproto "google.golang.org/protobuf/proto"
)

var _ proto.ContactSvcServer = (*ContactSvc)(nil)

type ContactSvc struct {
	getter cuckoo.CuckooGetter
	proto.UnimplementedContactSvcServer
}

func NewContactSvc(getter cuckoo.CuckooGetter) *ContactSvc {
	return &ContactSvc{
		getter: getter,
	}
}

func (c *ContactSvc) getContactSvc() (contactsvc.ContactServiceIface, error) {
	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return contactSvc, nil
}

func (c *ContactSvc) getAccountSvc() (accountsvc.AccountServiceIface, error) {
	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	accountSvc, err := cuckoo.GetAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return accountSvc, nil
}

func (c *ContactSvc) getFileSvc() (filesvc.FileServiceIface, error) {
	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	fileSvc, err := cuckoo.GetFileSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return fileSvc, nil
}

func (c *ContactSvc) getDepositSvc() (depositsvc.DepositServiceIface, error) {
	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	depositSvc, err := cuckoo.GetDepositSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return depositSvc, nil
}

func (c *ContactSvc) GetContact(ctx context.Context, request *proto.GetContactRequest) (reply *proto.GetContactReply, err error) {

	log.Infoln("GetContact request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetContact panic: ", e)
		} else if err != nil {
			log.Errorln("GetContact error: ", err.Error())
		} else {
			log.Infoln("GetContact reply: ", reply.String())
		}
	}()

	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return nil, fmt.Errorf("s.getContactSvc error: %w", err)
	}

	peerID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
	}

	contact, err := contactSvc.GetContact(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("svc get contact error: %w", err)
	}

	accountSvc, err := cuckoo.GetAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("get account svc error: %w", err)
	}

	onlineStateMap := accountSvc.GetOnlineState([]peer.ID{peerID})

	reply = &proto.GetContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Contact: &proto.Contact{
			Id:             contact.ID.String(),
			Name:           contact.Name,
			Avatar:         contact.Avatar,
			DepositAddress: contact.DepositAddress.String(),
			OnlineState:    encodeOnlineState(onlineStateMap[contact.ID]),
		},
	}
	return reply, nil
}

func (c *ContactSvc) GetContacts(ctx context.Context, request *proto.GetContactsRequest) (reply *proto.GetContactsReply, err error) {

	log.Infoln("GetContacts request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetContacts panic: ", e)
		} else if err != nil {
			log.Errorln("GetContacts error: ", err.Error())
		} else {
			log.Infoln("GetContacts reply: ", reply.String())
		}
	}()

	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return nil, fmt.Errorf("s.getContactSvc error: %w", err)
	}

	contacts, err := contactSvc.GetContacts(ctx)
	if err != nil {
		return nil, fmt.Errorf("contactSvc.GetContact error: %w", err)
	}

	var contactList []*proto.Contact

	if len(contacts) > 0 {
		var peerIDs []peer.ID
		for _, contact := range contacts {
			peerIDs = append(peerIDs, contact.ID)
		}

		accountSvc, err := cuckoo.GetAccountSvc()
		if err != nil {
			return nil, fmt.Errorf("get account svc error: %w", err)
		}

		onlineStateMap := accountSvc.GetOnlineState(peerIDs)

		for _, contact := range contacts {
			contactList = append(contactList, &proto.Contact{
				Id:             contact.ID.String(),
				Name:           contact.Name,
				Avatar:         contact.Avatar,
				DepositAddress: contact.DepositAddress.String(),
				OnlineState:    encodeOnlineState(onlineStateMap[contact.ID]),
			})
		}
	}

	reply = &proto.GetContactsReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Contacts: contactList,
	}
	return reply, nil
}

func (c *ContactSvc) GetSpecifiedContacts(ctx context.Context, request *proto.GetSpecifiedContactsRequest) (reply *proto.GetSpecifiedContactsReply, err error) {

	log.Infoln("GetSpecifiedContacts request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetSpecifiedContacts panic: ", e)
		} else if err != nil {
			log.Errorln("GetSpecifiedContacts error: ", err.Error())
		} else {
			log.Infoln("GetSpecifiedContacts reply: ", reply.String())
		}
	}()

	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get contact svc error: %w", err)
	}

	var peerIDs []peer.ID
	for _, pid := range request.ContactIds {
		peerID, err := peer.Decode(pid)
		if err != nil {
			return nil, fmt.Errorf("peer decode error: %w", err)
		}
		peerIDs = append(peerIDs, peerID)
	}

	contacts, err := contactSvc.GetContactsByPeerIDs(ctx, peerIDs)
	if err != nil {
		return nil, fmt.Errorf("svc get contacts by peer ids error: %w", err)
	}

	contactList := make([]*proto.Contact, 0)

	if len(contacts) > 0 {

		var peerIDs []peer.ID
		for _, contact := range contacts {
			peerIDs = append(peerIDs, contact.ID)
		}

		accountSvc, err := cuckoo.GetAccountSvc()
		if err != nil {
			return nil, fmt.Errorf("get account svc error: %w", err)
		}

		onlineStateMap := accountSvc.GetOnlineState(peerIDs)

		for _, contact := range contacts {
			contactList = append(contactList, &proto.Contact{
				Id:             contact.ID.String(),
				Name:           contact.Name,
				Avatar:         contact.Avatar,
				DepositAddress: contact.DepositAddress.String(),
				OnlineState:    encodeOnlineState(onlineStateMap[contact.ID]),
			})
		}
	}

	reply = &proto.GetSpecifiedContactsReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Contacts: contactList,
	}
	return reply, nil
}

func (c *ContactSvc) DeleteContact(ctx context.Context, request *proto.DeleteContactRequest) (reply *proto.DeleteContactReply, err error) {

	log.Infoln("DeleteContact request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("DeleteContact panic: ", e)
		} else if err != nil {
			log.Errorln("DeleteContact error: ", err.Error())
		} else {
			log.Infoln("DeleteContact reply: ", reply.String())
		}
	}()

	contactSvc, err := c.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get contact svc error: %w", err)
	}

	peerID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
	}

	if err = contactSvc.DeleteContact(ctx, peerID); err != nil {
		return nil, fmt.Errorf("svc delete contact error: %w", err)
	}

	reply = &proto.DeleteContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (c *ContactSvc) GetNearbyPeers(request *proto.GetNearbyPeersRequest, server proto.ContactSvc_GetNearbyPeersServer) (err error) {

	log.Infoln("GetNearbyPeers request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetNearbyPeers panic: ", e)
		} else if err != nil {
			log.Errorln("GetNearbyPeers error: ", err.Error())
		}
	}()

	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return fmt.Errorf("get cuckoo error: %w", err)
	}

	peerIDs, err := cuckoo.GetLanPeerIDs()
	if err != nil {
		return fmt.Errorf("cuckoo get lan peer ids error: %w", err)
	}

	if len(peerIDs) > 0 {
		accountSvc, err := cuckoo.GetAccountSvc()
		if err != nil {
			return fmt.Errorf("get account svc error: %w", err)
		}

		ctx := context.Background()
		for _, peerID := range peerIDs {

			peerAccount, err := accountSvc.GetPeer(ctx, peerID)
			if err != nil {
				log.Errorf("svc get peer error: %s", err.Error())
				continue
			}

			fileSvc, err := cuckoo.GetFileSvc()
			if err != nil {
				return fmt.Errorf("get account svc error: %w", err)
			}

			err = fileSvc.DownloadAvatar(ctx, peerID, peerAccount.Avatar)
			if err != nil {
				log.Errorf("svc download peer avatar error: %s", err.Error())
				continue
			}

			reply := &proto.GetNearbyPeersStreamReply{
				Result: &proto.Result{
					Code:    0,
					Message: "ok",
				},
				Peer: &proto.Peer{
					Id:     peerAccount.ID.String(),
					Name:   peerAccount.Name,
					Avatar: peerAccount.Avatar,
				},
			}
			if err = server.Send(reply); err != nil {
				return err
			}

			log.Infoln("GetNearbyContacts reply: ", reply.String())
		}
	}

	return nil
}

func (c *ContactSvc) GetContactMessage(ctx context.Context, request *proto.GetContactMessageRequest) (reply *proto.GetContactMessageReply, err error) {

	log.Infoln("GetContactMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetContactMessage panic: ", e)
		} else if err != nil {
			log.Errorln("GetContactMessage error: ", err.Error())
		} else {
			log.Infoln("GetContactMessage reply: ", reply.String())
		}
	}()

	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get contact svc error: %w", err)
	}

	accountSvc, err := cuckoo.GetAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("get account svc error: %w", err)
	}

	sessionSvc, err := cuckoo.GetSessionSvc()
	if err != nil {
		return nil, fmt.Errorf("get session svc error: %w", err)
	}

	peerID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
	}

	sessionID := mytype.ContactSessionID(peerID)
	if err = sessionSvc.ResetUnreads(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("svc reset unreads error: %w", err)
	}
	if err = sessionSvc.UpdateSessionTime(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("svc update session time error: %w", err)
	}

	contact, err := contactSvc.GetContact(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("svc get contact error: %w", err)
	}

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get account error: %w", err)
	}

	msg, err := contactSvc.GetMessage(ctx, peerID, request.MsgId)
	if err != nil {
		return nil, fmt.Errorf("svc get msg error: %w", err)
	}

	var sender proto.Contact
	var receiverID string
	if msg.FromPeerID == account.ID {
		sender = proto.Contact{
			Id:     account.ID.String(),
			Name:   account.Name,
			Avatar: account.Avatar,
		}
		receiverID = contact.ID.String()

	} else if msg.FromPeerID == contact.ID {
		sender = proto.Contact{
			Id:     contact.ID.String(),
			Name:   contact.Name,
			Avatar: contact.Avatar,
		}
		receiverID = account.ID.String()
	}

	reply = &proto.GetContactMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "",
		},
		Message: &proto.ContactMessage{
			Id:          msg.ID,
			FromContact: &sender,
			ToContactId: receiverID,
			MsgType:     encodeMsgType(msg.MsgType),
			MimeType:    msg.MimeType,
			Payload:     msg.Payload,
			IsDeposit:   msg.IsDeposit,
			State:       encodeMessageState(msg.State),
			CreateTime:  msg.Timestamp,
		},
	}

	return reply, nil
}

func (c *ContactSvc) GetContactMessages(ctx context.Context, request *proto.GetContactMessagesRequest) (reply *proto.GetContactMessagesReply, err error) {

	log.Infoln("GetContactMessages request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetContactMessages panic: ", e)
		} else if err != nil {
			log.Errorln("GetContactMessages error: ", err.Error())
		} else {
			log.Infoln("GetContactMessages reply: ", reply.String())
		}
	}()

	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get contact service error: %w", err)
	}

	accountSvc, err := cuckoo.GetAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("get account service error: %w", err)
	}

	sessionSvc, err := cuckoo.GetSessionSvc()
	if err != nil {
		return nil, fmt.Errorf("get session svc error: %w", err)
	}

	peerID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
	}

	sessionID := mytype.ContactSessionID(peerID)
	if err = sessionSvc.ResetUnreads(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("svc reset unreads error: %w", err)
	}
	if err = sessionSvc.UpdateSessionTime(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("svc update session time error: %w", err)
	}

	contact, err := contactSvc.GetContact(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("get contact error: %w", err)
	}

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account error: %w", err)
	}

	msgs, err := contactSvc.GetMessages(ctx, peerID, int(request.Offset), int(request.Limit))
	if err != nil {
		return nil, fmt.Errorf("get message error: %w", err)
	}

	var msglist []*proto.ContactMessage
	for _, msg := range msgs {
		var sender proto.Contact
		var receiverID string
		if msg.FromPeerID == account.ID {
			sender = proto.Contact{
				Id:     account.ID.String(),
				Name:   account.Name,
				Avatar: account.Avatar,
			}
			receiverID = contact.ID.String()

		} else if msg.FromPeerID == contact.ID {
			sender = proto.Contact{
				Id:     contact.ID.String(),
				Name:   contact.Name,
				Avatar: contact.Avatar,
			}
			receiverID = account.ID.String()
		}

		msglist = append(msglist, &proto.ContactMessage{
			Id:          msg.ID,
			FromContact: &sender,
			ToContactId: receiverID,
			MsgType:     encodeMsgType(msg.MsgType),
			MimeType:    msg.MimeType,
			Payload:     msg.Payload,
			IsDeposit:   msg.IsDeposit,
			State:       encodeMessageState(msg.State),
			CreateTime:  msg.Timestamp,
		})
	}

	reply = &proto.GetContactMessagesReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Messages: msglist,
	}

	return reply, nil
}

// SendContactTextMessage 发送文本
func (c *ContactSvc) SendContactTextMessage(request *proto.SendContactTextMessageRequest, server proto.ContactSvc_SendContactTextMessageServer) (err error) {

	log.Infoln("SendContactTextMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendContactTextMessage panic: ", e)

		} else if err != nil {
			log.Errorln("SendContactTextMessage error: ", err.Error())
		}
	}()

	return c.sendContactMessage(context.Background(), server, mytype.TextMsgType, request.ContactId, "text/plain", []byte(request.Content), "", nil)
}

// SendContactImageMessage 发送图片
func (c *ContactSvc) SendContactImageMessage(request *proto.SendContactImageMessageRequest, server proto.ContactSvc_SendContactImageMessageServer) (err error) {

	log.Infoln("SendContactImageMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendContactImageMessage panic: ", e)

		} else if err != nil {
			log.Errorln("SendContactImageMessage error: ", err.Error())
		}
	}()

	fileSvc, err := c.getFileSvc()
	if err != nil {
		return fmt.Errorf("get file svc error: %w", err)
	}

	ctx := context.Background()
	thumbnailID, err := fileSvc.CopyFileToResource(ctx, request.ThumbnailPath)
	if err != nil {
		return fmt.Errorf("copy file to resource error: %w", err)
	}

	imageID, err := fileSvc.CopyFileToFile(ctx, request.FilePath)
	if err != nil {
		return fmt.Errorf("copy file to file error: %w", err)
	}

	payload, err := goproto.Marshal(&proto.ImageMessagePayload{
		ThumbnailId: thumbnailID,
		ImageId:     imageID,
		Name:        request.Name,
		Size:        request.Size,
		Width:       request.Width,
		Height:      request.Height,
	})
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	file := mytype.FileInfo{
		FileID:      imageID,
		FileName:    request.Name,
		FileSize:    request.Size,
		FileType:    mytype.ImageFile,
		MimeType:    request.MimeType,
		ThumbnailID: thumbnailID,
		Width:       request.Width,
		Height:      request.Height,
	}

	return c.sendContactMessage(context.Background(), server, mytype.ImageMsgType, request.ContactId, request.MimeType, payload, thumbnailID, &file)
}

// SendContactVoiceMessage 发送语音
func (c *ContactSvc) SendContactVoiceMessage(request *proto.SendContactVoiceMessageRequest, server proto.ContactSvc_SendContactVoiceMessageServer) (err error) {

	log.Infoln("SendContactVoiceMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendContactVoiceMessage panic: ", e)

		} else if err != nil {
			log.Errorln("SendContactVoiceMessage error: ", err.Error())
		}
	}()

	fileSvc, err := c.getFileSvc()
	if err != nil {
		return fmt.Errorf("get file svc error: %w", err)
	}

	ctx := context.Background()
	voiceID, err := fileSvc.CopyFileToResource(ctx, request.FilePath)
	if err != nil {
		return fmt.Errorf("copy file to resource error: %w", err)
	}

	payload, err := goproto.Marshal(&proto.VoiceMessagePayload{
		VoiceId:  voiceID,
		Duration: request.Duration,
	})
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	return c.sendContactMessage(context.Background(), server, mytype.VoiceMsgType, request.ContactId, request.MimeType, payload, voiceID, nil)
}

// SendContactAudioMessage 发送音频
func (c *ContactSvc) SendContactAudioMessage(request *proto.SendContactAudioMessageRequest, server proto.ContactSvc_SendContactAudioMessageServer) (err error) {

	log.Infoln("SendContactAudioMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendContactAudioMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SendContactAudioMessage error: ", err.Error())
		}
	}()

	fileSvc, err := c.getFileSvc()
	if err != nil {
		return fmt.Errorf("get file svc error: %w", err)
	}

	ctx := context.Background()
	audioID, err := fileSvc.CopyFileToFile(ctx, request.FilePath)
	if err != nil {
		return fmt.Errorf("copy file to resource error: %w", err)
	}

	payload, err := goproto.Marshal(&proto.AudioMessagePayload{
		AudioId:  audioID,
		Name:     request.Name,
		Size:     request.Size,
		Duration: request.Duration,
	})
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	file := mytype.FileInfo{
		FileID:   audioID,
		FileName: request.Name,
		FileSize: request.Size,
		FileType: mytype.AudioFile,
		MimeType: request.MimeType,
		Duration: request.Duration,
	}

	return c.sendContactMessage(context.Background(), server, mytype.AudioMsgType, request.ContactId, request.MimeType, payload, "", &file)
}

// SendContactVideoMessage 发送视频
func (c *ContactSvc) SendContactVideoMessage(request *proto.SendContactVideoMessageRequest, server proto.ContactSvc_SendContactVideoMessageServer) (err error) {

	log.Infoln("SendContactVideoMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendContactVideoMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SendContactVideoMessage error: ", err.Error())
		}
	}()

	fileSvc, err := c.getFileSvc()
	if err != nil {
		return fmt.Errorf("get file svc error: %w", err)
	}

	ctx := context.Background()
	videoID, err := fileSvc.CopyFileToFile(ctx, request.FilePath)
	if err != nil {
		return fmt.Errorf("copy file to resource error: %w", err)
	}

	payload, err := goproto.Marshal(&proto.VideoMessagePayload{
		VideoId:  videoID,
		Name:     request.Name,
		Size:     request.Size,
		Duration: request.Duration,
	})
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	file := mytype.FileInfo{
		FileID:   videoID,
		FileName: request.Name,
		FileSize: request.Size,
		FileType: mytype.VideoFile,
		MimeType: request.MimeType,
		Duration: request.Duration,
	}

	return c.sendContactMessage(context.Background(), server, mytype.VideoMsgType, request.ContactId, request.MimeType, payload, "", &file)
}

// SendContactFileMessage 发送文件
func (c *ContactSvc) SendContactFileMessage(request *proto.SendContactFileMessageRequest, server proto.ContactSvc_SendContactFileMessageServer) (err error) {

	log.Infoln("SendContactFileMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendContactFileMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SendContactFileMessage error: ", err.Error())
		}
	}()

	fileSvc, err := c.getFileSvc()
	if err != nil {
		return fmt.Errorf("get file svc error: %w", err)
	}

	ctx := context.Background()
	fileID, err := fileSvc.CopyFileToFile(ctx, request.FilePath)
	if err != nil {
		return fmt.Errorf("copy file to resource error: %w", err)
	}

	payload, err := goproto.Marshal(&proto.FileMessagePayload{
		FileId: fileID,
		Name:   request.Name,
		Size:   request.Size,
	})
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	file := mytype.FileInfo{
		FileID:   fileID,
		FileName: request.Name,
		FileSize: request.Size,
		FileType: mytype.OtherFile,
		MimeType: request.MimeType,
	}

	return c.sendContactMessage(context.Background(), server, mytype.FileMsgType, request.ContactId, request.MimeType, payload, "", &file)
}

func (c *ContactSvc) sendContactMessage(ctx context.Context, server proto.ContactSvc_SendContactTextMessageServer,
	msgType string, contactID string, mimeType string, payload []byte, resourceID string, file *mytype.FileInfo) error {

	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	accountSvc, err := cuckoo.GetAccountSvc()
	if err != nil {
		return fmt.Errorf("get account service error: %w", err)
	}

	peerID, err := peer.Decode(contactID)
	if err != nil {
		return fmt.Errorf("peer decode error: %w", err)
	}

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("get account error: %w", err)
	}

	resultCh, err := contactSvc.SendMessage(ctx, peerID, msgType, mimeType, payload, resourceID, file)
	if err != nil {
		return fmt.Errorf("svc.SendMessage error: %w", err)
	}

	i := 0
	for msg := range resultCh {
		i++
		reply := proto.SendContactMessageReply{
			Result: &proto.Result{
				Code:    0,
				Message: "ok",
			},
			IsUpdated: i > 1,
			Message: &proto.ContactMessage{
				Id: msg.ID,
				FromContact: &proto.Contact{
					Id:             account.ID.String(),
					Name:           account.Name,
					Avatar:         account.Avatar,
					DepositAddress: account.DepositAddress.String(),
					OnlineState:    proto.ConnState_OnlineState,
				},
				ToContactId: msg.ToPeerID.String(),
				MsgType:     encodeMsgType(msg.MsgType),
				MimeType:    msg.MimeType,
				Payload:     msg.Payload,
				IsDeposit:   msg.IsDeposit,
				State:       encodeMessageState(msg.State),
				CreateTime:  msg.Timestamp,
			},
		}
		if err := server.Send(&reply); err != nil {
			return fmt.Errorf("server.Send error: %w", err)
		}

		log.Infoln("sendContactMessage reply: ", reply.String())
	}

	return nil
}

func (c *ContactSvc) ClearContactMessage(ctx context.Context, request *proto.ClearContactMessageRequest) (reply *proto.ClearContactMessageReply, err error) {

	log.Infoln("ClearContactMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("ClearContactMessage panic: ", e)
		} else if err != nil {
			log.Errorln("ClearContactMessage error: ", err.Error())
		} else {
			log.Infoln("ClearContactMessage reply: ", reply.String())
		}
	}()

	contactSvc, err := c.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get contact svc error: %w", err)
	}

	peerID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
	}

	err = contactSvc.ClearMessage(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("svc clear msg error: %w", err)
	}

	reply = &proto.ClearContactMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (c *ContactSvc) SetContactName(ctx context.Context, request *proto.SetContactNameRequest) (reply *proto.SetContactNameReply, err error) {

	log.Infoln("SetContactName request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetContactName panic: ", e)
		} else if err != nil {
			log.Errorln("SetContactName error: ", err.Error())
		} else {
			log.Infoln("SetContactName reply: ", reply.String())
		}
	}()

	contactSvc, err := c.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get contact svc error: %w", err)
	}

	peerID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
	}

	if err = contactSvc.SetContactName(ctx, peerID, request.Name); err != nil {
		return nil, fmt.Errorf("svc set contact name error: %w", err)
	}

	reply = &proto.SetContactNameReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Name: request.Name,
	}
	return reply, nil
}

func (c *ContactSvc) ApplyAddContact(ctx context.Context, request *proto.ApplyAddContactRequest) (reply *proto.ApplyAddContactReply, err error) {

	log.Infoln("ApplyAddContact request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("ApplyAddContact panic: ", e)
		} else if err != nil {
			log.Errorln("ApplyAddContact error: ", err.Error())
		} else {
			log.Infoln("ApplyAddContact reply: ", reply.String())
		}
	}()

	contactSvc, err := c.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("getContactSvc error: %s", err.Error())
	}

	peerID, err := peer.Decode(request.PeerId)
	if err != nil {
		return nil, fmt.Errorf("peer.Decode error: %s", err.Error())
	}

	peer0 := &mytype.Peer{
		ID:     peerID,
		Name:   request.Name,
		Avatar: request.Avatar,
	}

	err = contactSvc.ApplyAddContact(ctx, peer0, request.Content)
	if err != nil {
		return nil, fmt.Errorf("peerSvc.AddPeer error: %w", err)
	}

	reply = &proto.ApplyAddContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}
