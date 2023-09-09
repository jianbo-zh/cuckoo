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

func (c *SubscribeSvc) SubscribeCommonEvent(request *proto.SubscribeCommonEventRequest, server proto.SubscribeSvc_SubscribeCommonEventServer) error {
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

		case myevent.EvtReceiveGroupMessage:
			// todo:
		default:
			// nothing to do
		}
	}

	return nil
}
