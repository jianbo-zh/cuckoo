package ds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto/pb"
)

type AccountIface interface {
	ipfsds.Batching

	CreateAccount(context.Context, *pb.AccountMsg) error
	GetAccount(context.Context) (*pb.AccountMsg, error)
	UpdateAccount(context.Context, *pb.AccountMsg) error
}
