package service

import (
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	myevent "github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/mytype"
	goproto "google.golang.org/protobuf/proto"
)

var _ proto.SubscribeSvcServer = (*SubscribeSvc)(nil)

type SubscribeSvc struct {
	getter cuckoo.CuckooGetter
	proto.UnimplementedSubscribeSvcServer
}

func NewSubscribeSvc(getter cuckoo.CuckooGetter) *SubscribeSvc {
	return &SubscribeSvc{
		getter: getter,
	}
}

func (c *SubscribeSvc) SubscribeCommonEvent(request *proto.SubscribeCommonEventRequest, server proto.SubscribeSvc_SubscribeCommonEventServer) (err error) {

	log.Infoln("SubscribeCommonEvent request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SubscribeCommonEvent panic: ", e)
		} else if err != nil {
			log.Errorln("SubscribeCommonEvent error: ", err.Error())
		}
	}()

	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return fmt.Errorf("c.getter.GetCuckoo error: %w", err)
	}

	ebus, err := cuckoo.GetEbus()
	if err != nil {
		return fmt.Errorf("cuckoo.GetEbus error: %w", err)
	}

	sub, err := ebus.Subscribe([]any{
		new(myevent.EvtReceiveContactMessage),
		new(myevent.EvtReceiveGroupMessage),
		new(myevent.EvtPeerStateChanged),
		new(myevent.EvtSessionAdded),
		new(myevent.EvtSessionUpdated),
	})
	if err != nil {
		return fmt.Errorf("ebus.Subscribe error: %w", err)
	}
	defer sub.Close()

	for e := range sub.Out() {
		switch evt := e.(type) {
		case myevent.EvtReceiveContactMessage:
			payload, err := goproto.Marshal(&proto.CommonEvent_PayloadPeerMessage{
				MsgId:      evt.MsgID,
				FromPeerId: evt.FromPeerID.String(),
			})
			if err != nil {
				return fmt.Errorf("proto.Marshal error: %w", err)
			}

			reply := proto.SubscribeCommonEventReply{
				Event: &proto.CommonEvent{
					Type:    proto.CommonEvent_PeerMessageReceived,
					Payload: payload,
				},
			}

			err = server.Send(&reply)
			if err != nil {
				return fmt.Errorf("server.Send error: %w", err)
			}

			log.Infoln("SubscribeCommonEvent reply: ", reply.String())

		case myevent.EvtReceiveGroupMessage:
			payload, err := goproto.Marshal(&proto.CommonEvent_PayloadGroupMessage{
				MsgId:   evt.MsgID,
				GroupId: evt.GroupID,
			})
			if err != nil {
				return fmt.Errorf("proto.Marshal error: %w", err)
			}

			reply := proto.SubscribeCommonEventReply{
				Event: &proto.CommonEvent{
					Type:    proto.CommonEvent_GroupMessageReceived,
					Payload: payload,
				},
			}

			err = server.Send(&reply)
			if err != nil {
				return fmt.Errorf("server.Send error: %w", err)
			}

			log.Infoln("SubscribeCommonEvent reply: ", reply.String())

		case myevent.EvtPeerStateChanged:
			payload, err := goproto.Marshal(&proto.CommonEvent_PayloadPeerState{
				PeerId: evt.PeerID.String(),
				Online: evt.Online,
			})
			if err != nil {
				return fmt.Errorf("proto.Marshal error: %w", err)
			}

			reply := proto.SubscribeCommonEventReply{
				Event: &proto.CommonEvent{
					Type:    proto.CommonEvent_PeerStateChanged,
					Payload: payload,
				},
			}

			err = server.Send(&reply)
			if err != nil {
				return fmt.Errorf("server.Send error: %w", err)
			}

			log.Infoln("SubscribeCommonEvent reply: ", reply.String())

		case myevent.EvtSessionAdded:

			var sessionType proto.SessionType
			switch evt.Type {
			case mytype.ContactSession:
				sessionType = proto.SessionType_ContactSessionType
			case mytype.GroupSession:
				sessionType = proto.SessionType_GroupSessionType
			default:
				return fmt.Errorf("unknown session type")
			}

			payload, err := goproto.Marshal(&proto.CommonEvent_PayloadSessionAdd{
				Session: &proto.CommonEvent_AddSession{
					SessionType: sessionType,
					Id:          evt.ID,
					Name:        evt.Name,
					Avatar:      evt.Avatar,
					RelId:       evt.RelID,
				},
			})
			if err != nil {
				return fmt.Errorf("proto.Marshal error: %w", err)
			}

			reply := proto.SubscribeCommonEventReply{
				Event: &proto.CommonEvent{
					Type:    proto.CommonEvent_SessionAdded,
					Payload: payload,
				},
			}

			err = server.Send(&reply)
			if err != nil {
				return fmt.Errorf("server.Send error: %w", err)
			}

			log.Infoln("SubscribeCommonEvent reply: ", reply.String())

		case myevent.EvtSessionUpdated:
			var sessionStates []*proto.CommonEvent_SessionState
			for _, session := range evt.Sessions {
				sessionStates = append(sessionStates, &proto.CommonEvent_SessionState{
					Id:      session.ID,
					LastMsg: session.LastMsg,
					Unreads: int32(session.Unreads),
				})
			}
			payload, err := goproto.Marshal(&proto.CommonEvent_PayloadSessionUpdate{
				SessionStates: sessionStates,
			})
			if err != nil {
				return fmt.Errorf("proto.Marshal error: %w", err)
			}

			reply := proto.SubscribeCommonEventReply{
				Event: &proto.CommonEvent{
					Type:    proto.CommonEvent_SessionUpdated,
					Payload: payload,
				},
			}

			err = server.Send(&reply)
			if err != nil {
				return fmt.Errorf("server.Send error: %w", err)
			}

			log.Infoln("SubscribeCommonEvent reply: ", reply.String())
		default:
			// nothing to do
		}
	}

	return nil
}
