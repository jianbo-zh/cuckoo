package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
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

func (c *ContactSvc) ClearContactMessage(ctx context.Context, request *proto.ClearContactMessageRequest) (*proto.ClearContactMessageReply, error) {

	contactMessages = make([]*proto.ContactMessage, 0)

	reply := &proto.ClearContactMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (c *ContactSvc) DeleteContact(ctx context.Context, request *proto.DeleteContactRequest) (*proto.DeleteContactReply, error) {

	contacts = make([]*proto.Contact, 0)

	reply := &proto.DeleteContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (c *ContactSvc) GetContact(ctx context.Context, request *proto.GetContactRequest) (*proto.GetContactReply, error) {

	var contact *proto.Contact = nil

	if len(contacts) > 0 {
		contact = contacts[0]
	}

	reply := &proto.GetContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Contact: contact,
	}
	return reply, nil
}

func (c *ContactSvc) GetContacts(ctx context.Context, request *proto.GetContactsRequest) (*proto.GetContactsReply, error) {

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
			Id:     contact.PeerID.String(),
			Name:   contact.Name,
			Avatar: contact.Avatar,
		})
	}

	reply := &proto.GetContactsReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Contacts: contactList,
	}
	return reply, nil
}

func (c *ContactSvc) GetContactIDs(ctx context.Context, request *proto.GetContactIdsRequest) (*proto.GetContactIdsReply, error) {

	var peerIDs []string
	for _, contact := range contacts {
		peerIDs = append(peerIDs, contact.Id)
	}

	reply := &proto.GetContactIdsReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		ContactIds: peerIDs,
	}
	return reply, nil
}

func (c *ContactSvc) GetSpecifiedContacts(ctx context.Context, request *proto.GetSpecifiedContactsRequest) (*proto.GetSpecifiedContactsReply, error) {

	contactSvc, err := c.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("s.getContactSvc error: %w", err)
	}

	var peerIDs []peer.ID
	for _, pid := range request.ContactIds {
		peerID, err := peer.Decode(pid)
		if err != nil {
			return nil, fmt.Errorf("peer.Decode error: %w", err)
		}
		peerIDs = append(peerIDs, peerID)
	}

	contacts, err := contactSvc.GetContactsByPeerIDs(ctx, peerIDs)
	if err != nil {
		return nil, fmt.Errorf("contactSvc.GetContactsByPeerIDs error: %w", err)
	}

	contactList := make([]*proto.Contact, 0)
	for _, contact := range contacts {
		contactList = append(contactList, &proto.Contact{
			Id:     contact.PeerID.String(),
			Name:   contact.Name,
			Avatar: contact.Avatar,
		})
	}
	reply := &proto.GetSpecifiedContactsReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Contacts: contactList,
	}
	return reply, nil
}

func (c *ContactSvc) GetNearbyContacts(request *proto.GetNearbyContactsRequest, server proto.ContactSvc_GetNearbyContactsServer) error {
	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return fmt.Errorf("c.getter.GetCuckoo error: %w", err)
	}

	peerIDs, err := cuckoo.GetLanPeerIDs()
	if err != nil {
		return fmt.Errorf("cuckoo.GetLanPeerIDs error: %w", err)
	}

	fmt.Printf("get lan peer ids: %d\n", len(peerIDs))

	if len(peerIDs) > 0 {
		accountSvc, err := cuckoo.GetAccountSvc()
		if err != nil {
			return fmt.Errorf("cuckoo.GetAccountSvc error: %w", err)
		}

		ctx := context.Background()
		for _, peerID := range peerIDs {
			fmt.Printf("get lan peer: %s\n", peerID.String())

			peerAccount, err := accountSvc.GetPeer(ctx, peerID)
			if err != nil {
				fmt.Printf("accountSvc.GetPeerAccount error: %s", err.Error())
				continue
			}

			fmt.Println("download peer avatar start")
			err = accountSvc.DownloadPeerAvatar(ctx, peerID, peerAccount.Avatar)
			if err != nil {
				fmt.Printf("accountSvc.DownloadPeerAvatar error: %s\n", err.Error())
				continue
			}
			fmt.Println("download peer avatar end")

			server.Send(&proto.GetNearbyContactsStreamReply{
				Result: &proto.Result{
					Code:    0,
					Message: "ok",
				},
				Contact: &proto.Contact{
					Id:     peerAccount.PeerID.String(),
					Name:   peerAccount.Name,
					Avatar: peerAccount.Avatar,
				},
			})
			fmt.Println("server.Sended nearby")
		}
	}

	return nil
}

func (c *ContactSvc) GetContactMessage(ctx context.Context, request *proto.GetContactMessageRequest) (*proto.GetContactMessageReply, error) {
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

	msg, err := contactSvc.GetMessage(ctx, peerID, request.MsgId)
	if err != nil {
		return nil, fmt.Errorf("contactSvc.GetMessage error: %w", err)
	}

	var sender proto.Contact
	var receiver proto.Contact
	if msg.FromPeerID == account.PeerID {
		sender = proto.Contact{
			Id:     account.PeerID.String(),
			Name:   account.Name,
			Avatar: account.Avatar,
		}
		receiver = proto.Contact{
			Id:     contact.PeerID.String(),
			Name:   contact.Name,
			Avatar: contact.Avatar,
		}
	} else if msg.FromPeerID == contact.PeerID {
		sender = proto.Contact{
			Id:     contact.PeerID.String(),
			Name:   contact.Name,
			Avatar: contact.Avatar,
		}
		receiver = proto.Contact{
			Id:     account.PeerID.String(),
			Name:   account.Name,
			Avatar: account.Avatar,
		}
	}

	return &proto.GetContactMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "",
		},
		Message: &proto.ContactMessage{
			Id:         msg.ID,
			FromPeer:   &sender,
			ToPeerId:   receiver.Id,
			MsgType:    "text",
			MimeType:   msg.MimeType,
			Payload:    msg.Payload,
			State:      "",
			CreateTime: msg.Timestamp,
		},
	}, nil
}

func (c *ContactSvc) GetContactMessages(ctx context.Context, request *proto.GetContactMessagesRequest) (*proto.GetContactMessagesReply, error) {

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
		var receiver proto.Contact
		if msg.FromPeerID == account.PeerID {
			sender = proto.Contact{
				Id:     account.PeerID.String(),
				Name:   account.Name,
				Avatar: account.Avatar,
			}
			receiver = proto.Contact{
				Id:     contact.PeerID.String(),
				Name:   contact.Name,
				Avatar: contact.Avatar,
			}
		} else if msg.FromPeerID == contact.PeerID {
			sender = proto.Contact{
				Id:     contact.PeerID.String(),
				Name:   contact.Name,
				Avatar: contact.Avatar,
			}
			receiver = proto.Contact{
				Id:     account.PeerID.String(),
				Name:   account.Name,
				Avatar: account.Avatar,
			}
		}

		msglist = append(msglist, &proto.ContactMessage{
			Id:         msg.ID,
			FromPeer:   &sender,
			ToPeerId:   receiver.Id,
			MsgType:    "text",
			MimeType:   msg.MimeType,
			Payload:    msg.Payload,
			State:      "",
			CreateTime: msg.Timestamp,
		})
	}

	reply := &proto.GetContactMessagesReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Messages: msglist,
	}

	return reply, nil
}

// todo: 客户端和后端借口不一致，需要调整下
func (c *ContactSvc) SendContactMessage(ctx context.Context, request *proto.SendContactMessageRequest) (*proto.SendContactMessageReply, error) {

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

	err = contactSvc.SendTextMessage(ctx, peerID, string(request.Data))
	if err != nil {
		return nil, fmt.Errorf("send text message error: %w", err)
	}

	reply := &proto.SendContactMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Message: &proto.ContactMessage{
			Id: "id",
			FromPeer: &proto.Contact{
				Id:     account.PeerID.String(),
				Name:   account.Name,
				Avatar: account.Avatar,
			},
			ToPeerId:   contact.PeerID.String(),
			MsgType:    request.MsgType,
			MimeType:   request.MimeType,
			Payload:    request.Data,
			State:      "sending",
			CreateTime: time.Now().Unix(),
		},
	}
	return reply, nil
}

func (c *ContactSvc) SetContactName(ctx context.Context, request *proto.SetContactNameRequest) (*proto.SetContactNameReply, error) {

	if len(contacts) > 0 {
		for i, contact := range contacts {
			if contact.Id == request.ContactId {
				contacts[i].Name = request.Name
			}
		}
	}

	reply := &proto.SetContactNameReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Name: request.Name,
	}
	return reply, nil
}
