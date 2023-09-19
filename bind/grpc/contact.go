package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/contactsvc"
	"github.com/libp2p/go-libp2p/core/peer"
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

	contactSvc, err := c.getContactSvc()
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

	reply = &proto.GetContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Contact: &proto.Contact{
			Id:     contact.ID.String(),
			Name:   contact.Name,
			Avatar: contact.Avatar,
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

	contactSvc, err := c.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("s.getContactSvc error: %w", err)
	}

	contacts, err := contactSvc.GetContacts(ctx)
	if err != nil {
		return nil, fmt.Errorf("contactSvc.GetContact error: %w", err)
	}

	var contactList []*proto.Contact
	for _, contact := range contacts {
		contactList = append(contactList, &proto.Contact{
			Id:     contact.ID.String(),
			Name:   contact.Name,
			Avatar: contact.Avatar,
		})
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

	contactSvc, err := c.getContactSvc()
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
	for _, contact := range contacts {
		contactList = append(contactList, &proto.Contact{
			Id:     contact.ID.String(),
			Name:   contact.Name,
			Avatar: contact.Avatar,
		})
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

			err = accountSvc.DownloadPeerAvatar(ctx, peerID, peerAccount.Avatar)
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

	contactSvc, err := c.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get contact svc error: %w", err)
	}

	accountSvc, err := c.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("get account svc error: %w", err)
	}

	peerID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
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
			State:       "",
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

	contactSvc, err := c.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get contact service error: %w", err)
	}

	accountSvc, err := c.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("get account service error: %w", err)
	}

	peerID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
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
			State:       "",
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

// todo: 客户端和后端借口不一致，需要调整下
func (c *ContactSvc) SendContactMessage(ctx context.Context, request *proto.SendContactMessageRequest) (reply *proto.SendContactMessageReply, err error) {

	log.Infoln("SendContactMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendContactMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SendContactMessage error: ", err.Error())
		} else {
			log.Infoln("SendContactMessage reply: ", reply.String())
		}
	}()

	contactSvc, err := c.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get contact service error: %w", err)
	}

	accountSvc, err := c.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("get account service error: %w", err)
	}

	peerID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
	}

	contact, err := contactSvc.GetContact(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("get contact error: %w", err)
	}

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account error: %w", err)
	}

	err = contactSvc.SendMessage(ctx, peerID, decodeMsgType(request.MsgType), request.MimeType, request.Payload)
	if err != nil {
		return nil, fmt.Errorf("svc send message error: %w", err)
	}

	reply = &proto.SendContactMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Message: &proto.ContactMessage{
			Id: "id",
			FromContact: &proto.Contact{
				Id:     account.ID.String(),
				Name:   account.Name,
				Avatar: account.Avatar,
			},
			ToContactId: contact.ID.String(),
			MsgType:     request.MsgType,
			MimeType:    request.MimeType,
			Payload:     request.Payload,
			State:       "sending",
			CreateTime:  time.Now().Unix(),
		},
	}
	return reply, nil
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

	peer0 := &types.Peer{
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
