package systemsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/types"
)

type SystemServiceIface interface {
	GetSystemMessageList(ctx context.Context, offset int, limit int) ([]types.SystemMessage, error)
	ApplyAddContact(ctx context.Context, peer0 *types.Peer, content string) error
	// ApplyJoinGroup(ctx context.Context, groupID string)
	AgreeAddContact(ctx context.Context, ackMsgID string) error
	RejectAddContact(ctx context.Context, ackMsgID string) error

	Close()
}
