package systemsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/mytype"
)

type SystemServiceIface interface {
	GetSystemMessageList(ctx context.Context, offset int, limit int) ([]mytype.SystemMessage, error)
	ApplyAddContact(ctx context.Context, peer0 *mytype.Peer, content string) error
	// ApplyJoinGroup(ctx context.Context, groupID string)
	AgreeAddContact(ctx context.Context, msgID string) error
	RejectAddContact(ctx context.Context, msgID string) error

	Close()
}
