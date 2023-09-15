package accountsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto/pb"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

type AccountSvc struct {
	accountProto *accountproto.AccountProto
}

func NewAccountService(ctx context.Context, conf config.AccountServiceConfig, lhost host.Host, ids ipfsds.Batching, ebus event.Bus,
	rdiscvry *drouting.RoutingDiscovery) (*AccountSvc, error) {

	var err error

	accountsvc := &AccountSvc{}

	accountsvc.accountProto, err = accountproto.NewAccountProto(lhost, ids, ebus, conf.AvatarDir)
	if err != nil {
		return nil, fmt.Errorf("peerpeer.NewAccountSvc error: %s", err.Error())
	}

	return accountsvc, nil
}

func (a *AccountSvc) CreateAccount(ctx context.Context, account types.Account) (*types.Account, error) {

	pbAccount, err := a.accountProto.CreateAccount(ctx, &pb.Account{
		PeerId:         []byte(account.ID),
		Name:           account.Name,
		Avatar:         account.Avatar,
		AutoAddContact: account.AutoAddContact,
		AutoJoinGroup:  account.AutoJoinGroup,
	})
	if err != nil {
		return nil, fmt.Errorf("proto create account error: %w", err)
	}

	return &types.Account{
		ID:             peer.ID(pbAccount.PeerId),
		Name:           pbAccount.Name,
		Avatar:         pbAccount.Avatar,
		AutoAddContact: pbAccount.AutoAddContact,
		AutoJoinGroup:  pbAccount.AutoJoinGroup,
	}, nil

}

func (a *AccountSvc) GetAccount(ctx context.Context) (*types.Account, error) {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("proto get account error: %w", err)
	}

	return &types.Account{
		ID:             peer.ID(pbAccount.PeerId),
		Name:           pbAccount.Name,
		Avatar:         pbAccount.Avatar,
		AutoAddContact: pbAccount.AutoAddContact,
		AutoJoinGroup:  pbAccount.AutoJoinGroup,
	}, nil
}

func (a *AccountSvc) SetAccountName(ctx context.Context, name string) error {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("proto get account error: %w", err)
	}
	pbAccount.Name = name
	return a.accountProto.UpdateAccount(ctx, pbAccount)
}

func (a *AccountSvc) SetAccountAvatar(ctx context.Context, avatar string) error {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("proto get account error: %w", err)
	}
	pbAccount.Avatar = avatar
	return a.accountProto.UpdateAccount(ctx, pbAccount)
}

func (a *AccountSvc) SetAccountAutoAddContact(ctx context.Context, autoAddContact bool) error {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("proto get account error: %w", err)
	}
	pbAccount.AutoAddContact = autoAddContact
	return a.accountProto.UpdateAccount(ctx, pbAccount)
}

func (a *AccountSvc) SetAccountAutoJoinGroup(ctx context.Context, autoJoinGroup bool) error {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("proto get account error: %w", err)
	}
	pbAccount.AutoJoinGroup = autoJoinGroup
	return a.accountProto.UpdateAccount(ctx, pbAccount)
}

func (a *AccountSvc) GetPeer(ctx context.Context, peerID peer.ID) (*types.Peer, error) {
	pbPeer, err := a.accountProto.GetPeer(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("proto get peer error: %w", err)
	}

	return &types.Peer{
		ID:     peer.ID(pbPeer.PeerId),
		Name:   pbPeer.Name,
		Avatar: pbPeer.Avatar,
	}, nil
}

func (a *AccountSvc) DownloadPeerAvatar(ctx context.Context, peerID peer.ID, avatar string) error {
	return a.accountProto.DownloadPeerAvatar(ctx, peerID, avatar)
}

func (a *AccountSvc) Close() {}
