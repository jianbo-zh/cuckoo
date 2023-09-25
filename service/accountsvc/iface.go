package accountsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/libp2p/go-libp2p/core/peer"
)

type AccountServiceIface interface {
	GetAccount(ctx context.Context) (*types.Account, error)
	CreateAccount(ctx context.Context, account types.Account) (*types.Account, error)
	SetAccountName(ctx context.Context, name string) error
	SetAccountAvatar(ctx context.Context, avatar string) error
	SetAccountAutoAddContact(ctx context.Context, autoAddContact bool) error
	SetAccountAutoJoinGroup(ctx context.Context, autoJoinGroup bool) error

	SetAccountAutoSendDeposit(ctx context.Context, autoSendDeposit bool) error
	SetAccountDepositAddress(ctx context.Context, depositPeerID peer.ID) error
	SetAccountEnableDepositService(ctx context.Context, enableDepositService bool) error

	// 获取在线状态
	GetOnlineState(ctx context.Context, peerIDs []peer.ID) (map[peer.ID]bool, error)

	// peer is stranger no contact
	GetPeer(ctx context.Context, peerID peer.ID) (*types.Peer, error)
	DownloadPeerAvatar(ctx context.Context, peerID peer.ID, avatar string) error

	Close()
}
