package service

import (
	"context"
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
)

var _ proto.ContactSvcServer = (*ContactSvc)(nil)

type ContactSvc struct {
	proto.UnimplementedContactSvcServer
}

func (c *ContactSvc) AddContact(ctx context.Context, request *proto.AddContactRequest) (*proto.AddContactReply, error) {

	contacts = append(contacts, &proto.Contact{
		PeerID: request.GetPeerID(),
		Avatar: "avatar1",
		Name:   "name1",
		Alias:  "alias1",
	})

	reply := &proto.AddContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
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
	reply := &proto.GetContactListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		ContactList: contacts,
	}
	return reply, nil
}

func (c *ContactSvc) GetNearbyContactList(context.Context, *proto.GetNearbyContactListRequest) (*proto.GetNearbyContactListReply, error) {
	reply := &proto.GetNearbyContactListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		ContactList: contacts,
	}
	return reply, nil
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
	contactMessages = append(contactMessages, &proto.ContactMessage{
		ID: "id",
		Sender: &proto.Contact{
			PeerID: "peerID",
			Avatar: "avatar",
			Name:   "name",
			Alias:  "alias",
		},
		Receiver: &proto.Contact{
			PeerID: "peerID1",
			Avatar: "avatar1",
			Name:   "name1",
			Alias:  "alias1",
		},
		MsgType:    request.MsgType,
		MimeType:   request.MimeType,
		Data:       request.Data,
		CreateTime: time.Now().Unix(),
	})

	reply := &proto.SendContactMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (c *ContactSvc) SetContactAlias(ctx context.Context, request *proto.SetContactAliasRequest) (*proto.SetContactAliasReply, error) {

	if len(contacts) > 0 {
		contacts[0].Alias = request.Alias
	}

	reply := &proto.SetContactAliasReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}
