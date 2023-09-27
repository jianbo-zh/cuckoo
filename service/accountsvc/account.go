package accountsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto/pb"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

type AccountSvc struct {
	accountProto *accountproto.AccountProto
}

func NewAccountService(ctx context.Context, conf config.AccountServiceConfig, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus,
	rdiscvry *drouting.RoutingDiscovery) (*AccountSvc, error) {

	var err error

	accountsvc := &AccountSvc{}

	accountsvc.accountProto, err = accountproto.NewAccountProto(lhost, ids, ebus, conf.AvatarDir)
	if err != nil {
		return nil, fmt.Errorf("peerpeer.NewAccountSvc error: %s", err.Error())
	}

	// 确保账号ID存在
	if err = accountsvc.accountProto.InitAccount(ctx); err != nil {
		return nil, fmt.Errorf("proto init account error: %w", err)
	}

	return accountsvc, nil
}

func (a *AccountSvc) CreateAccount(ctx context.Context, account types.Account) (*types.Account, error) {

	pbAccount, err := a.accountProto.CreateAccount(ctx, &pb.Account{
		Name:                 account.Name,
		Avatar:               account.Avatar,
		AutoAddContact:       account.AutoAddContact,
		AutoJoinGroup:        account.AutoJoinGroup,
		AutoDepositMessage:   account.AutoDepositMessage,
		DepositAddress:       []byte(account.DepositAddress),
		EnableDepositService: account.EnableDepositService,
	})
	if err != nil {
		return nil, fmt.Errorf("proto create account error: %w", err)
	}

	return &types.Account{
		ID:                   peer.ID(pbAccount.Id),
		Name:                 pbAccount.Name,
		Avatar:               pbAccount.Avatar,
		AutoAddContact:       pbAccount.AutoAddContact,
		AutoJoinGroup:        pbAccount.AutoJoinGroup,
		AutoDepositMessage:   pbAccount.AutoDepositMessage,
		DepositAddress:       peer.ID(pbAccount.DepositAddress),
		EnableDepositService: pbAccount.EnableDepositService,
	}, nil

}

func (a *AccountSvc) GetAccount(ctx context.Context) (*types.Account, error) {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("proto get account error: %w", err)
	}

	return &types.Account{
		ID:                   peer.ID(pbAccount.Id),
		Name:                 pbAccount.Name,
		Avatar:               pbAccount.Avatar,
		AutoAddContact:       pbAccount.AutoAddContact,
		AutoJoinGroup:        pbAccount.AutoJoinGroup,
		AutoDepositMessage:   pbAccount.AutoDepositMessage,
		DepositAddress:       peer.ID(pbAccount.DepositAddress),
		EnableDepositService: pbAccount.EnableDepositService,
	}, nil
}

func (a *AccountSvc) SetAccountName(ctx context.Context, name string) error {
	return a.accountProto.UpdateAccountName(ctx, name)
}

func (a *AccountSvc) SetAccountAvatar(ctx context.Context, avatar string) error {
	return a.accountProto.UpdateAccountAvatar(ctx, avatar)
}

func (a *AccountSvc) SetAccountAutoAddContact(ctx context.Context, autoAddContact bool) error {
	return a.accountProto.UpdateAccountAutoAddContact(ctx, autoAddContact)
}

func (a *AccountSvc) SetAccountAutoJoinGroup(ctx context.Context, autoJoinGroup bool) error {
	return a.accountProto.UpdateAccountAutoJoinGroup(ctx, autoJoinGroup)
}

func (a *AccountSvc) SetAutoDepositMessage(ctx context.Context, autoDeposit bool) error {
	return a.accountProto.UpdateAccountAutoDepositMessage(ctx, autoDeposit)
}

func (a *AccountSvc) SetAccountDepositAddress(ctx context.Context, depositPeerID peer.ID) error {
	return a.accountProto.UpdateAccountDepositAddress(ctx, depositPeerID)
}

func (a *AccountSvc) SetAccountEnableDepositService(ctx context.Context, enableDepositService bool) error {
	return a.accountProto.UpdateAccountEnableDepositService(ctx, enableDepositService)
}

func (a *AccountSvc) GetPeer(ctx context.Context, peerID peer.ID) (*types.Peer, error) {
	pbPeer, err := a.accountProto.GetPeer(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("proto get peer error: %w", err)
	}

	return &types.Peer{
		ID:     peer.ID(pbPeer.Id),
		Name:   pbPeer.Name,
		Avatar: pbPeer.Avatar,
	}, nil
}

func (a *AccountSvc) DownloadPeerAvatar(ctx context.Context, peerID peer.ID, avatar string) error {
	return a.accountProto.DownloadPeerAvatar(ctx, peerID, avatar)
}

func (a *AccountSvc) GetOnlineState(ctx context.Context, peerIDs []peer.ID) (res map[peer.ID]bool, err error) {
	return a.accountProto.GetOnlineState(ctx, peerIDs)
}

func (a *AccountSvc) Close() {}
