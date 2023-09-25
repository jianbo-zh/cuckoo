package ds

import (
	"context"
	"errors"
	"fmt"
	"sync"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/internal/datastore"
	msgpb "github.com/jianbo-zh/dchat/service/contactsvc/protocol/messageproto/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

var _ PeerMessageIface = (*MessageDS)(nil)

var contactDsKey = &datastore.ContactDsKey{}

type MessageDS struct {
	ipfsds.Batching

	lamportMutex sync.Mutex
}

func Wrap(b ipfsds.Batching) *MessageDS {
	return &MessageDS{Batching: b}
}

// HasMessage 消息是否存在
func (m *MessageDS) HasMessage(ctx context.Context, peerID peer.ID, msgID string) (bool, error) {
	return m.Has(ctx, contactDsKey.MsgLogKey(peerID, msgID))
}

// SaveMessage 保存消息
func (m *MessageDS) SaveMessage(ctx context.Context, peerID peer.ID, msg *msgpb.Message) error {

	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	batch, err := m.Batch(ctx)
	if err != nil {
		return err
	}

	if err = batch.Put(ctx, contactDsKey.MsgLogKey(peerID, msg.Id), bs); err != nil {
		return err
	}

	head, err := m.Get(ctx, contactDsKey.MsgHeadKey(peerID))
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return err
	}

	if len(head) == 0 {
		if err = batch.Put(ctx, contactDsKey.MsgHeadKey(peerID), []byte(msg.Id)); err != nil {
			return err
		}
	}

	if err = batch.Put(ctx, contactDsKey.MsgTailKey(peerID), []byte(msg.Id)); err != nil {
		return err
	}

	return batch.Commit(ctx)
}

func (m *MessageDS) GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*msgpb.Message, error) {

	val, err := m.Get(ctx, contactDsKey.MsgLogKey(peerID, msgID))
	if err != nil {
		return nil, fmt.Errorf("m.Get error: %w", err)
	}

	var msg msgpb.Message
	err = proto.Unmarshal(val, &msg)
	if err != nil {
		return nil, fmt.Errorf("proto.Unmarshal error: %w", err)
	}

	return &msg, nil
}

func (m *MessageDS) DeleteMessage(ctx context.Context, peerID peer.ID, msgID string) error {
	return m.Delete(ctx, contactDsKey.MsgLogKey(peerID, msgID))
}

func (m *MessageDS) GetMessageData(ctx context.Context, peerID peer.ID, msgID string) ([]byte, error) {

	val, err := m.Get(ctx, contactDsKey.MsgLogKey(peerID, msgID))
	if err != nil {
		return nil, fmt.Errorf("ds get error: %w", err)
	}

	return val, nil
}

// GetMessages 获取消息列表
func (m *MessageDS) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]*msgpb.Message, error) {
	results, err := m.Query(ctx, query.Query{
		Prefix: contactDsKey.MsgLogPrefix(peerID),
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

func (m *MessageDS) ClearMessage(ctx context.Context, peerID peer.ID) error {
	results, err := m.Query(ctx, query.Query{
		Prefix:   contactDsKey.MsgPrefix(peerID),
		KeysOnly: true,
	})
	if err != nil {
		return fmt.Errorf("ds query error: %w", err)
	}

	for result := range results.Next() {
		if result.Error != nil {
			return fmt.Errorf("results next error: %w", result.Error)
		}

		if err = m.Delete(ctx, ipfsds.NewKey(result.Key)); err != nil {
			return fmt.Errorf("ds delete key error: %w", err)
		}
	}

	return nil
}

func reverse[T any](s []T) []T {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
