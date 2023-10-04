package groupnetworkds

import (
	"bytes"
	"context"
	"errors"
	"sync"

	ds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/datastore"
	"github.com/multiformats/go-varint"
)

var _ NetworkIface = (*NetworkDs)(nil)

var adminDsKey = &datastore.GroupDsKey{}

type NetworkDs struct {
	ds.Batching

	networkLamportMutex sync.Mutex
}

func NetworkWrap(d ds.Batching) *NetworkDs {
	return &NetworkDs{Batching: d}
}

func (n *NetworkDs) GetLamportTime(ctx context.Context, groupID string) (uint64, error) {
	n.networkLamportMutex.Lock()
	defer n.networkLamportMutex.Unlock()

	tbs, err := n.Get(ctx, adminDsKey.NetworkLamptimeKey(groupID))
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return varint.ReadUvarint(bytes.NewReader(tbs))
}

func (n *NetworkDs) SetLamportTime(ctx context.Context, groupID string, lamptime uint64) error {
	n.networkLamportMutex.Lock()
	defer n.networkLamportMutex.Unlock()

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime)

	return n.Put(ctx, adminDsKey.NetworkLamptimeKey(groupID), buff[:len])
}

func (n *NetworkDs) TickLamportTime(ctx context.Context, groupID string) (uint64, error) {
	n.networkLamportMutex.Lock()
	defer n.networkLamportMutex.Unlock()

	lamptimeKey := adminDsKey.NetworkLamptimeKey(groupID)

	lamptime := uint64(0)

	if tbs, err := n.Get(ctx, lamptimeKey); err != nil {
		if !errors.Is(err, ds.ErrNotFound) {
			return 0, err
		}

	} else if lamptime, err = varint.ReadUvarint(bytes.NewReader(tbs)); err != nil {
		return 0, err
	}

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime+1)

	if err := n.Put(ctx, lamptimeKey, buff[:len]); err != nil {
		return 0, err
	}

	return lamptime + 1, nil
}
