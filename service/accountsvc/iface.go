package accountsvc

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type AccountServiceIface interface {
	GetAccount(ctx context.Context) (*Account, error)
	CreateAccount(ctx context.Context, account Account) (*Account, error)
	SetAccountName(ctx context.Context, name string) error
	SetAccountAvatar(ctx context.Context, avatar string) error
	SetAccountAutoAddContact(ctx context.Context, autoAddContact bool) error
	SetAccountAutoJoinGroup(ctx context.Context, autoJoinGroup bool) error

	// peer is stranger no contact
	GetPeer(ctx context.Context, peerID peer.ID) (*Peer, error)
	DownloadPeerAvatar(ctx context.Context, peerID peer.ID, avatar string) error

	Close()
}
