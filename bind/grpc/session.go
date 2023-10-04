package service

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

var _ proto.SessionSvcServer = (*SessionSvc)(nil)

type SessionSvc struct {
	getter cuckoo.CuckooGetter
	proto.UnimplementedSessionSvcServer
}

func NewSessionSvc(getter cuckoo.CuckooGetter) *SessionSvc {
	return &SessionSvc{
		getter: getter,
	}
}

func (s *SessionSvc) GetSessions(ctx context.Context, request *proto.GetSessionsRequest) (reply *proto.GetSessionsReply, err error) {

	log.Infoln("GetSessions request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetSessions panic: ", e)
		} else if err != nil {
			log.Errorln("GetSessions error: ", err.Error())
		} else {
			log.Infoln("GetSessions reply: ", reply.String())
		}
	}()

	cuckoo, err := s.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	groupSvc, err := cuckoo.GetGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("s.getGroupSvc error: %w", err)
	}

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return nil, fmt.Errorf("s.getContactSvc error: %w", err)
	}

	accountSvc, err := cuckoo.GetAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("get account svc error: %w", err)
	}

	sessionSvc, err := cuckoo.GetSessionSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetSessionSvc error: %w", err)
	}

	sessions, err := sessionSvc.GetSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get sessions error: %w", err)
	}

	var contactIDs []peer.ID
	for _, session := range sessions {
		switch session.ID.Type {
		case mytype.ContactSession:
			contactIDs = append(contactIDs, peer.ID(session.ID.Value))
		default:
			// nothing
		}
	}

	onlineStateMap := accountSvc.GetOnlineState(contactIDs)

	var sessionList []*proto.Session
	for _, session := range sessions {
		switch session.ID.Type {
		case mytype.ContactSession:
			contactID := peer.ID(session.ID.Value)
			contact, err := contactSvc.GetContact(ctx, contactID)
			if err != nil {
				return nil, fmt.Errorf("svc get contact error: %w", err)
			}
			sessionList = append(sessionList, &proto.Session{
				Type:    proto.Session_ContactSession,
				Id:      contactID.String(), // 这里直接用联系人ID
				Name:    contact.Name,
				Avatar:  contact.Avatar,
				State:   encodeOnlineState(onlineStateMap[contactID]),
				Lastmsg: session.Content,
				Unreads: int64(session.Unreads),
			})

		case mytype.GroupSession:
			groupID := string(session.ID.Value)
			group, err := groupSvc.GetGroup(ctx, groupID)
			if err != nil {
				return nil, fmt.Errorf("svc get group error: %w", err)
			}
			sessionList = append(sessionList, &proto.Session{
				Type:    proto.Session_GroupSession,
				Id:      groupID, // 这里直接用群组ID
				Name:    group.Name,
				Avatar:  group.Avatar,
				State:   proto.ConnState_UnknownState,
				Lastmsg: session.Username + ": " + session.Content,
				Unreads: int64(session.Unreads),
			})
		default:
			// nothing
			return nil, fmt.Errorf("unknown session type")
		}
	}

	reply = &proto.GetSessionsReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Sessions: sessionList,
	}
	return reply, nil
}
