package service

import (
	"context"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
)

var _ proto.AccountSvcServer = (*AccountSvc)(nil)

type AccountSvc struct {
	proto.UnimplementedAccountSvcServer
}

func (a *AccountSvc) CreateAccount(ctx context.Context, request *proto.CreateAccountRequest) (*proto.CreateAccountReply, error) {
	account = proto.Account{
		PeerID:                  "peerID",
		Avatar:                  request.GetAvatar(),
		Name:                    request.GetName(),
		AddContactWithoutReview: true,
		JoinGroupWithoutReview:  true,
	}

	reply := &proto.CreateAccountReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}
func (a *AccountSvc) GetAccount(ctx context.Context, request *proto.GetAccountRequest) (*proto.GetAccountReply, error) {

	reply := &proto.GetAccountReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Account: &account,
	}

	return reply, nil
}
func (a *AccountSvc) SetAccountAvatar(ctx context.Context, request *proto.SetAccountAvatarRequest) (*proto.SetAccountAvatarReply, error) {

	account.Avatar = request.GetAvatar()

	reply := &proto.SetAccountAvatarReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Avatar: account.Avatar,
	}
	return reply, nil
}
func (a *AccountSvc) SetAccountName(ctx context.Context, request *proto.SetAccountNameRequest) (*proto.SetAccountNameReply, error) {

	account.Name = request.GetName()

	reply := &proto.SetAccountNameReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Name: account.Name,
	}
	return reply, nil
}
func (a *AccountSvc) SetAutoReviewAddContact(ctx context.Context, request *proto.SetAutoReviewAddContactRequest) (*proto.SetAutoReviewAddContactReply, error) {

	account.AddContactWithoutReview = request.GetIsReview()

	reply := &proto.SetAutoReviewAddContactReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		IsReview: account.AddContactWithoutReview,
	}
	return reply, nil
}
func (a *AccountSvc) SetAutoReviewJoinGroup(ctx context.Context, request *proto.SetAutoReviewJoinGroupRequest) (*proto.SetAutoReviewJoinGroupReply, error) {

	account.JoinGroupWithoutReview = request.GetIsReview()

	reply := &proto.SetAutoReviewJoinGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		IsReview: account.JoinGroupWithoutReview,
	}
	return reply, nil
}
