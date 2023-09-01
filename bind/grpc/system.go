package service

import (
	"context"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
)

var _ proto.SystemSvcServer = (*SystemSvc)(nil)

type SystemSvc struct {
	proto.UnimplementedSystemSvcServer
}

func (s *SystemSvc) ClearSystemMessage(ctx context.Context, request *proto.ClearSystemMessageRequest) (*proto.ClearSystemMessageReply, error) {

	systemMessage = make([]*proto.SystemMessage, 0)

	reply := &proto.ClearSystemMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (s *SystemSvc) GetSystemMessageList(ctx context.Context, request *proto.GetSystemMessageListRequest) (*proto.GetSystemMessageListReply, error) {

	reply := &proto.GetSystemMessageListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		MessageList: systemMessage,
	}
	return reply, nil
}
