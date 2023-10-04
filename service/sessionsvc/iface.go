package sessionsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/mytype"
)

type SessionServiceIface interface {
	SetSessionID(ctx context.Context, sessionID string) error
	GetSessions(ctx context.Context) ([]mytype.Session, error)
	UpdateSessionTime(ctx context.Context, sessionID string) error
	SetLastMessage(ctx context.Context, sessionID string, username string, content string) error
	IncrUnreads(ctx context.Context, sessionID string) error
	ResetUnreads(ctx context.Context, sessionID string) error
	Close()
}
