package service

import (
	"context"
	"errors"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/internal/types"
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

func (a *AccountSvc) CreateAccount(ctx context.Context, request *proto.CreateAccountRequest) (reply *proto.CreateAccountReply, err error) {

	log.Infoln("CreateAccount request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("CreateAccount panic: ", e)
		} else if err != nil {
			log.Errorln("CreateAccount error: ", err.Error())
		} else {
			log.Infoln("CreateAccount reply: ", reply.String())
		}
	}()

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	fullAccount, err := accountSvc.CreateAccount(ctx, types.Account{
		Name:           request.Name,
		Avatar:         request.Avatar,
		AutoAddContact: true,
		AutoJoinGroup:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("accountSvc.CreateAccount error: %w", err)
	}

	account := proto.Account{
		PeerId:         fullAccount.ID.String(),
		Name:           fullAccount.Name,
		Avatar:         fullAccount.Avatar,
		AutoAddContact: fullAccount.AutoAddContact,
		AutoJoinGroup:  fullAccount.AutoJoinGroup,
	}

	reply = &proto.CreateAccountReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Account: &account,
	}
	return reply, nil
}

func (a *AccountSvc) GetAccount(ctx context.Context, request *proto.GetAccountRequest) (reply *proto.GetAccountReply, err error) {

	log.Infoln("GetAccount request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetAccount panic: ", e)
		} else if err != nil {
			log.Errorln("GetAccount error: ", err.Error())
		} else {
			log.Infoln("GetAccount reply: ", reply.String())
		}
	}()

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
			PeerId:         account.ID.String(),
			Name:           account.Name,
			Avatar:         account.Avatar,
			AutoAddContact: account.AutoAddContact,
			AutoJoinGroup:  account.AutoJoinGroup,
		}
	}

	reply = &proto.GetAccountReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Account: &protoAccount,
	}

	return reply, nil
}

func (a *AccountSvc) SetAccountAvatar(ctx context.Context, request *proto.SetAccountAvatarRequest) (reply *proto.SetAccountAvatarReply, err error) {

	log.Infoln("SetAccountAvatar request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetAccountAvatar panic: ", e)
		} else if err != nil {
			log.Errorln("SetAccountAvatar error: ", err.Error())
		} else {
			log.Infoln("SetAccountAvatar reply: ", reply.String())
		}
	}()

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	err = accountSvc.SetAccountAvatar(ctx, request.GetAvatar())
	if err != nil {
		return nil, fmt.Errorf("accountSvc.SetAvatar error: %w", err)
	}

	reply = &proto.SetAccountAvatarReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Avatar: request.GetAvatar(),
	}
	return reply, nil
}

func (a *AccountSvc) SetAccountName(ctx context.Context, request *proto.SetAccountNameRequest) (reply *proto.SetAccountNameReply, err error) {

	log.Infoln("SetAccountName request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetAccountName panic: ", e)
		} else if err != nil {
			log.Errorln("SetAccountName error: ", err.Error())
		} else {
			log.Infoln("SetAccountName reply: ", reply.String())
		}
	}()

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	err = accountSvc.SetAccountName(ctx, request.GetName())
	if err != nil {
		return nil, fmt.Errorf("accountSvc.SetName error: %w", err)
	}

	reply = &proto.SetAccountNameReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Name: request.GetName(),
	}
	return reply, nil
}

func (a *AccountSvc) SetAutoAddContact(ctx context.Context, request *proto.SetAutoAddContactRequest) (reply *proto.SetAutoAddContactReply, err error) {

	log.Infoln("SetAutoAddContact request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetAutoAddContact panic: ", e)
		} else if err != nil {
			log.Errorln("SetAutoAddContact error: ", err.Error())
		} else {
			log.Infoln("SetAutoAddContact reply: ", reply.String())
		}
	}()

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	err = accountSvc.SetAccountAutoAddContact(ctx, request.GetIsAuto())
	if err != nil {
		return nil, fmt.Errorf("accountSvc.SetAutoAddContact error: %w", err)
	}

	reply = &proto.SetAutoAddContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		IsAuto: request.GetIsAuto(),
	}
	return reply, nil
}

func (a *AccountSvc) SetAutoJoinGroup(ctx context.Context, request *proto.SetAutoJoinGroupRequest) (reply *proto.SetAutoJoinGroupReply, err error) {

	log.Infoln("SetAutoJoinGroup request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetAutoJoinGroup panic: ", e)
		} else if err != nil {
			log.Errorln("SetAutoJoinGroup error: ", err.Error())
		} else {
			log.Infoln("SetAutoJoinGroup reply: ", reply.String())
		}
	}()

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	err = accountSvc.SetAccountAutoJoinGroup(ctx, request.GetIsAuto())
	if err != nil {
		return nil, fmt.Errorf("accountSvc.SetAutoJoinGroup error: %w", err)
	}

	reply = &proto.SetAutoJoinGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		IsAuto: request.GetIsAuto(),
	}
	return reply, nil
}
