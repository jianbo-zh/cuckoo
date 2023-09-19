package ds

import (
	"context"
	"errors"
	"fmt"
	"sync"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/internal/datastore"
	msgpb "github.com/jianbo-zh/dchat/service/groupsvc/protocol/messageproto/pb"
	"google.golang.org/protobuf/proto"
)

var _ MessageIface = (*MessageDs)(nil)

var adminDsKey = &datastore.GroupDsKey{}

type MessageDs struct {
	ds.Batching

	messageLamportMutex sync.Mutex
}

func MessageWrap(d ds.Batching) *MessageDs {
	return &MessageDs{Batching: d}
}

func (m *MessageDs) GetMessage(ctx context.Context, groupID GroupID, msgID string) (*msgpb.Message, error) {
	bs, err := m.Get(ctx, adminDsKey.MsgLogKey(groupID, msgID))
	if err != nil {
		return nil, err
	}

	var msg msgpb.Message
	if err := proto.Unmarshal(bs, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (m *MessageDs) SaveMessage(ctx context.Context, groupID GroupID, msg *msgpb.Message) error {

	batch, err := m.Batch(ctx)
	if err != nil {
		return err
	}

	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	if err := batch.Put(ctx, adminDsKey.MsgLogKey(groupID, msg.Id), bs); err != nil {
		return err
	}

	head, err := m.Get(ctx, adminDsKey.MsgLogHeadKey(groupID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}

	if len(head) == 0 {
		if err = batch.Put(ctx, adminDsKey.MsgLogHeadKey(groupID), []byte(msg.Id)); err != nil {
			return err
		}
	}

	if err = batch.Put(ctx, adminDsKey.MsgLogTailKey(groupID), []byte(msg.Id)); err != nil {
		return err
	}

	return batch.Commit(ctx)
}

func (m *MessageDs) GetMessages(ctx context.Context, groupID GroupID, offset int, limit int) ([]*msgpb.Message, error) {

	results, err := m.Query(ctx, query.Query{
		Prefix: adminDsKey.MsgLogPrefix(groupID),
		Orders: []query.Order{query.OrderByKeyDescending{}},
		Offset: offset,
		Limit:  limit,
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

	return reverse(msgs), nil
}

func (m *MessageDs) ClearMessage(ctx context.Context, groupID GroupID) error {
	results, err := m.Query(ctx, query.Query{
		Prefix: adminDsKey.MsgPrefix(groupID),
	})
	if err != nil {
		return fmt.Errorf("ds query error: %w", err)
	}

	for result := range results.Next() {
		if result.Error != nil {
			return fmt.Errorf("ds results next error: %w", err)
		}

		if err = m.Delete(ctx, ds.NewKey(result.Key)); err != nil {
			return fmt.Errorf("ds delete key error: %w", err)
		}
	}

	return nil
}

func (m *MessageDs) GetMessageHead(ctx context.Context, groupID GroupID) (string, error) {
	head, err := m.Get(ctx, adminDsKey.MsgLogHeadKey(groupID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", err
	}

	return string(head), nil
}
func (m *MessageDs) GetMessageTail(ctx context.Context, groupID GroupID) (string, error) {
	tail, err := m.Get(ctx, adminDsKey.MsgLogTailKey(groupID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", err
	}

	return string(tail), nil
}

func (m *MessageDs) GetMessageLength(context.Context, GroupID) (int32, error) {
	return 0, nil
}

func reverse[T any](s []T) []T {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
