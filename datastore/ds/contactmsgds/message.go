package contactmsgds

import (
	"context"
	"errors"
	"fmt"
	"sync"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/internal/datastore"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/contactpb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

var _ PeerMessageIface = (*MessageDS)(nil)

var contactDsKey = &datastore.ContactDsKey{}

type msgStats struct {
	HeadID string
	TailID string
}

type MessageDS struct {
	ipfsds.Batching

	msgCache      map[peer.ID]msgStats
	msgCacheMutex sync.Mutex

	lamportMutex sync.Mutex
}

func Wrap(b ipfsds.Batching) *MessageDS {
	return &MessageDS{
		Batching: b,
		msgCache: make(map[peer.ID]msgStats),
	}
}

// HasMessage 消息是否存在
func (m *MessageDS) HasMessage(ctx context.Context, peerID peer.ID, msgID string) (bool, error) {
	return m.Has(ctx, contactDsKey.MsgLogKey(peerID, msgID))
}

// SaveMessage 保存消息
func (m *MessageDS) SaveMessage(ctx context.Context, peerID peer.ID, msg *pb.ContactMessage) (isLatest bool, err error) {

	bs, err := proto.Marshal(msg)
	if err != nil {
		return false, err
	}

	batch, err := m.Batch(ctx)
	if err != nil {
		return false, err
	}

	if err = batch.Put(ctx, contactDsKey.MsgLogKey(peerID, msg.Id), bs); err != nil {
		return false, err
	}

	m.msgCacheMutex.Lock()
	defer m.msgCacheMutex.Unlock()

	var headID, tailID string
	if _, exists := m.msgCache[peerID]; !exists {
		headVal, err := m.Get(ctx, contactDsKey.MsgHeadKey(peerID))
		if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
			return false, err
		}

		tailVal, err := m.Get(ctx, contactDsKey.MsgHeadKey(peerID))
		if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
			return false, err
		}

		headID = string(headVal)
		tailID = string(tailVal)

	} else {
		headID = m.msgCache[peerID].HeadID
		tailID = m.msgCache[peerID].TailID
	}

	if len(headID) == 0 || msg.Id < headID {
		if err = batch.Put(ctx, contactDsKey.MsgHeadKey(peerID), []byte(msg.Id)); err != nil {
			return false, err
		}
		headID = msg.Id
	}

	if len(tailID) == 0 || msg.Id > tailID {
		if err = batch.Put(ctx, contactDsKey.MsgTailKey(peerID), []byte(msg.Id)); err != nil {
			return false, err
		}
		tailID = msg.Id
		isLatest = true
	}

	m.msgCache[peerID] = msgStats{
		HeadID: headID,
		TailID: tailID,
	}

	if err = batch.Commit(ctx); err != nil {
		return false, fmt.Errorf("batch.Commit error: %w", err)
	}

	return isLatest, nil
}

func (m *MessageDS) GetMessage(ctx context.Context, peerID peer.ID, msgID string) (*pb.ContactMessage, error) {

	val, err := m.Get(ctx, contactDsKey.MsgLogKey(peerID, msgID))
	if err != nil {
		return nil, fmt.Errorf("m.Get error: %w", err)
	}

	var msg pb.ContactMessage
	err = proto.Unmarshal(val, &msg)
	if err != nil {
		return nil, fmt.Errorf("proto.Unmarshal error: %w", err)
	}

	return &msg, nil
}

func (m *MessageDS) UpdateMessageState(ctx context.Context, peerID peer.ID, msgID string, isSucc bool) (*pb.ContactMessage, error) {

	val, err := m.Get(ctx, contactDsKey.MsgLogKey(peerID, msgID))
	if err != nil {
		return nil, fmt.Errorf("m.Get error: %w", err)
	}

	var msg pb.ContactMessage
	err = proto.Unmarshal(val, &msg)
	if err != nil {
		return nil, fmt.Errorf("proto.Unmarshal error: %w", err)
	}

	if isSucc {
		msg.State = pb.ContactMessage_Success
	} else {
		msg.State = pb.ContactMessage_Fail
	}

	bs, err := proto.Marshal(&msg)
	if err != nil {
		return nil, fmt.Errorf("proto marshal error: %w", err)
	}
	if err := m.Put(ctx, contactDsKey.MsgLogKey(peerID, msgID), bs); err != nil {
		return nil, fmt.Errorf("ds put key error: %w", err)
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
func (m *MessageDS) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]*pb.ContactMessage, error) {
	results, err := m.Query(ctx, query.Query{
		Prefix: contactDsKey.MsgLogPrefix(peerID),
		Orders: []query.Order{query.OrderByKeyDescending{}},
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	var msgs []*pb.ContactMessage
	for result := range results.Next() {
		if result.Error != nil {
			return nil, err
		}

		var msg pb.ContactMessage
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
