package accountsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

type AccountSvc struct {
	accountProto *accountproto.AccountProto
}

func NewAccountService(conf config.AccountServiceConfig, lhost host.Host, ids ipfsds.Batching, ebus event.Bus,
	rdiscvry *drouting.RoutingDiscovery) (*AccountSvc, error) {

	var err error

	accountsvc := &AccountSvc{}

	accountsvc.accountProto, err = accountproto.NewAccountProto(lhost, ids, ebus, conf.AvatarDir)
	if err != nil {
		return nil, fmt.Errorf("peerpeer.NewAccountSvc error: %s", err.Error())
	}

	return accountsvc, nil
}

func (a *AccountSvc) CreateAccount(ctx context.Context, account Account) (*Account, error) {

	pAccount, err := a.accountProto.CreateAccount(ctx, accountproto.Account{
		PeerID:         account.PeerID,
		Name:           account.Name,
		Avatar:         account.Avatar,
		AutoAddContact: account.AutoAddContact,
		AutoJoinGroup:  account.AutoJoinGroup,
	})
	if err != nil {
		return nil, fmt.Errorf("accountProto.CreateAccount error: %w", err)
	}

	return &Account{
		PeerID:         pAccount.PeerID,
		Name:           pAccount.Name,
		Avatar:         pAccount.Avatar,
		AutoAddContact: pAccount.AutoAddContact,
		AutoJoinGroup:  pAccount.AutoJoinGroup,
	}, nil

}

func (a *AccountSvc) GetAccount(ctx context.Context) (*Account, error) {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("a.accountProto.GetAccount error: %w", err)
	}
	fmt.Println("account name: ", pbAccount.Name)

	return &Account{
		PeerID:         pbAccount.PeerID,
		Name:           pbAccount.Name,
		Avatar:         pbAccount.Avatar,
		AutoAddContact: pbAccount.AutoAddContact,
		AutoJoinGroup:  pbAccount.AutoJoinGroup,
	}, nil
}

func (a *AccountSvc) SetAccountName(ctx context.Context, name string) error {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("get account error: %s", err.Error())
	}
	pbAccount.Name = name
	return a.accountProto.UpdateAccount(ctx, *pbAccount)
}

func (a *AccountSvc) SetAccountAvatar(ctx context.Context, avatar string) error {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("get account error: %s", err.Error())
	}
	pbAccount.Avatar = avatar
	return a.accountProto.UpdateAccount(ctx, *pbAccount)
}

func (a *AccountSvc) SetAccountAutoAddContact(ctx context.Context, autoAddContact bool) error {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("get account error: %s", err.Error())
	}
	pbAccount.AutoAddContact = autoAddContact
	return a.accountProto.UpdateAccount(ctx, *pbAccount)
}

func (a *AccountSvc) SetAccountAutoJoinGroup(ctx context.Context, autoJoinGroup bool) error {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("get account error: %s", err.Error())
	}
	pbAccount.AutoJoinGroup = autoJoinGroup
	return a.accountProto.UpdateAccount(ctx, *pbAccount)
}

func (a *AccountSvc) GetPeer(ctx context.Context, peerID peer.ID) (*Peer, error) {

	pbPeer, err := a.accountProto.GetPeer(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("a.accountProto.GetAccount error: %w", err)
	}

	return &Peer{
		PeerID: pbPeer.PeerID,
		Name:   pbPeer.Name,
		Avatar: pbPeer.Avatar,
	}, nil
}

func (a *AccountSvc) DownloadPeerAvatar(ctx context.Context, peerID peer.ID, avatar string) error {
	return a.accountProto.DownloadPeerAvatar(ctx, peerID, avatar)
}

func (a *AccountSvc) Close() {}
