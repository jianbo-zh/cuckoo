package service

import (
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	myevent "github.com/jianbo-zh/dchat/event"
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

	sub, err := ebus.Subscribe([]any{new(myevent.EvtReceivePeerMessage), new(myevent.EvtReceiveGroupMessage)})
	if err != nil {
		return fmt.Errorf("ebus.Subscribe error: %w", err)
	}
	defer sub.Close()

	for e := range sub.Out() {
		switch evt := e.(type) {
		case myevent.EvtReceivePeerMessage:
			payload, err := goproto.Marshal(&proto.CommonEvent_PayloadPeerMessage{
				MsgId:      evt.MsgID,
				FromPeerId: evt.FromPeerID.String(),
			})
			if err != nil {
				return fmt.Errorf("proto.Marshal error: %w", err)
			}

			reply := proto.SubscribeCommonEventReply{
				Event: &proto.CommonEvent{
					Type:    proto.CommonEvent_PeerMessage,
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
					Type:    proto.CommonEvent_GroupMessage,
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
