package ds

import (
	"bytes"
	"context"
	"errors"

	ds "github.com/ipfs/go-datastore"
	"github.com/multiformats/go-varint"
)

func (a *AdminDs) GetLamportTime(ctx context.Context, groupID string) (uint64, error) {
	a.lamportMutex.Lock()
	defer a.lamportMutex.Unlock()

	key := lamportKey(groupID)

	tbs, err := a.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return varint.ReadUvarint(bytes.NewReader(tbs))
}

func (a *AdminDs) TickLamportTime(ctx context.Context, groupID string) (uint64, error) {
	a.lamportMutex.Lock()
	defer a.lamportMutex.Unlock()

	key := lamportKey(groupID)

	lamptime := uint64(0)

	if tbs, err := a.Get(ctx, key); err != nil {

		if !errors.Is(err, ds.ErrNotFound) {
			return 0, err
		}

	} else if lamptime, err = varint.ReadUvarint(bytes.NewReader(tbs)); err != nil {
		return 0, err
	}

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime+1)

	if err := a.Put(ctx, key, buff[:len]); err != nil {
		return 0, err
	}

	return lamptime + 1, nil
}

func (a *AdminDs) MergeLamportTime(ctx context.Context, groupID string, lamptime uint64) error {
	a.lamportMutex.Lock()
	defer a.lamportMutex.Unlock()

	key := lamportKey(groupID)

	bs, err := a.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) { // 没找到，更新时间
			buff := make([]byte, varint.MaxLenUvarint63)
			len := varint.PutUvarint(buff, lamptime)

			return a.Put(ctx, key, buff[:len])
		}

		return err
	}

	lamptimeCur, err := varint.ReadUvarint(bytes.NewReader(bs))
	if err != nil {
		return err
	}

	if lamptime > lamptimeCur {
		// 比当前大，更新时间
		buff := make([]byte, varint.MaxLenUvarint63)
		len := varint.PutUvarint(buff, lamptime)
		return a.Put(ctx, key, buff[:len])
	}

	return nil
}

func lamportKey(groupID string) ds.Key {
	return ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "lamportime"})
}
