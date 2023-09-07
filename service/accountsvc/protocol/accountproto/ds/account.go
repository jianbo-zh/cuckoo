package ds

import (
	"context"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto/pb"
	"google.golang.org/protobuf/proto"
)

var _ AccountIface = (*AccountDS)(nil)

type AccountDS struct {
	ipfsds.Batching
}

func Wrap(b ipfsds.Batching) *AccountDS {
	return &AccountDS{Batching: b}
}

func (a *AccountDS) CreateAccount(ctx context.Context, info *pb.AccountMsg) error {
	value, err := proto.Marshal(info)
	if err != nil {
		return err
	}

	key := ipfsds.NewKey("/dchat/peer/account")
	return a.Put(ctx, key, value)
}

func (a *AccountDS) GetAccount(ctx context.Context) (*pb.AccountMsg, error) {
	key := ipfsds.NewKey("/dchat/peer/account")

	value, err := a.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var info pb.AccountMsg
	if err := proto.Unmarshal(value, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (a *AccountDS) UpdateAccount(ctx context.Context, info *pb.AccountMsg) error {
	value, err := proto.Marshal(info)
	if err != nil {
		return err
	}

	key := ipfsds.NewKey("/dchat/peer/account")
	return a.Put(ctx, key, value)
}
