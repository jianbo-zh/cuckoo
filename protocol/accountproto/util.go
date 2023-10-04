package accountproto

import (
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/accountpb"
	"github.com/libp2p/go-libp2p/core/peer"
)

func DecodeAccount(account *pb.Account) mytype.Account {
	return mytype.Account{
		ID:                 peer.ID(account.Id),
		Name:               account.Name,
		Avatar:             account.Avatar,
		AutoAddContact:     account.AutoAddContact,
		AutoJoinGroup:      account.AutoJoinGroup,
		AutoDepositMessage: account.AutoDepositMessage,
		DepositAddress:     peer.ID(account.DepositAddress),
	}
}

func DecodeAccountPeer(account *pb.Account) mytype.AccountPeer {
	return mytype.AccountPeer{
		ID:             peer.ID(account.Id),
		Name:           account.Name,
		Avatar:         account.Avatar,
		DepositAddress: peer.ID(account.DepositAddress),
	}
}
