package groupnetworkds

import (
	"context"

	ds "github.com/ipfs/go-datastore"
)

type NetworkIface interface {
	ds.Batching

	GetLamportTime(ctx context.Context, groupID string) (uint64, error)
	SetLamportTime(ctx context.Context, groupID string, lamptime uint64) error
	TickLamportTime(ctx context.Context, groupID string) (uint64, error)
}
