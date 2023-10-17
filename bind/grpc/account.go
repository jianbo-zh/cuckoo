package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/internal/util"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/libp2p/go-libp2p/core/peer"
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

func (a *AccountSvc) GetAccountQRCodeToken(ctx context.Context, request *proto.GetAccountQRCodeTokenRequest) (reply *proto.GetAccountQRCodeTokenReply, err error) {

	log.Infoln("GetAccountQRCodeToken request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetAccountQRCodeToken panic: ", e)
		} else if err != nil {
			log.Errorln("GetAccountQRCodeToken error: ", err.Error())
		} else {
			log.Infoln("GetAccountQRCodeToken reply: ", reply.String())
		}
	}()

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("get account svc error: %w", err)
	}

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc.GetAccount error: %w", err)
	}

	QRCodeToken := util.EncodePeerToken(account.ID, account.DepositAddress, account.Name, account.Avatar)

	reply = &proto.GetAccountQRCodeTokenReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Token: QRCodeToken,
	}
	return reply, nil
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

	cuckoo, err := a.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	accountSvc, err := cuckoo.GetAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetAccountSvc error: %s", err.Error())
	}

	fileSvc, err := cuckoo.GetFileSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetFileSvc error: %s", err.Error())
	}

	avatarID, err := fileSvc.CopyFileToResource(ctx, request.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("svc.CopyFileToResource error: %w", err)
	}

	fullAccount, err := accountSvc.CreateAccount(ctx, mytype.Account{
		Name:           request.Name,
		Avatar:         avatarID,
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
	if err != nil {
		return nil, fmt.Errorf("svc get account error: %w", err)
	}

	protoAccount = proto.Account{
		PeerId:             account.ID.String(),
		Name:               account.Name,
		Avatar:             account.Avatar,
		AutoAddContact:     account.AutoAddContact,
		AutoJoinGroup:      account.AutoJoinGroup,
		AutoDepositMessage: account.AutoDepositMessage,
		DepositAddress:     account.DepositAddress.String(),
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

	cuckoo, err := a.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	accountSvc, err := cuckoo.GetAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetAccountSvc error: %s", err.Error())
	}

	fileSvc, err := cuckoo.GetFileSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetFileSvc error: %s", err.Error())
	}

	avatarID, err := fileSvc.CopyFileToResource(ctx, request.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("svc.CopyFileToResource error: %w", err)
	}

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get account error: %w", err)
	}

	if avatarID != account.Avatar {
		err = accountSvc.SetAccountAvatar(ctx, avatarID)
		if err != nil {
			return nil, fmt.Errorf("accountSvc.SetAvatar error: %w", err)
		}
	}

	reply = &proto.SetAccountAvatarReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Avatar: avatarID,
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

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get account error: %w", err)
	}

	if strings.TrimSpace(request.GetName()) != account.Name {
		err = accountSvc.SetAccountName(ctx, strings.TrimSpace(request.GetName()))
		if err != nil {
			return nil, fmt.Errorf("svc set name error: %w", err)
		}
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

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get account error: %w", err)
	}

	if request.GetIsAuto() != account.AutoAddContact {
		err = accountSvc.SetAccountAutoAddContact(ctx, request.GetIsAuto())
		if err != nil {
			return nil, fmt.Errorf("svc set auto add contact error: %w", err)
		}
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

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get account error: %w", err)
	}

	if request.GetIsAuto() != account.AutoJoinGroup {
		err = accountSvc.SetAccountAutoJoinGroup(ctx, request.GetIsAuto())
		if err != nil {
			return nil, fmt.Errorf("accountSvc.SetAutoJoinGroup error: %w", err)
		}
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

func (a *AccountSvc) SetAutoDepositMessage(ctx context.Context, request *proto.SetAutoDepositMessageRequest) (reply *proto.SetAutoDepositMessageReply, err error) {

	log.Infoln("SetAutoDepositMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetAutoDepositMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SetAutoDepositMessage error: ", err.Error())
		} else {
			log.Infoln("SetAutoDepositMessage reply: ", reply.String())
		}
	}()

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get account error: %w", err)
	}

	if request.GetIsAuto() != account.AutoDepositMessage {
		err = accountSvc.SetAutoDepositMessage(ctx, request.GetIsAuto())
		if err != nil {
			return nil, fmt.Errorf("svc set auto send deposit error: %w", err)
		}
	}

	reply = &proto.SetAutoDepositMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		IsAuto: request.GetIsAuto(),
	}
	return reply, nil
}

func (a *AccountSvc) SetAccountDepositAddress(ctx context.Context, request *proto.SetAccountDepositAddressRequest) (reply *proto.SetAccountDepositAddressReply, err error) {

	log.Infoln("SetAccountDepositAddress request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetAccountDepositAddress panic: ", e)
		} else if err != nil {
			log.Errorln("SetAccountDepositAddress error: ", err.Error())
		} else {
			log.Infoln("SetAccountDepositAddress reply: ", reply.String())
		}
	}()

	accountSvc, err := a.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("a.getAccountSvc error: %w", err)
	}

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get account error: %w", err)
	}

	var peerID peer.ID
	if strings.TrimSpace(request.DepositAddress) != "" {
		peerID, err = peer.Decode(strings.TrimSpace(request.DepositAddress))
		if err != nil || peerID.Validate() != nil {
			return nil, fmt.Errorf("peer decode error")
		}
	}

	if peerID != account.DepositAddress {
		err = accountSvc.SetAccountDepositAddress(ctx, peerID)
		if err != nil {
			return nil, fmt.Errorf("svc set deposit address error: %w", err)
		}
	}

	reply = &proto.SetAccountDepositAddressReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		DepositAddress: peerID.String(),
	}
	return reply, nil
}
