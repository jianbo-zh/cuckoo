package datastore

import (
	ipfsds "github.com/ipfs/go-datastore"
)

type DepositDataStore struct {
	ipfsds.Batching
}

func DepositWrap(d ipfsds.Batching) *DepositDataStore {
	return &DepositDataStore{Batching: d}
}
