package accountds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/datastore"
	pb "github.com/jianbo-zh/dchat/service/accountsvc/protobuf/pb/accountpb"
	"google.golang.org/protobuf/proto"
)

var _ AccountIface = (*AccountDS)(nil)

var accountDsKey = &datastore.AccountDsKey{}

type AccountDS struct {
	ipfsds.Batching
}

func Wrap(b ipfsds.Batching) *AccountDS {
	return &AccountDS{Batching: b}
}

func (a *AccountDS) CreateAccount(ctx context.Context, info *pb.Account) error {
	value, err := proto.Marshal(info)
	if err != nil {
		return err
	}

	return a.Put(ctx, accountDsKey.Key(), value)
}

func (a *AccountDS) GetAccount(ctx context.Context) (*pb.Account, error) {

	value, err := a.Get(ctx, accountDsKey.Key())
	if err != nil {
		return nil, err
	}

	var info pb.Account
	if err := proto.Unmarshal(value, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (a *AccountDS) UpdateAccount(ctx context.Context, info *pb.Account) error {
	value, err := proto.Marshal(info)
	if err != nil {
		return err
	}

	return a.Put(ctx, accountDsKey.Key(), value)
}
