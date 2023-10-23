package accountsvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

type AccountServiceIface interface {
	GetAccount(ctx context.Context) (*mytype.Account, error)
	CreateAccount(ctx context.Context, account mytype.Account) (*mytype.Account, error)
	SetAccountName(ctx context.Context, name string) error
	SetAccountAvatar(ctx context.Context, avatar string) error
	SetAccountAutoAddContact(ctx context.Context, autoAddContact bool) error
	SetAccountAutoJoinGroup(ctx context.Context, autoJoinGroup bool) error

	SetAutoDepositMessage(ctx context.Context, autoDepositMessage bool) error
	SetAccountDepositAddress(ctx context.Context, depositPeerID peer.ID) error

	// 获取在线状态
	GetOnlineState(peerIDs []peer.ID) map[peer.ID]mytype.OnlineState
	AsyncCheckOnlineState(peerID peer.ID)

	// peer is stranger no contact
	GetPeer(ctx context.Context, peerID peer.ID) (*mytype.Peer, error)

	Close()
}
