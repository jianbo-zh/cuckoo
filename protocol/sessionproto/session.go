package sessionproto

import (
	"context"
	"errors"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	ds "github.com/jianbo-zh/dchat/datastore/ds/sessionds"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/sessionpb"
	"github.com/libp2p/go-libp2p/core/event"
)

type SessionProto struct {
	data ds.SessionIface
}

func NewSessionProto(ids ipfsds.Batching, ebus event.Bus) (*SessionProto, error) {
	sessionProto := SessionProto{
		data: ds.SessionWrap(ids),
	}
	// 订阅器
	sub, err := ebus.Subscribe([]any{new(myevent.EvtClearSession), new(myevent.EvtDeleteSession)})
	if err != nil {
		return nil, fmt.Errorf("ebus subscribe error: %w", err)
	}

	go sessionProto.subscribeHandler(context.Background(), sub)

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
			}

		case <-ctx.Done():
			return
		}
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
