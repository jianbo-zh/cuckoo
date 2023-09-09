package service

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
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

func (s *SystemSvc) ClearSystemMessage(ctx context.Context, request *proto.ClearSystemMessageRequest) (*proto.ClearSystemMessageReply, error) {

	reply := &proto.ClearSystemMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (c *SystemSvc) ApplyAddContact(ctx context.Context, request *proto.ApplyAddContactRequest) (*proto.ApplyAddContactReply, error) {

	systemSvc, err := c.getSystemSvc()
	if err != nil {
		return nil, fmt.Errorf("getContactSvc error: %s", err.Error())
	}

	peerID, err := peer.Decode(request.PeerID)
	if err != nil {
		return nil, fmt.Errorf("peer.Decode error: %s", err.Error())
	}

	fmt.Println("systemSvc.ApplyAddContact", request.Name, request.Avatar, request.Content)
	err = systemSvc.ApplyAddContact(ctx, peerID, request.Name, request.Avatar, request.Content)
	if err != nil {
		return nil, fmt.Errorf("peerSvc.AddPeer error: %w", err)
	}

	reply := &proto.ApplyAddContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (c *SystemSvc) AgreeAddContact(ctx context.Context, request *proto.AgreeAddContactRequest) (*proto.AgreeAddContactReply, error) {

	systemSvc, err := c.getSystemSvc()
	if err != nil {
		return nil, fmt.Errorf("getContactSvc error: %s", err.Error())
	}

	err = systemSvc.AgreeAddContact(ctx, request.AckMsgID)
	if err != nil {
		return nil, fmt.Errorf("systemSvc.AgreeAddContact error: %w", err)
	}

	reply := &proto.AgreeAddContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		AckMsgID: request.AckMsgID,
	}
	return reply, nil
}

func (c *SystemSvc) RejectAddContact(ctx context.Context, request *proto.RejectAddContactRequest) (*proto.RejectAddContactReply, error) {

	systemSvc, err := c.getSystemSvc()
	if err != nil {
		return nil, fmt.Errorf("getContactSvc error: %s", err.Error())
	}

	err = systemSvc.RejectAddContact(ctx, request.AckMsgID)
	if err != nil {
		return nil, fmt.Errorf("systemSvc.RejectAddContact error: %w", err)
	}

	reply := &proto.RejectAddContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		AckMsgID: request.AckMsgID,
	}
	return reply, nil
}

func (s *SystemSvc) GetSystemMessageList(ctx context.Context, request *proto.GetSystemMessageListRequest) (*proto.GetSystemMessageListReply, error) {

	systemSvc, err := s.getSystemSvc()
	if err != nil {
		return nil, nil
	}

	msgs, err := systemSvc.GetSystemMessageList(ctx, int(request.Offset), int(request.Limit))
	if err != nil {
		return nil, fmt.Errorf("systemSvc.GetSystemMessageList error: %w", err)
	}

	var msglist []*proto.SystemMessage
	for _, msg := range msgs {
		var msgType proto.SystemType
		switch msg.Type {
		case systemsvc.TypeContactApply:
			msgType = proto.SystemType_APPLY_ADD_CONTACT
		default:
			msgType = proto.SystemType_APPLY_JOIN_GROUP
		}

		var msgState proto.SystemState
		switch msg.State {
		case systemsvc.StateIsSended:
			msgState = proto.SystemState_IS_SENDED
		case systemsvc.StateIsAgree:
			msgState = proto.SystemState_IS_AGREE
		case systemsvc.StateIsReject:
			msgState = proto.SystemState_IS_REJECT
		default:
			// nothing
		}

		msglist = append(msglist, &proto.SystemMessage{
			ID:      msg.ID,
			Type:    msgType,
			GroupID: msg.GroupID,
			Sender: &proto.Peer{
				PeerID: msg.Sender.PeerID.String(),
				Name:   msg.Sender.Name,
				Avatar: msg.Sender.Avatar,
			},
			Receiver: &proto.Peer{
				PeerID: msg.Receiver.PeerID.String(),
				Name:   msg.Receiver.Name,
				Avatar: msg.Receiver.Avatar,
			},
			Content: msg.Content,
			State:   msgState,
			Ctime:   msg.Ctime,
			Utime:   msg.Utime,
		})
	}

	reply := &proto.GetSystemMessageListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		MessageList: msglist,
	}
	return reply, nil
}
