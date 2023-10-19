package sessionsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/protocol/sessionproto"
	"github.com/libp2p/go-libp2p/core/event"
)

var _ SessionServiceIface = (*SessionService)(nil)

type SessionService struct {
	sessionProto *sessionproto.SessionProto
}

func NewSessionService(ctx context.Context, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus) (*SessionService, error) {

	sessionProto, err := sessionproto.NewSessionProto(ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("new download client proto error: %w", err)
	}

	session := SessionService{
		sessionProto: sessionProto,
	}

	return &session, nil
}

func (s *SessionService) SetSessionID(ctx context.Context, sessionID string) error {
	return s.sessionProto.SetSessionID(ctx, sessionID)
}

func (s *SessionService) GetSessions(ctx context.Context) ([]mytype.Session, error) {
	return s.sessionProto.GetSessions(ctx)
}

func (s *SessionService) UpdateSessionTime(ctx context.Context, sessionID string) error {
	return s.sessionProto.UpdateSessionTime(ctx, sessionID)
}

func (s *SessionService) SetLastMessage(ctx context.Context, sessionID string, username string, content string) error {
	return s.sessionProto.SetLastMessage(ctx, sessionID, username, content)
}

func (s *SessionService) IncrUnreads(ctx context.Context, sessionID string) error {
	return s.sessionProto.IncrUnreads(ctx, sessionID)
}

func (s *SessionService) ResetUnreads(ctx context.Context, sessionID string) error {
	return s.sessionProto.ResetUnreads(ctx, sessionID)
}

func (s *SessionService) Close() {}
