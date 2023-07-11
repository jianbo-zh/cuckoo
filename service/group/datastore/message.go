package datastore

import (
	"bytes"
	"context"
	"errors"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	msgpb "github.com/jianbo-zh/dchat/service/group/protocol/message/pb"
	"github.com/multiformats/go-varint"
	"google.golang.org/protobuf/proto"
)

func (gds *GroupDataStore) GetMessageLamportTime(ctx context.Context, groupID GroupID) (uint64, error) {
	gds.messageLamportMutex.Lock()
	defer gds.messageLamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "lamportime"})

	tbs, err := gds.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return varint.ReadUvarint(bytes.NewReader(tbs))
}

func (gds *GroupDataStore) SetMessageLamportTime(ctx context.Context, groupID GroupID, lamptime uint64) error {
	gds.messageLamportMutex.Lock()
	defer gds.messageLamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "lamportime"})

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime)

	return gds.Put(ctx, key, buff[:len])
}

func (gds *GroupDataStore) TickMessageLamportTime(ctx context.Context, groupID GroupID) (uint64, error) {
	gds.messageLamportMutex.Lock()
	defer gds.messageLamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "lamportime"})

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

func (gds *GroupDataStore) GetMessage(ctx context.Context, groupID GroupID, msgID string) (*msgpb.GroupMsg, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "logs", msgID})
	bs, err := gds.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var msg msgpb.GroupMsg
	if err := proto.Unmarshal(bs, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (gds *GroupDataStore) PutMessage(ctx context.Context, groupID GroupID, msg *msgpb.GroupMsg) error {

	msgPrefix := []string{"dchat", "group", string(groupID), "message"}

	key := ds.KeyWithNamespaces(append(msgPrefix, "logs", msg.Id))
	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	batch, err := gds.Batch(ctx)
	if err != nil {
		return err
	}

	if err := batch.Put(ctx, key, bs); err != nil {
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

func (gds *GroupDataStore) ListMessages(ctx context.Context, groupID GroupID) ([]*msgpb.GroupMsg, error) {

	results, err := gds.Query(ctx, query.Query{
		Prefix: "/dchat/group/" + string(groupID) + "/message/logs",
		Orders: []query.Order{query.OrderByKeyDescending{}},
	})
	if err != nil {
		return nil, err
	}

	var msgs []*msgpb.GroupMsg
	for result := range results.Next() {
		if result.Error != nil {
			return nil, err
		}

		var msg msgpb.GroupMsg
		if err := proto.Unmarshal(result.Entry.Value, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (gds *GroupDataStore) GetMessageHeadID(ctx context.Context, groupID GroupID) (string, error) {
	headKey := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "head"})
	head, err := gds.Get(ctx, headKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", err
	}

	return string(head), nil
}
func (gds *GroupDataStore) GetMessageTailID(ctx context.Context, groupID GroupID) (string, error) {
	tailKey := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "tail"})
	tail, err := gds.Get(ctx, tailKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", err
	}

	return string(tail), nil
}

func (gds *GroupDataStore) GetMessageLength(context.Context, GroupID) (int32, error) {
	return 0, nil
}
