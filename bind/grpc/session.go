package service

import (
	"context"
	"strings"
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
			if request.Keywords != "" && !strings.Contains(group.Name, request.Keywords) {
				continue
			}

			sessions = append(sessions, &proto.Session{
				SessionType:       proto.SessionType_GROUP_SESSION,
				SessionID:         group.GroupID,
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
			if request.Keywords != "" && !strings.Contains(contact.Name, request.Keywords) {
				continue
			}

			sessions = append(sessions, &proto.Session{
				SessionType:       proto.SessionType_CONTACT_SESSION,
				SessionID:         contact.PeerID,
				Avatar:            contact.Avatar,
				Name:              contact.Name,
				LastMessage:       "lastmessage",
				LastMessageTime:   time.Now().Unix(),
				HaveUnreadMessage: i%2 == 1,
			})
		}
	}

	var sessionList []*proto.Session
	if int(request.Offset) < len(sessions) {

		endOffset := request.Offset + request.Limit
		if endOffset > int32(len(sessions)) {
			endOffset = int32(len(sessions))
		}

		sessionList = sessions[request.Offset:endOffset]

	} else {
		sessionList = make([]*proto.Session, 0)
	}

	reply := &proto.GetSessionListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		SessionList: sessionList,
	}
	return reply, nil
}
