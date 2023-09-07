package ds

import (
	"context"
	"errors"
	"strings"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/message/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

func (m *MessageDS) GetMessageHead(ctx context.Context, peerID peer.ID) (string, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "peer", peerID.String(), "message", "head"})
	val, err := m.Get(ctx, key)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", err
	}

	return string(val), nil
}

func (m *MessageDS) GetMessageTail(ctx context.Context, peerID peer.ID) (string, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "peer", peerID.String(), "message", "tail"})
	val, err := m.Get(ctx, key)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", err
	}

	return string(val), nil
}

func (m *MessageDS) GetMessageLength(ctx context.Context, peerID peer.ID) (int32, error) {
	return 0, nil
}

func (m *MessageDS) GetRangeIDs(peerID peer.ID, startID string, endID string) ([]string, error) {

	results, err := m.Query(context.Background(), query.Query{
		Prefix:   "/dchat/peer/" + peerID.String() + "/message/logs/",
		Filters:  []query.Filter{NewIDRangeFilter(startID, endID)},
		Orders:   []query.Order{query.OrderByKey{}},
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	var msgIDs []string
	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		keys := strings.Split(result.Entry.Key, "/")

		msgIDs = append(msgIDs, keys[len(keys)-1])
	}

	return msgIDs, nil
}

func (m *MessageDS) GetRangeMessages(peerID peer.ID, startID string, endID string) ([]*pb.Message, error) {
	results, err := m.Query(context.Background(), query.Query{
		Prefix:  "/dchat/peer/" + peerID.String() + "/message/logs/",
		Filters: []query.Filter{NewIDRangeFilter(startID, endID)},
		Orders:  []query.Order{query.OrderByKey{}},
	})
	if err != nil {
		return nil, err
	}

	var msgs []*pb.Message

	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		var msg pb.Message
		if err := proto.Unmarshal(result.Entry.Value, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (m *MessageDS) GetMessagesByIDs(peerID peer.ID, msgIDs []string) ([]*pb.Message, error) {

	ctx := context.Background()
	prefix := "/dchat/peer/" + peerID.String() + "/message/logs/"

	var msgs []*pb.Message

	for _, msgID := range msgIDs {
		val, err := m.Get(ctx, ds.NewKey(prefix+msgID))
		if err != nil {
			return nil, err
		}

		var msg pb.Message
		if err := proto.Unmarshal(val, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}
