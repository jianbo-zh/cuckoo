package ds

import (
	"bytes"
	"context"
	"errors"
	"sync"

	ds "github.com/ipfs/go-datastore"
	"github.com/multiformats/go-varint"
)

var _ NetworkIface = (*NetworkDs)(nil)

type NetworkDs struct {
	ds.Batching

	networkLamportMutex sync.Mutex
}

func NetworkWrap(d ds.Batching) *NetworkDs {
	return &NetworkDs{Batching: d}
}

func (n *NetworkDs) GetLamportTime(ctx context.Context, groupID GroupID) (uint64, error) {
	n.networkLamportMutex.Lock()
	defer n.networkLamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "network", "lamportime"})

	tbs, err := n.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return varint.ReadUvarint(bytes.NewReader(tbs))
}

func (n *NetworkDs) SetLamportTime(ctx context.Context, groupID GroupID, lamptime uint64) error {
	n.networkLamportMutex.Lock()
	defer n.networkLamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "network", "lamportime"})

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime)

	return n.Put(ctx, key, buff[:len])
}

func (n *NetworkDs) TickLamportTime(ctx context.Context, groupID GroupID) (uint64, error) {
	n.networkLamportMutex.Lock()
	defer n.networkLamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "network", "lamportime"})

	lamptime := uint64(0)

	if tbs, err := n.Get(ctx, key); err != nil {
		if !errors.Is(err, ds.ErrNotFound) {
			return 0, err
		}

	} else if lamptime, err = varint.ReadUvarint(bytes.NewReader(tbs)); err != nil {
		return 0, err
	}

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime+1)

	if err := n.Put(ctx, key, buff[:len]); err != nil {
		return 0, err
	}

	return lamptime + 1, nil
}
