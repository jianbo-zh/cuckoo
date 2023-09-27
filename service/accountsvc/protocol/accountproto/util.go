package accountproto

import (
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

func DecodeAccount(account *pb.Account) types.Account {
	return types.Account{
		ID:                   peer.ID(account.Id),
		Name:                 account.Name,
		Avatar:               account.Avatar,
		AutoAddContact:       account.AutoAddContact,
		AutoJoinGroup:        account.AutoJoinGroup,
		AutoDepositMessage:   account.AutoDepositMessage,
		DepositAddress:       peer.ID(account.DepositAddress),
		EnableDepositService: account.EnableDepositService,
	}
}

func DecodeAccountPeer(account *pb.Account) types.AccountPeer {
	return types.AccountPeer{
		ID:             peer.ID(account.Id),
		Name:           account.Name,
		Avatar:         account.Avatar,
		DepositAddress: peer.ID(account.DepositAddress),
	}
}
