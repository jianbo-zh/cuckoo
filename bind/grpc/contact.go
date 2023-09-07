package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/service/peersvc"
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

func (c *ContactSvc) getPeerSvc() (peersvc.PeerServiceIface, error) {
	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	peerSvc, err := cuckoo.GetPeerSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return peerSvc, nil
}

func (c *ContactSvc) AddContact(ctx context.Context, request *proto.AddContactRequest) (*proto.AddContactReply, error) {

	peerSvc, err := c.getPeerSvc()
	if err != nil {
		return nil, fmt.Errorf("getPeerSvc error: %s", err.Error())
	}

	peerID, err := peer.Decode(request.PeerID)
	if err != nil {
		return nil, fmt.Errorf("peer.Decode error: %s", err.Error())
	}

	err = peerSvc.ApplyAddContact(ctx, peerID, request.Content)
	if err != nil {
		return nil, fmt.Errorf("peerSvc.AddPeer error: %w", err)
	}

	contacts = append(contacts, &proto.Contact{
		PeerID:      request.GetPeerID(),
		Avatar:      "md5_f4b3ae325c43e3fb08c0c7fbbc57ea63.jpg",
		Name:        request.GetContent(),
		Alias:       "alias1",
		LastMessage: "last message",
		UpdateTime:  time.Now().Unix(),
	})

	reply := &proto.AddContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		PeerID: request.PeerID,
	}
	return reply, nil
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

	if len(peerIDs) > 0 {
		peerSvc, err := cuckoo.GetPeerSvc()
		if err != nil {
			return fmt.Errorf("cuckoo.GetAccountSvc error: %w", err)
		}

		ctx := context.Background()
		for _, peerID := range peerIDs {
			accountBase, err := peerSvc.GetPeer(ctx, peerID)
			if err != nil {
				fmt.Printf("accountSvc.GetPeerAccount error: %s", err.Error())
				continue
			}

			err = peerSvc.DownloadPeerAvatar(ctx, peerID, accountBase.Avatar)
			if err != nil {
				fmt.Printf("accountSvc.DownloadPeerAvatar error: %s", err.Error())
				continue
			}

			server.Send(&proto.GetNearbyContactListStreamReply{
				Result: &proto.Result{
					Code:    0,
					Message: "ok",
				},
				Contact: &proto.Contact{
					PeerID:      accountBase.PeerID.String(),
					Name:        accountBase.Name,
					Avatar:      accountBase.Avatar,
					Alias:       "",
					LastMessage: "",
					UpdateTime:  time.Now().Unix(),
				},
			})
		}
	}

	return nil
}

func (c *ContactSvc) GetContactMessageList(ctx context.Context, request *proto.GetContactMessageListRequest) (*proto.GetContactMessageListReply, error) {
	reply := &proto.GetContactMessageListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		MessageList: contactMessages,
	}
	return reply, nil
}

func (c *ContactSvc) SendContactMessage(ctx context.Context, request *proto.SendContactMessageRequest) (*proto.SendContactMessageReply, error) {

	sendMsg := proto.ContactMessage{
		ID: "id",
		Sender: &proto.Contact{
			PeerID: account.PeerID,
			Avatar: account.Avatar,
			Name:   account.Name,
			Alias:  "",
		},
		Receiver: &proto.Contact{
			PeerID: request.PeerID,
			Avatar: "md5_f4b3ae325c43e3fb08c0c7fbbc57ea63.jpg",
			Name:   "name-8081",
			Alias:  "",
		},
		MsgType:    request.MsgType,
		MimeType:   request.MimeType,
		Data:       request.Data,
		MsgState:   "sending",
		CreateTime: time.Now().Unix(),
	}

	contactMessages = append(contactMessages, &sendMsg)

	reply := &proto.SendContactMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Message: &sendMsg,
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
