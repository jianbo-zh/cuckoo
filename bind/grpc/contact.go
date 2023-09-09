package service

import (
	"context"
	"fmt"
	"strings"
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

func (c *ContactSvc) GetContactList(ctx context.Context, request *proto.GetContactListRequest) (*proto.GetContactListReply, error) {

	var contactList []*proto.Contact
	if request.Keywords == "" {
		contactList = contacts

	} else {
		contactList = make([]*proto.Contact, 0)
		for _, contact := range contacts {
			if strings.Contains(contact.Name, request.Keywords) {
				contactList = append(contactList, contact)
			}
		}
	}

	reply := &proto.GetContactListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		ContactList: contactList,
	}
	return reply, nil
}

func (c *ContactSvc) GetContactIDs(ctx context.Context, request *proto.GetContactIDsRequest) (*proto.GetContactIDsReply, error) {

	var peerIDs []string
	for _, contact := range contacts {
		peerIDs = append(peerIDs, contact.PeerID)
	}

	reply := &proto.GetContactIDsReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		PeerIDs: peerIDs,
	}
	return reply, nil
}

func (c *ContactSvc) GetSpecifiedContactList(ctx context.Context, request *proto.GetSpecifiedContactListRequest) (*proto.GetSpecifiedContactListReply, error) {

	contactList := make([]*proto.Contact, 0)
	if len(request.PeerIDs) > 0 {
		peersMap := make(map[string]struct{})
		for _, peerID := range request.PeerIDs {
			peersMap[peerID] = struct{}{}
		}

		for _, contact := range contacts {
			if _, exists := peersMap[contact.PeerID]; exists {
				contactList = append(contactList, contact)
			}
		}
	}

	reply := &proto.GetSpecifiedContactListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		ContactList: contactList,
	}
	return reply, nil
}

func (c *ContactSvc) GetNearbyContactList(request *proto.GetNearbyContactListRequest, server proto.ContactSvc_GetNearbyContactListServer) error {
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

			server.Send(&proto.GetNearbyContactListStreamReply{
				Result: &proto.Result{
					Code:    0,
					Message: "ok",
				},
				Contact: &proto.Contact{
					PeerID:      peerAccount.PeerID.String(),
					Name:        peerAccount.Name,
					Avatar:      peerAccount.Avatar,
					Alias:       "",
					LastMessage: "",
					UpdateTime:  time.Now().Unix(),
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

	peerID, err := peer.Decode(request.PeerID)
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

	msg, err := contactSvc.GetMessage(ctx, peerID, request.MsgID)
	if err != nil {
		return nil, fmt.Errorf("contactSvc.GetMessage error: %w", err)
	}

	var sender proto.Contact
	var receiver proto.Contact
	if msg.FromPeerID == account.PeerID {
		sender = proto.Contact{
			PeerID: account.PeerID.String(),
			Name:   account.Name,
			Avatar: account.Avatar,
		}
		receiver = proto.Contact{
			PeerID: contact.PeerID.String(),
			Name:   contact.Name,
			Avatar: contact.Avatar,
		}
	} else if msg.FromPeerID == contact.PeerID {
		sender = proto.Contact{
			PeerID: contact.PeerID.String(),
			Name:   contact.Name,
			Avatar: contact.Avatar,
		}
		receiver = proto.Contact{
			PeerID: account.PeerID.String(),
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
			ID:         msg.ID,
			Sender:     &sender,
			Receiver:   &receiver,
			MsgType:    proto.MsgType_TEXT_MSG,
			MimeType:   msg.MimeType,
			Data:       msg.Payload,
			CreateTime: msg.Timestamp,
			MsgState:   "",
		},
	}, nil
}

func (c *ContactSvc) GetContactMessageList(ctx context.Context, request *proto.GetContactMessageListRequest) (*proto.GetContactMessageListReply, error) {

	contactSvc, err := c.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get contact service error: %w", err)
	}

	accountSvc, err := c.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("get account service error: %w", err)
	}

	peerID, err := peer.Decode(request.PeerID)
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
				PeerID: account.PeerID.String(),
				Name:   account.Name,
				Avatar: account.Avatar,
			}
			receiver = proto.Contact{
				PeerID: contact.PeerID.String(),
				Name:   contact.Name,
				Avatar: contact.Avatar,
			}
		} else if msg.FromPeerID == contact.PeerID {
			sender = proto.Contact{
				PeerID: contact.PeerID.String(),
				Name:   contact.Name,
				Avatar: contact.Avatar,
			}
			receiver = proto.Contact{
				PeerID: account.PeerID.String(),
				Name:   account.Name,
				Avatar: account.Avatar,
			}
		}

		msglist = append(msglist, &proto.ContactMessage{
			ID:         msg.ID,
			Sender:     &sender,
			Receiver:   &receiver,
			MsgType:    proto.MsgType_TEXT_MSG,
			MimeType:   msg.MimeType,
			Data:       msg.Payload,
			CreateTime: msg.Timestamp,
			MsgState:   "",
		})
	}

	reply := &proto.GetContactMessageListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		MessageList: msglist,
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

	peerID, err := peer.Decode(request.PeerID)
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
			ID: "id",
			Sender: &proto.Contact{
				PeerID: account.PeerID.String(),
				Avatar: account.Avatar,
				Name:   account.Name,
				Alias:  "",
			},
			Receiver: &proto.Contact{
				PeerID: contact.PeerID.String(),
				Avatar: contact.Avatar,
				Name:   contact.Name,
				Alias:  "",
			},
			MsgType:    request.MsgType,
			MimeType:   request.MimeType,
			Data:       request.Data,
			MsgState:   "sending",
			CreateTime: time.Now().Unix(),
		},
	}
	return reply, nil
}

func (c *ContactSvc) SetContactAlias(ctx context.Context, request *proto.SetContactAliasRequest) (*proto.SetContactAliasReply, error) {

	if len(contacts) > 0 {
		for i, contact := range contacts {
			if contact.PeerID == request.PeerID {
				contacts[i].Alias = request.Alias
			}
		}
	}

	reply := &proto.SetContactAliasReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Alias: request.Alias,
	}
	return reply, nil
}
