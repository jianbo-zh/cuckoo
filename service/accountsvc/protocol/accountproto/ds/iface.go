package ds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto/pb"
)

type AccountIface interface {
	ipfsds.Batching

	CreateAccount(context.Context, *pb.Account) error
	GetAccount(context.Context) (*pb.Account, error)
	UpdateAccount(context.Context, *pb.Account) error
}
