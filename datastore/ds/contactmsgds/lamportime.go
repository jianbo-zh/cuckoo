package contactmsgds

import (
	"bytes"
	"context"
	"errors"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-varint"
)

// GetLamportTime 获取lamport时间
func (m *MessageDS) GetLamportTime(ctx context.Context, peerID peer.ID) (uint64, error) {
	m.lamportMutex.Lock()
	defer m.lamportMutex.Unlock()

	key := lamportKey(peerID)

	tbs, err := m.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ipfsds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return varint.ReadUvarint(bytes.NewReader(tbs))
}

// TickLamportTime lamport时间走1，并返回
func (m *MessageDS) TickLamportTime(ctx context.Context, peerID peer.ID) (uint64, error) {
	m.lamportMutex.Lock()
	defer m.lamportMutex.Unlock()

	key := lamportKey(peerID)

	lamptime := uint64(0)
	if bs, err := m.Get(ctx, key); err != nil {
		if !errors.Is(err, ipfsds.ErrNotFound) {
			return 0, err
		}

	} else if lamptime, err = varint.ReadUvarint(bytes.NewReader(bs)); err != nil {
		return 0, err
	}

	nextLamptime := lamptime + 1

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, nextLamptime)

	if err := m.Put(ctx, key, buff[:len]); err != nil {
		return 0, err
	}

	return nextLamptime, nil
}

// MergeLamportTime 合并lamport时间
func (m *MessageDS) MergeLamportTime(ctx context.Context, peerID peer.ID, lamptime uint64) error {
	m.lamportMutex.Lock()
	defer m.lamportMutex.Unlock()

	key := lamportKey(peerID)

	bs, err := m.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ipfsds.ErrNotFound) {
			// 没找到，更新时间
			buff := make([]byte, varint.MaxLenUvarint63)
			len := varint.PutUvarint(buff, lamptime)

			return m.Put(ctx, key, buff[:len])
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
		return m.Put(ctx, key, buff[:len])
	}

	return nil
}

func lamportKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.KeyWithNamespaces([]string{"dchat", "addrbook", peerID.String(), "message", "lamptime"})
}
