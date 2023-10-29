package sessionproto

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/mytype"
	ds "github.com/jianbo-zh/dchat/service/common/datastore/ds/sessionds"
	pb "github.com/jianbo-zh/dchat/service/common/protobuf/pb/sessionpb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
)

var log = logging.Logger("cuckoo/sessionproto")

type SessionNotify struct {
	Updates map[string]int64
	Mutex   sync.Mutex
	Signal  chan bool
}

type SessionProto struct {
	data ds.SessionIface

	sessionNotify SessionNotify

	emitters struct {
		evtSessionAdded   event.Emitter
		evtSessionUpdated event.Emitter
	}
}

func NewSessionProto(ctx context.Context, ids ipfsds.Batching, ebus event.Bus) (*SessionProto, error) {
	var err error
	sessionProto := SessionProto{
		data: ds.SessionWrap(ids),
		sessionNotify: SessionNotify{
			Updates: make(map[string]int64),
			Signal:  make(chan bool, 1),
		},
	}

	if sessionProto.emitters.evtSessionAdded, err = ebus.Emitter(&myevent.EvtSessionAdded{}); err != nil {
		return nil, fmt.Errorf("ebus.Emitter: %s", err.Error())
	}

	if sessionProto.emitters.evtSessionUpdated, err = ebus.Emitter(&myevent.EvtSessionUpdated{}); err != nil {
		return nil, fmt.Errorf("ebus.Emitter: %s", err.Error())
	}

	// 订阅器
	sub, err := ebus.Subscribe([]any{
		new(myevent.EvtClearSession), new(myevent.EvtDeleteSession),
		new(myevent.EvtContactAdded), new(myevent.EvtGroupAdded),
		new(myevent.EvtReceiveContactMessage), new(myevent.EvtReceiveGroupMessage),
	})
	if err != nil {
		return nil, fmt.Errorf("ebus subscribe error: %w", err)
	}

	go sessionProto.subscribeHandler(ctx, sub)

	go sessionProto.loopSessionNotify(ctx)

	return &sessionProto, nil
}

func (s *SessionProto) subscribeHandler(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}

			switch evt := e.(type) {
			case myevent.EvtClearSession:
				go s.handleClearSessionEvent(ctx, evt)

			case myevent.EvtDeleteSession:
				go s.handleDeleteSessionEvent(ctx, evt)

			case myevent.EvtContactAdded:
				go s.handleContactAddedEvent(ctx, evt)

			case myevent.EvtGroupAdded:
				go s.handleGroupAddedEvent(ctx, evt)

			case myevent.EvtReceiveContactMessage:
				go s.handleReceiveContactMessageEvent(ctx, evt)

			case myevent.EvtReceiveGroupMessage:
				go s.handleReceiveGroupMessageEvent(ctx, evt)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (s *SessionProto) loopSessionNotify(ctx context.Context) {
	timeout := time.NewTimer(time.Minute)
	defer func() {
		if !timeout.Stop() {
			<-timeout.C
		}
	}()

	for {
		select {
		case <-s.sessionNotify.Signal:
			if !timeout.Stop() {
				<-timeout.C
			}
			s.checkSessionNotify(false)

		case <-timeout.C:
			s.checkSessionNotify(true)

		case <-ctx.Done():
			return
		}

		timeout.Reset(time.Minute)
	}
}

func (s *SessionProto) checkSessionNotify(timeout bool) {

	s.sessionNotify.Mutex.Lock()
	defer s.sessionNotify.Mutex.Unlock()

	updatedTimes := int64(0)
	for _, num := range s.sessionNotify.Updates {
		updatedTimes += num
	}

	ctx := context.Background()
	if updatedTimes >= 2 || (timeout && updatedTimes > 0) {
		var updateSessions []myevent.UpdateSession
		for sessionID := range s.sessionNotify.Updates {
			lastMsg, err := s.data.GetLastMessage(ctx, sessionID)
			if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
				log.Errorf("data.GetLastMessage error: %w", err)
				return
			}
			unreads, err := s.data.GetUnreads(ctx, sessionID)
			if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
				log.Errorf("data.GetUnreads error: %w", err)
				return
			}

			lastMsgStr := ""
			if lastMsg.Username != "" {
				lastMsgStr += lastMsg.Username + ": "
			}
			if lastMsg.Content != "" {
				lastMsgStr += lastMsg.Content
			}

			updateSessions = append(updateSessions, myevent.UpdateSession{
				ID:      sessionID,
				LastMsg: lastMsgStr,
				Unreads: int64(unreads),
			})

			// 删除原来的
			delete(s.sessionNotify.Updates, sessionID)
		}

		s.emitters.evtSessionUpdated.Emit(myevent.EvtSessionUpdated{
			Sessions: updateSessions,
		})
	}
}

func (s *SessionProto) handleReceiveContactMessageEvent(ctx context.Context, evt myevent.EvtReceiveContactMessage) {
	s.sessionNotify.Mutex.Lock()
	sessionID := mytype.ContactSessionID(evt.FromPeerID)

	if _, exists := s.sessionNotify.Updates[sessionID.String()]; !exists {
		s.sessionNotify.Updates[sessionID.String()] = 1
	} else {
		s.sessionNotify.Updates[sessionID.String()] += 1
	}

	s.sessionNotify.Mutex.Unlock()

	select {
	case s.sessionNotify.Signal <- true:
	default:
	}
}

func (s *SessionProto) handleReceiveGroupMessageEvent(ctx context.Context, evt myevent.EvtReceiveGroupMessage) {
	s.sessionNotify.Mutex.Lock()
	sessionID := mytype.GroupSessionID(evt.GroupID)

	if _, exists := s.sessionNotify.Updates[sessionID.String()]; !exists {
		s.sessionNotify.Updates[sessionID.String()] = 1
	} else {
		s.sessionNotify.Updates[sessionID.String()] += 1
	}

	s.sessionNotify.Mutex.Unlock()
	select {
	case s.sessionNotify.Signal <- true:
	default:
	}
}

func (s *SessionProto) handleContactAddedEvent(ctx context.Context, evt myevent.EvtContactAdded) {
	sessionID := mytype.ContactSessionID(evt.ID)
	if err := s.emitters.evtSessionAdded.Emit(myevent.EvtSessionAdded{
		ID:     sessionID.String(),
		Name:   evt.Name,
		Avatar: evt.Avatar,
		Type:   mytype.ContactSession,
		RelID:  evt.ID.String(),
	}); err != nil {
		log.Errorf("emit new session discover event error: %w", err)
		return
	}
}

func (s *SessionProto) handleGroupAddedEvent(ctx context.Context, evt myevent.EvtGroupAdded) {
	sessionID := mytype.GroupSessionID(evt.ID)
	if err := s.emitters.evtSessionAdded.Emit(myevent.EvtSessionAdded{
		ID:     sessionID.String(),
		Name:   evt.Name,
		Avatar: evt.Avatar,
		Type:   mytype.GroupSession,
		RelID:  evt.ID,
	}); err != nil {
		log.Errorf("emit new session discover event error: %w", err)
		return
	}
}

func (s *SessionProto) handleClearSessionEvent(ctx context.Context, evt myevent.EvtClearSession) {
	var resultErr error
	defer func() {
		evt.Result <- resultErr
		close(evt.Result)
	}()

	if err := s.data.ClearSession(ctx, evt.SessionID); err != nil {
		resultErr = fmt.Errorf("data.ClearSession error: %w", err)
		return
	}
}

func (s *SessionProto) handleDeleteSessionEvent(ctx context.Context, evt myevent.EvtDeleteSession) {
	var resultErr error
	defer func() {
		evt.Result <- resultErr
		close(evt.Result)
	}()

	if err := s.data.DeleteSession(ctx, evt.SessionID); err != nil {
		resultErr = fmt.Errorf("data.DeleteSession error: %w", err)
		return
	}
}

func (s *SessionProto) SetSessionID(ctx context.Context, sessionID string) error {
	return s.data.SetSessionID(ctx, sessionID)
}

func (s *SessionProto) GetSessions(ctx context.Context) ([]mytype.Session, error) {
	var sessions []mytype.Session

	sessionIDs, err := s.data.GetSessionIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("data.GetSessionIDs error: %w", err)
	}

	for _, sessionID := range sessionIDs {
		sid, err := mytype.DecodeSessionID(sessionID)
		if err != nil {
			return nil, fmt.Errorf("decode session id error: %w", err)
		}

		lastmsg, err := s.data.GetLastMessage(ctx, sessionID)
		if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
			return nil, fmt.Errorf("data.GetLastMessage error: %w", err)
		}

		var username, content string
		if lastmsg != nil {
			username = lastmsg.Username
			content = lastmsg.Content
		}

		unreads, err := s.data.GetUnreads(ctx, sessionID)
		if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
			return nil, fmt.Errorf("data.GetUnreads error: %w", err)
		}

		sessions = append(sessions, mytype.Session{
			ID:       *sid,
			Username: username,
			Content:  content,
			Unreads:  unreads,
		})
	}

	return sessions, nil
}

func (s *SessionProto) UpdateSessionTime(ctx context.Context, sessionID string) error {
	return s.data.UpdateSessionTime(ctx, sessionID)
}

func (s *SessionProto) SetLastMessage(ctx context.Context, sessionID string, username string, content string) error {
	return s.data.SetLastMessage(ctx, sessionID, &pb.SessionLastMessage{
		Username: username,
		Content:  content,
	})
}

func (s *SessionProto) IncrUnreads(ctx context.Context, sessionID string) error {
	return s.data.IncrUnreads(ctx, sessionID)
}

func (s *SessionProto) ResetUnreads(ctx context.Context, sessionID string) error {
	return s.data.ResetUnreads(ctx, sessionID)
}
