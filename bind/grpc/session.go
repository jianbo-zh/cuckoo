package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/service/contactsvc"
	"github.com/jianbo-zh/dchat/service/groupsvc"
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

func (c *SessionSvc) getContactSvc() (contactsvc.ContactServiceIface, error) {
	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return contactSvc, nil
}

func (c *SessionSvc) getGroupSvc() (groupsvc.GroupServiceIface, error) {
	cuckoo, err := c.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	groupSvc, err := cuckoo.GetGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return groupSvc, nil
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

	var sessions []*proto.Session
	groupSvc, err := s.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("s.getGroupSvc error: %w", err)
	}

	groups, err := groupSvc.GetGroupSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.ListGroups error: %w", err)
	}

	for _, group := range groups {
		sessions = append(sessions, &proto.Session{
			Type:      proto.Session_GroupSession,
			SessionId: group.ID,
			Name:      group.Name,
			Avatar:    group.Avatar,
		})
	}

	contactSvc, err := s.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("s.getContactSvc error: %w", err)
	}

	contacts, err := contactSvc.GetContactSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("contactSvc.GetContact error: %w", err)
	}

	for i, contact := range contacts {
		if request.Keywords != "" && !strings.Contains(contact.Name, request.Keywords) {
			continue
		}

		sessions = append(sessions, &proto.Session{
			Type:              proto.Session_ContactSession,
			SessionId:         contact.ID.String(),
			Name:              contact.Name,
			Avatar:            contact.Avatar,
			LastMessage:       "",
			LastMessageTime:   time.Now().Unix(),
			HaveUnreadMessage: i%2 == 1,
		})
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

	reply = &proto.GetSessionsReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Sessions: sessionList,
	}
	return reply, nil
}
