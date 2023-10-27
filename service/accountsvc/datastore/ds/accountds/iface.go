package accountds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	pb "github.com/jianbo-zh/dchat/service/accountsvc/protobuf/pb/accountpb"
)

type AccountIface interface {
	ipfsds.Batching

	CreateAccount(context.Context, *pb.Account) error
	GetAccount(context.Context) (*pb.Account, error)
	UpdateAccount(context.Context, *pb.Account) error
}
