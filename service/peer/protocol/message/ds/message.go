package ds

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	msgpb "github.com/jianbo-zh/dchat/service/peer/protocol/message/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

var _ PeerMessageIface = (*MessageDS)(nil)

type MessageDS struct {
	ipfsds.Batching

	lamportMutex sync.Mutex
}

func Wrap(b ipfsds.Batching) *MessageDS {
	return &MessageDS{Batching: b}
}

// HasMessage 消息是否存在
func (m *MessageDS) HasMessage(ctx context.Context, peerID peer.ID, msgID string) (bool, error) {
	key := ipfsds.KeyWithNamespaces([]string{"dchat", "peer", peerID.String(), "message", "logs", msgID})
	return m.Has(ctx, key)
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

	msgPrefix := []string{"dchat", "peer", peerID.String(), "message"}

	msgKey := ipfsds.KeyWithNamespaces(append(msgPrefix, "logs", msg.Id))
	if err = batch.Put(ctx, msgKey, bs); err != nil {
		return err
	}

	headKey := ipfsds.KeyWithNamespaces(append(msgPrefix, "head"))
	head, err := m.Get(ctx, headKey)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return err
	}

	if len(head) == 0 {
		if err = batch.Put(ctx, headKey, []byte(msg.Id)); err != nil {
			return err
		}
	}

	tailKey := ipfsds.KeyWithNamespaces(append(msgPrefix, "tail"))
	if err = batch.Put(ctx, tailKey, []byte(msg.Id)); err != nil {
		return err
	}

	return batch.Commit(ctx)
}

// GetMessages 获取消息列表
func (m *MessageDS) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]*msgpb.Message, error) {
	results, err := m.Query(ctx, query.Query{
		Prefix: fmt.Sprintf("/dchat/peer/%s/message/logs", peerID.String()),
		Orders: []query.Order{query.OrderByKey{}},
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

	return msgs, nil
}

func parseMsgID(msgID string) (lamptime uint64, peerID string, err error) {
	idArr := strings.SplitN(msgID, "_", 2)
	if len(idArr) <= 1 {
		err = fmt.Errorf("msgID <%s> format error", msgID)
		return
	}

	if lamptime, err = strconv.ParseUint(idArr[0], 10, 64); err != nil {
		return
	}

	return lamptime, idArr[1], nil
}
