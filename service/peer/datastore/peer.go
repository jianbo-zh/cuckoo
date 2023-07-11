package datastore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	msgpb "github.com/jianbo-zh/dchat/service/peer/protocol/message/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-varint"
	"google.golang.org/protobuf/proto"
)

// peer 相关数据存储操作及接口

// 保存消息
// 消息列表

type PeerIface interface {
	ds.Batching

	StoreMessage(context.Context, peer.ID, *msgpb.Message) error
	GetMessages(context.Context, peer.ID) ([]*msgpb.Message, error)

	GetLamportTime(context.Context, peer.ID) (uint64, error)
	SetLamportTime(context.Context, peer.ID, uint64) error
	TickLamportTime(context.Context, peer.ID) (uint64, error)
}

var _ PeerIface = (*PeerDataStore)(nil)

type PeerDataStore struct {
	ds.Batching

	lamportMutex sync.Mutex
}

func PeerWrap(d ds.Batching) *PeerDataStore {
	return &PeerDataStore{Batching: d}
}

func (gds *PeerDataStore) StoreMessage(ctx context.Context, peerID peer.ID, msg *msgpb.Message) error {

	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	batch, err := gds.Batch(ctx)
	if err != nil {
		return err
	}

	msgPrefix := []string{"dchat", "addrbook", peerID.String(), "message"}

	msgKey := ds.KeyWithNamespaces(append(msgPrefix, "logs", msg.Id))
	if err = batch.Put(ctx, msgKey, bs); err != nil {
		return err
	}

	headKey := ds.KeyWithNamespaces(append(msgPrefix, "head"))
	head, err := gds.Get(ctx, headKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}

	if len(head) == 0 {
		if err = batch.Put(ctx, headKey, []byte(msg.Id)); err != nil {
			return err
		}
	}

	tailKey := ds.KeyWithNamespaces(append(msgPrefix, "tail"))
	if err = batch.Put(ctx, tailKey, []byte(msg.Id)); err != nil {
		return err
	}

	return batch.Commit(ctx)
}

func (gds *PeerDataStore) GetMessages(ctx context.Context, peerID peer.ID) ([]*msgpb.Message, error) {
	results, err := gds.Query(ctx, query.Query{
		Prefix: fmt.Sprintf("/dchat/addrbook/%s/message/logs", peerID.String()),
		Orders: []query.Order{query.OrderByKey{}},
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		return nil, err
	}

	var msgs []*msgpb.Message
	for result := range results.Next() {
		if result.Error != nil {
			return nil, err
		}

		var msg msgpb.Message
		err = proto.Unmarshal(result.Entry.Value, &msg)
		if err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (gds *PeerDataStore) GetLamportTime(ctx context.Context, peerID peer.ID) (uint64, error) {
	gds.lamportMutex.Lock()
	defer gds.lamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "addrbook", peerID.String(), "message", "lamptime"})

	tbs, err := gds.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return varint.ReadUvarint(bytes.NewReader(tbs))
}

func (gds *PeerDataStore) SetLamportTime(ctx context.Context, peerID peer.ID, lamptime uint64) error {
	gds.lamportMutex.Lock()
	defer gds.lamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "addrbook", peerID.String(), "message", "lamptime"})

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime+1)

	return gds.Put(ctx, key, buff[:len])
}

func (gds *PeerDataStore) TickLamportTime(ctx context.Context, peerID peer.ID) (uint64, error) {
	gds.lamportMutex.Lock()
	defer gds.lamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "addrbook", peerID.String(), "message", "lamptime"})

	lamptime := uint64(0)

	if tbs, err := gds.Get(ctx, key); err != nil {

		if !errors.Is(err, ds.ErrNotFound) {
			return 0, err
		}

	} else if lamptime, err = varint.ReadUvarint(bytes.NewReader(tbs)); err != nil {
		return 0, err
	}

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime+1)

	if err := gds.Put(ctx, key, buff[:len]); err != nil {
		return 0, err
	}

	return lamptime + 1, nil
}
