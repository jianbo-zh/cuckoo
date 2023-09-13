package ds

import (
	"context"
	"errors"
	"fmt"
	"sync"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	msgpb "github.com/jianbo-zh/dchat/service/groupsvc/protocol/message/pb"
	"google.golang.org/protobuf/proto"
)

var _ MessageIface = (*MessageDs)(nil)

type MessageDs struct {
	ds.Batching

	messageLamportMutex sync.Mutex
}

func MessageWrap(d ds.Batching) *MessageDs {
	return &MessageDs{Batching: d}
}

func (m *MessageDs) GetMessage(ctx context.Context, groupID GroupID, msgID string) (*msgpb.Message, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "logs", msgID})
	bs, err := m.Get(ctx, key)
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

	fmt.Println("save message groupid: ", groupID, "msgid: ", msg.Id)

	batch, err := m.Batch(ctx)
	if err != nil {
		return err
	}

	msgPrefix := []string{"dchat", "group", string(groupID), "message"}
	msgKey := ds.KeyWithNamespaces(append(msgPrefix, "logs", msg.Id))
	fmt.Println("msgkey: ", msgKey.String())
	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	if err := batch.Put(ctx, msgKey, bs); err != nil {
		return err
	}

	headKey := ds.KeyWithNamespaces(append(msgPrefix, "head"))
	head, err := m.Get(ctx, headKey)
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

func (m *MessageDs) ListMessages(ctx context.Context, groupID GroupID, offset int, limit int) ([]*msgpb.Message, error) {

	fmt.Printf("list messages groupId:%s offset:%d limit:%d\n", groupID, offset, limit)

	fmt.Println("prefix: ", "/dchat/group/"+string(groupID)+"/message/logs")
	results, err := m.Query(ctx, query.Query{
		Prefix: "/dchat/group/" + string(groupID) + "/message/logs",
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

		fmt.Println("adf2222")

		var msg msgpb.Message
		if err := proto.Unmarshal(result.Entry.Value, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (m *MessageDs) GetMessageHead(ctx context.Context, groupID GroupID) (string, error) {
	headKey := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "head"})
	head, err := m.Get(ctx, headKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", err
	}

	return string(head), nil
}
func (m *MessageDs) GetMessageTail(ctx context.Context, groupID GroupID) (string, error) {
	tailKey := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "tail"})
	tail, err := m.Get(ctx, tailKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", err
	}

	return string(tail), nil
}

func (m *MessageDs) GetMessageLength(context.Context, GroupID) (int32, error) {
	return 0, nil
}
