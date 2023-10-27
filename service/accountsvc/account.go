package accountsvc

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/service/accountsvc/protobuf/pb/accountpb"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocols/accountproto"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var log = logging.Logger("accountsvc")

var _ AccountServiceIface = (*AccountSvc)(nil)

type AccountSvc struct {
	accountProto *accountproto.AccountProto
}

func NewAccountService(ctx context.Context, avatarDir string, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus,
	rdiscvry *drouting.RoutingDiscovery) (*AccountSvc, error) {

	var err error

	accountsvc := &AccountSvc{}

	accountsvc.accountProto, err = accountproto.NewAccountProto(lhost, ids, ebus, avatarDir)
	if err != nil {
		return nil, fmt.Errorf("peerpeer.NewAccountSvc error: %s", err.Error())
	}

	// 确保账号ID存在
	if err = accountsvc.accountProto.InitAccount(ctx); err != nil {
		return nil, fmt.Errorf("proto init account error: %w", err)
	}

	return accountsvc, nil
}

func (a *AccountSvc) CreateAccount(ctx context.Context, account mytype.Account) (*mytype.Account, error) {

	pbAccount, err := a.accountProto.CreateAccount(ctx, &pb.Account{
		Name:               account.Name,
		Avatar:             account.Avatar,
		AutoAddContact:     account.AutoAddContact,
		AutoJoinGroup:      account.AutoJoinGroup,
		AutoDepositMessage: account.AutoDepositMessage,
		DepositAddress:     []byte(account.DepositAddress),
	})
	if err != nil {
		return nil, fmt.Errorf("proto create account error: %w", err)
	}

	return &mytype.Account{
		ID:                 peer.ID(pbAccount.Id),
		Name:               pbAccount.Name,
		Avatar:             pbAccount.Avatar,
		AutoAddContact:     pbAccount.AutoAddContact,
		AutoJoinGroup:      pbAccount.AutoJoinGroup,
		AutoDepositMessage: pbAccount.AutoDepositMessage,
		DepositAddress:     peer.ID(pbAccount.DepositAddress),
	}, nil

}

func (a *AccountSvc) GetAccount(ctx context.Context) (*mytype.Account, error) {
	pbAccount, err := a.accountProto.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("proto get account error: %w", err)
	}

	return &mytype.Account{
		ID:                 peer.ID(pbAccount.Id),
		Name:               pbAccount.Name,
		Avatar:             pbAccount.Avatar,
		AutoAddContact:     pbAccount.AutoAddContact,
		AutoJoinGroup:      pbAccount.AutoJoinGroup,
		AutoDepositMessage: pbAccount.AutoDepositMessage,
		DepositAddress:     peer.ID(pbAccount.DepositAddress),
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

func (a *AccountSvc) GetPeer(ctx context.Context, peerID peer.ID) (*mytype.Peer, error) {
	pbPeer, err := a.accountProto.GetPeer(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("proto get peer error: %w", err)
	}

	return &mytype.Peer{
		ID:             peer.ID(pbPeer.Id),
		Name:           pbPeer.Name,
		Avatar:         pbPeer.Avatar,
		DepositAddress: peer.ID(pbPeer.DepositAddress),
	}, nil
}

func (a *AccountSvc) GetOnlineState(peerIDs []peer.ID) map[peer.ID]mytype.OnlineState {
	return a.accountProto.GetOnlineState(peerIDs)
}

// AsyncCheckOnlineState 异步探测Peer是否在线
func (a *AccountSvc) AsyncCheckOnlineState(peerID peer.ID) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := a.accountProto.CheckOnlineState(ctx, peerID); err != nil {
			log.Errorf("proto.CheckOnlineState error: %v", err)
		}

	}()
}

func (a *AccountSvc) Close() {}
