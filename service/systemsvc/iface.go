package systemsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/mytype"
)

type SystemServiceIface interface {
	GetSystemMessageList(ctx context.Context, offset int, limit int) ([]mytype.SystemMessage, error)

	AgreeAddContact(ctx context.Context, msgID string) error
	RejectAddContact(ctx context.Context, msgID string) error

	AgreeJoinGroup(ctx context.Context, msgID string) error
	RejectJoinGroup(ctx context.Context, ackID string) error

	DeleteSystemMessage(ctx context.Context, msgIDs []string) error

	Close()
}
