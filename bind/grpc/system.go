package service

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/service/systemsvc"
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

func (s *SystemSvc) DeleteSystemMessage(ctx context.Context, request *proto.DeleteSystemMessageRequest) (reply *proto.DeleteSystemMessageReply, err error) {

	log.Infoln("DeleteSystemMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("DeleteSystemMessage panic: ", e)
		} else if err != nil {
			log.Errorln("DeleteSystemMessage error: ", err.Error())
		} else {
			log.Infoln("DeleteSystemMessage reply: ", reply.String())
		}
	}()

	systemSvc, err := s.getSystemSvc()
	if err != nil {
		return nil, fmt.Errorf("get system svc error: %w", err)
	}

	if len(request.MessageIds) > 0 {
		if err = systemSvc.DeleteSystemMessage(ctx, request.MessageIds); err != nil {
			return nil, fmt.Errorf("svc delete system message error: %w", err)
		}
	}

	reply = &proto.DeleteSystemMessageReply{
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

func (c *SystemSvc) AgreeJoinGroup(ctx context.Context, request *proto.AgreeJoinGroupRequest) (reply *proto.AgreeJoinGroupReply, err error) {

	log.Infoln("AgreeJoinGroup request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("AgreeJoinGroup panic: ", e)
		} else if err != nil {
			log.Errorln("AgreeJoinGroup error: ", err.Error())
		} else {
			log.Infoln("AgreeJoinGroup reply: ", reply.String())
		}
	}()

	systemSvc, err := c.getSystemSvc()
	if err != nil {
		return nil, fmt.Errorf("getContactSvc error: %s", err.Error())
	}

	err = systemSvc.AgreeJoinGroup(ctx, request.AckMsgId)
	if err != nil {
		return nil, fmt.Errorf("systemSvc.AgreeJoinGroup error: %w", err)
	}

	reply = &proto.AgreeJoinGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		AckMsgId: request.AckMsgId,
	}
	return reply, nil
}

func (c *SystemSvc) RejectJoinGroup(ctx context.Context, request *proto.RejectJoinGroupRequest) (reply *proto.RejectJoinGroupReply, err error) {

	log.Infoln("RejectJoinGroup request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("RejectJoinGroup panic: ", e)
		} else if err != nil {
			log.Errorln("RejectJoinGroup error: ", err.Error())
		} else {
			log.Infoln("RejectJoinGroup reply: ", reply.String())
		}
	}()

	systemSvc, err := c.getSystemSvc()
	if err != nil {
		return nil, fmt.Errorf("getContactSvc error: %s", err.Error())
	}

	err = systemSvc.RejectJoinGroup(ctx, request.AckMsgId)
	if err != nil {
		return nil, fmt.Errorf("systemSvc.RejectJoinGroup error: %w", err)
	}

	reply = &proto.RejectJoinGroupReply{
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
		case mytype.SystemTypeApplyAddContact:
			msgType = proto.SystemMessage_ApplyAddContact
		default:
			msgType = proto.SystemMessage_InviteJoinGroup
		}

		msglist = append(msglist, &proto.SystemMessage{
			Id:         msg.ID,
			SystemType: msgType,
			GroupId:    msg.GroupID,
			FromPeer: &proto.Peer{
				Id:     msg.FromPeer.ID.String(),
				Name:   msg.FromPeer.Name,
				Avatar: msg.FromPeer.Avatar,
			},
			ToPeerId:    msg.ToPeerID.String(),
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
