package sessionds

import (
	"context"

	pb "github.com/jianbo-zh/dchat/protobuf/pb/sessionpb"
)

type SessionIface interface {
	SetSessionID(ctx context.Context, sessionID string) error
	GetSessionIDs(ctx context.Context) ([]string, error)
	GetLastMessage(ctx context.Context, sessionID string) (*pb.SessionLastMessage, error)
	GetUnreads(ctx context.Context, sessionID string) (int, error)
	UpdateSessionTime(ctx context.Context, sessionID string) error
	SetLastMessage(ctx context.Context, sessionID string, lastMsg *pb.SessionLastMessage) error
	IncrUnreads(ctx context.Context, sessionID string) error
	ResetUnreads(ctx context.Context, sessionID string) error
}
