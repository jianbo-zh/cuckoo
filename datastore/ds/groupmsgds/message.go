package groupmsgds

import (
	"context"
	"errors"
	"fmt"
	"sync"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/internal/datastore"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/grouppb"
	"google.golang.org/protobuf/proto"
)

var _ MessageIface = (*MessageDs)(nil)

var adminDsKey = &datastore.GroupDsKey{}

type msgStats struct {
	HeadID string
	TailID string
}

type MessageDs struct {
	ipfsds.Batching

	msgCache      map[string]msgStats
	msgCacheMutex sync.Mutex

	messageLamportMutex sync.Mutex
}

func MessageWrap(d ipfsds.Batching) *MessageDs {
	return &MessageDs{
		Batching: d,
		msgCache: make(map[string]msgStats),
	}
}

func (m *MessageDs) HasMessage(ctx context.Context, groupID string, msgID string) (bool, error) {
	return m.Has(ctx, adminDsKey.MsgLogKey(groupID, msgID))
}

func (m *MessageDs) GetMessage(ctx context.Context, groupID string, msgID string) (*pb.GroupMessage, error) {
	bs, err := m.Get(ctx, adminDsKey.MsgLogKey(groupID, msgID))
	if err != nil {
		return nil, err
	}

	var msg pb.GroupMessage
	if err := proto.Unmarshal(bs, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (m *MessageDs) GetCoreMessage(ctx context.Context, groupID string, msgID string) (*pb.CoreMessage, error) {
	val, err := m.Get(ctx, adminDsKey.MsgLogKey(groupID, msgID))
	if err != nil {
		return nil, fmt.Errorf("m.Get error: %w", err)
	}

	var msg pb.GroupMessage
	err = proto.Unmarshal(val, &msg)
	if err != nil {
		return nil, fmt.Errorf("proto.Unmarshal error: %w", err)

	} else if msg.CoreMessage == nil {
		return nil, fmt.Errorf("msg.RawMessage nil")
	}

	return msg.CoreMessage, nil
}

func (m *MessageDs) GetMessageData(ctx context.Context, groupID string, msgID string) ([]byte, error) {
	bs, err := m.Get(ctx, adminDsKey.MsgLogKey(groupID, msgID))
	if err != nil {
		return nil, fmt.Errorf("ds get key error: %w", err)
	}

	return bs, nil
}

func (m *MessageDs) DeleteMessage(ctx context.Context, groupID string, msgID string) error {
	err := m.Delete(ctx, adminDsKey.MsgLogKey(groupID, msgID))
	if err != nil {
		return fmt.Errorf("ds delete key error: %w", err)
	}

	return nil
}

func (m *MessageDs) UpdateMessageSendState(ctx context.Context, groupID string, msgID string, isDeposit bool, isSucc bool) (*pb.GroupMessage, error) {
	val, err := m.Get(ctx, adminDsKey.MsgLogKey(groupID, msgID))
	if err != nil {
		return nil, fmt.Errorf("m.Get error: %w", err)
	}

	var msg pb.GroupMessage
	err = proto.Unmarshal(val, &msg)
	if err != nil {
		return nil, fmt.Errorf("proto.Unmarshal error: %w", err)
	}

	msg.IsDeposit = isDeposit

	if isSucc {
		msg.SendState = pb.GroupMessage_SendSucc
	} else {
		msg.SendState = pb.GroupMessage_SendFail
	}

	bs, err := proto.Marshal(&msg)
	if err != nil {
		return nil, fmt.Errorf("proto marshal error: %w", err)
	}
	if err := m.Put(ctx, adminDsKey.MsgLogKey(groupID, msgID), bs); err != nil {
		return nil, fmt.Errorf("ds put key error: %w", err)
	}

	return &msg, nil
}

func (m *MessageDs) SaveMessage(ctx context.Context, groupID string, msg *pb.GroupMessage) (isLatest bool, err error) {

	batch, err := m.Batch(ctx)
	if err != nil {
		return false, err
	}

	bs, err := proto.Marshal(msg)
	if err != nil {
		return false, err
	}
	if err := batch.Put(ctx, adminDsKey.MsgLogKey(groupID, msg.Id), bs); err != nil {
		return false, err
	}

	m.msgCacheMutex.Lock()
	defer m.msgCacheMutex.Unlock()

	var headID, tailID string
	if _, exists := m.msgCache[groupID]; !exists {
		headVal, err := m.Get(ctx, adminDsKey.MsgLogHeadKey(groupID))
		if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
			return false, err
		}

		tailVal, err := m.Get(ctx, adminDsKey.MsgLogTailKey(groupID))
		if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
			return false, err
		}

		headID = string(headVal)
		tailID = string(tailVal)

	} else {
		headID = m.msgCache[groupID].HeadID
		tailID = m.msgCache[groupID].TailID
	}

	if len(headID) == 0 || msg.Id < headID {
		if err = batch.Put(ctx, adminDsKey.MsgLogHeadKey(groupID), []byte(msg.Id)); err != nil {
			return false, err
		}
		headID = msg.Id
	}

	if len(tailID) == 0 || msg.Id > tailID {
		if err = batch.Put(ctx, adminDsKey.MsgLogTailKey(groupID), []byte(msg.Id)); err != nil {
			return false, err
		}
		tailID = msg.Id
		isLatest = true
	}

	m.msgCache[groupID] = msgStats{
		HeadID: headID,
		TailID: tailID,
	}

	if err = batch.Commit(ctx); err != nil {
		return false, fmt.Errorf("batch.Commit error: %w", err)
	}

	return isLatest, nil
}

func (m *MessageDs) GetMessages(ctx context.Context, groupID string, offset int, limit int) ([]*pb.GroupMessage, error) {

	results, err := m.Query(ctx, query.Query{
		Prefix: adminDsKey.MsgLogPrefix(groupID),
		Orders: []query.Order{query.OrderByKeyDescending{}},
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	var msgs []*pb.GroupMessage
	for result := range results.Next() {
		if result.Error != nil {
			return nil, err
		}

		var msg pb.GroupMessage
		err = proto.Unmarshal(result.Entry.Value, &msg)
		if err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return reverse(msgs), nil
}

func (m *MessageDs) ClearMessage(ctx context.Context, groupID string) error {
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

		if err = m.Delete(ctx, ipfsds.NewKey(result.Key)); err != nil {
			return fmt.Errorf("ds delete key error: %w", err)
		}
	}

	return nil
}

func (m *MessageDs) GetMessageHead(ctx context.Context, groupID string) (string, error) {
	head, err := m.Get(ctx, adminDsKey.MsgLogHeadKey(groupID))
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return "", err
	}

	return string(head), nil
}
func (m *MessageDs) GetMessageTail(ctx context.Context, groupID string) (string, error) {
	tail, err := m.Get(ctx, adminDsKey.MsgLogTailKey(groupID))
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return "", err
	}

	return string(tail), nil
}

func (m *MessageDs) GetMessageLength(ctx context.Context, groupID string) (int32, error) {
	return 0, nil
}

func reverse[T any](s []T) []T {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
