package service

import (
	"context"
	"errors"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/service/accountsvc"
)

var _ proto.AccountSvcServer = (*AccountSvc)(nil)

type AccountSvc struct {
	getter cuckoo.CuckooGetter
	proto.UnimplementedAccountSvcServer
}

func NewAccountSvc(getter cuckoo.CuckooGetter) *AccountSvc {
	return &AccountSvc{
		getter: getter,
	}
}

func (a *AccountSvc) getAccountSvc() (accountsvc.AccountServiceIface, error) {
	cuckoo, err := a.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	accountSvc, err := cuckoo.GetAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetAccountSvc error: %s", err.Error())
	}

	return accountSvc, nil
}

func (a *AccountSvc) CreateAccount(ctx context.Context, request *proto.CreateAccountRequest) (*proto.CreateAccountReply, error) {

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	fullAccount, err := accountSvc.CreateAccount(ctx, accountsvc.Account{
		Name:           request.Name,
		Avatar:         request.Avatar,
		AutoAddContact: true,
		AutoJoinGroup:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("accountSvc.CreateAccount error: %w", err)
	}

	account = proto.Account{
		PeerID:         fullAccount.PeerID.String(),
		Name:           fullAccount.Name,
		Avatar:         fullAccount.Avatar,
		AutoAddContact: fullAccount.AutoAddContact,
		AutoJoinGroup:  fullAccount.AutoJoinGroup,
	}

	reply := &proto.CreateAccountReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Account: &account,
	}
	return reply, nil
}

func (a *AccountSvc) GetAccount(ctx context.Context, request *proto.GetAccountRequest) (*proto.GetAccountReply, error) {

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	var protoAccount proto.Account
	account, err := accountSvc.GetAccount(ctx)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("accountSvc.GetAccount error: %w", err)

	} else if account != nil {
		protoAccount = proto.Account{
			PeerID:         account.PeerID.String(),
			Name:           account.Name,
			Avatar:         account.Avatar,
			AutoAddContact: account.AutoAddContact,
			AutoJoinGroup:  account.AutoJoinGroup,
		}
	}

	reply := &proto.GetAccountReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Account: &protoAccount,
	}

	return reply, nil
}

func (a *AccountSvc) SetAccountAvatar(ctx context.Context, request *proto.SetAccountAvatarRequest) (*proto.SetAccountAvatarReply, error) {

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	err = accountSvc.SetAccountAvatar(ctx, request.GetAvatar())
	if err != nil {
		return nil, fmt.Errorf("accountSvc.SetAvatar error: %w", err)
	}

	reply := &proto.SetAccountAvatarReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Avatar: request.GetAvatar(),
	}
	return reply, nil
}

func (a *AccountSvc) SetAccountName(ctx context.Context, request *proto.SetAccountNameRequest) (*proto.SetAccountNameReply, error) {

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	err = accountSvc.SetAccountName(ctx, request.GetName())
	if err != nil {
		return nil, fmt.Errorf("accountSvc.SetName error: %w", err)
	}

	reply := &proto.SetAccountNameReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Name: request.GetName(),
	}
	return reply, nil
}

func (a *AccountSvc) SetAutoAddContact(ctx context.Context, request *proto.SetAutoAddContactRequest) (*proto.SetAutoAddContactReply, error) {

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	err = accountSvc.SetAccountAutoAddContact(ctx, request.GetIsReview())
	if err != nil {
		return nil, fmt.Errorf("accountSvc.SetAutoAddContact error: %w", err)
	}

	reply := &proto.SetAutoAddContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		IsReview: request.GetIsReview(),
	}
	return reply, nil
}

func (a *AccountSvc) SetAutoJoinGroup(ctx context.Context, request *proto.SetAutoJoinGroupRequest) (*proto.SetAutoJoinGroupReply, error) {

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	err = accountSvc.SetAccountAutoJoinGroup(ctx, request.GetIsReview())
	if err != nil {
		return nil, fmt.Errorf("accountSvc.SetAutoJoinGroup error: %w", err)
	}

	reply := &proto.SetAutoJoinGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		IsReview: request.GetIsReview(),
	}
	return reply, nil
}
