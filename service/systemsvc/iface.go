package systemsvc

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type SystemServiceIface interface {
	ApplyAddContact(ctx context.Context, peerID peer.ID, content string) error
	AgreeAddContact(ctx context.Context, peerID peer.ID, ackID string) error

	Close()
}
