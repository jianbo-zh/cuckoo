package service

import (
	"context"
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
)

var _ proto.SessionSvcServer = (*SessionSvc)(nil)

type SessionSvc struct {
	proto.UnimplementedSessionSvcServer
}

func (s *SessionSvc) GetSessionList(ctx context.Context, request *proto.GetSessionListRequest) (*proto.GetSessionListReply, error) {

	var sessions []*proto.Session
	if len(groups) > 0 {
		for i, group := range groups {
			sessions = append(sessions, &proto.Session{
				Avatar:            group.Avatar,
				Name:              group.Name,
				LastMessage:       "lastmessage",
				LastMessageTime:   time.Now().Unix(),
				HaveUnreadMessage: i%2 == 0,
			})
		}
	}

	if len(contacts) > 0 {
		for i, contact := range contacts {
			sessions = append(sessions, &proto.Session{
				Avatar:            contact.Avatar,
				Name:              contact.Name,
				LastMessage:       "lastmessage",
				LastMessageTime:   time.Now().Unix(),
				HaveUnreadMessage: i%2 == 1,
			})
		}
	}

	reply := &proto.GetSessionListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		SessionList: sessions,
	}
	return reply, nil
}
