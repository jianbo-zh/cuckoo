package service

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/systemsvc"
	"github.com/libp2p/go-libp2p/core/peer"
)

var _ proto.SystemSvcServer = (*SystemSvc)(nil)

type SystemSvc struct {
	getter cuckoo.CuckooGetter
	proto.UnimplementedSystemSvcServer
}

func NewSystemSvc(getter cuckoo.CuckooGetter) *SystemSvc {
	return &SystemSvc{
		getter: getter,
	}
}

func (c *SystemSvc) getSystemSvc() (systemsvc.SystemServiceIface, error) {
	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	systemSvc, err := cuckoo.GetSystemSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return systemSvc, nil
}

func (s *SystemSvc) ClearSystemMessage(ctx context.Context, request *proto.ClearSystemMessageRequest) (reply *proto.ClearSystemMessageReply, err error) {

	log.Infoln("ClearSystemMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("ClearSystemMessage panic: ", e)
		} else if err != nil {
			log.Errorln("ClearSystemMessage error: ", err.Error())
		} else {
			log.Infoln("ClearSystemMessage reply: ", reply.String())
		}
	}()

	reply = &proto.ClearSystemMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (c *SystemSvc) ApplyAddContact(ctx context.Context, request *proto.ApplyAddContactRequest) (reply *proto.ApplyAddContactReply, err error) {

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

	systemSvc, err := c.getSystemSvc()
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

	fmt.Println("systemSvc.ApplyAddContact", request.Name, request.Avatar, request.Content)
	err = systemSvc.ApplyAddContact(ctx, peer0, request.Content)
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

func (c *SystemSvc) AgreeAddContact(ctx context.Context, request *proto.AgreeAddContactRequest) (reply *proto.AgreeAddContactReply, err error) {

	log.Infoln("AgreeAddContact request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("AgreeAddContact panic: ", e)
		} else if err != nil {
			log.Errorln("AgreeAddContact error: ", err.Error())
		} else {
			log.Infoln("AgreeAddContact reply: ", reply.String())
		}
	}()

	systemSvc, err := c.getSystemSvc()
	if err != nil {
		return nil, fmt.Errorf("getContactSvc error: %s", err.Error())
	}

	err = systemSvc.AgreeAddContact(ctx, request.AckMsgId)
	if err != nil {
		return nil, fmt.Errorf("systemSvc.AgreeAddContact error: %w", err)
	}

	reply = &proto.AgreeAddContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		AckMsgId: request.AckMsgId,
	}
	return reply, nil
}

func (c *SystemSvc) RejectAddContact(ctx context.Context, request *proto.RejectAddContactRequest) (reply *proto.RejectAddContactReply, err error) {

	log.Infoln("RejectAddContact request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("RejectAddContact panic: ", e)
		} else if err != nil {
			log.Errorln("RejectAddContact error: ", err.Error())
		} else {
			log.Infoln("RejectAddContact reply: ", reply.String())
		}
	}()

	systemSvc, err := c.getSystemSvc()
	if err != nil {
		return nil, fmt.Errorf("getContactSvc error: %s", err.Error())
	}

	err = systemSvc.RejectAddContact(ctx, request.AckMsgId)
	if err != nil {
		return nil, fmt.Errorf("systemSvc.RejectAddContact error: %w", err)
	}

	reply = &proto.RejectAddContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		AckMsgId: request.AckMsgId,
	}
	return reply, nil
}

func (s *SystemSvc) GetSystemMessages(ctx context.Context, request *proto.GetSystemMessagesRequest) (reply *proto.GetSystemMessagesReply, err error) {

	log.Infoln("GetSystemMessages request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetSystemMessages panic: ", e)
		} else if err != nil {
			log.Errorln("GetSystemMessages error: ", err.Error())
		} else {
			log.Infoln("GetSystemMessages reply: ", reply.String())
		}
	}()

	systemSvc, err := s.getSystemSvc()
	if err != nil {
		return nil, nil
	}

	msgs, err := systemSvc.GetSystemMessageList(ctx, int(request.Offset), int(request.Limit))
	if err != nil {
		return nil, fmt.Errorf("systemSvc.GetSystemMessages error: %w", err)
	}

	var msglist []*proto.SystemMessage
	for _, msg := range msgs {
		var msgType proto.SystemMessage_SystemType
		switch msg.SystemType {
		case types.SystemTypeApplyAddContact:
			msgType = proto.SystemMessage_ApplyAddContact
		default:
			msgType = proto.SystemMessage_InviteJoinGroup
		}

		msglist = append(msglist, &proto.SystemMessage{
			Id:         msg.ID,
			SystemType: msgType,
			GroupId:    msg.GroupID,
			FromPeer: &proto.Peer{
				Id:     msg.Sender.ID.String(),
				Name:   msg.Sender.Name,
				Avatar: msg.Sender.Avatar,
			},
			ToPeerId:    msg.Receiver.ID.String(),
			Content:     msg.Content,
			SystemState: msg.SystemState,
			CreateTime:  msg.CreateTime,
			UpdateTime:  msg.UpdateTime,
		})
	}

	reply = &proto.GetSystemMessagesReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Messages: msglist,
	}
	return reply, nil
}
