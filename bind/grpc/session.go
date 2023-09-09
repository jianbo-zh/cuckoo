package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/service/contactsvc"
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

	contactSvc, err := s.getContactSvc()
	if err != nil {
		return nil, fmt.Errorf("s.getContactSvc error: %w", err)
	}

	contacts, err := contactSvc.GetContacts(ctx)
	if err != nil {
		return nil, fmt.Errorf("contactSvc.GetContact error: %w", err)
	}
	fmt.Printf("get contacts size: %d\n", len(contacts))
	for i, contact := range contacts {
		if request.Keywords != "" && !strings.Contains(contact.Name, request.Keywords) {
			continue
		}

		sessions = append(sessions, &proto.Session{
			SessionType:       proto.SessionType_CONTACT_SESSION,
			SessionID:         contact.PeerID.String(),
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

	reply := &proto.GetSessionListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		SessionList: sessionList,
	}
	return reply, nil
}
