package ds

import (
	"context"

	ds "github.com/ipfs/go-datastore"
)

type NetworkIface interface {
	ds.Batching

	GetLamportTime(context.Context, GroupID) (uint64, error)
	SetLamportTime(context.Context, GroupID, uint64) error
	TickLamportTime(context.Context, GroupID) (uint64, error)
}
