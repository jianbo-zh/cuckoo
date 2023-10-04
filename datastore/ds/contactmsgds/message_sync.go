package contactmsgds

import (
	"context"
	"errors"
	"strings"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/datastore/filter"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/contactpb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

func (m *MessageDS) GetMessageHead(ctx context.Context, peerID peer.ID) (string, error) {
	val, err := m.Get(ctx, contactDsKey.MsgHeadKey(peerID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", err
	}

	return string(val), nil
}

func (m *MessageDS) GetMessageTail(ctx context.Context, peerID peer.ID) (string, error) {
	val, err := m.Get(ctx, contactDsKey.MsgTailKey(peerID))
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
		Prefix:   contactDsKey.MsgLogPrefix(peerID),
		Filters:  []query.Filter{filter.NewIDRangeFilter(startID, endID)},
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

func (m *MessageDS) GetRangeMessages(peerID peer.ID, startID string, endID string) ([]*pb.ContactMessage, error) {
	results, err := m.Query(context.Background(), query.Query{
		Prefix:  contactDsKey.MsgLogPrefix(peerID),
		Filters: []query.Filter{filter.NewIDRangeFilter(startID, endID)},
		Orders:  []query.Order{query.OrderByKey{}},
	})
	if err != nil {
		return nil, err
	}

	var msgs []*pb.ContactMessage

	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		var msg pb.ContactMessage
		if err := proto.Unmarshal(result.Entry.Value, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (m *MessageDS) GetMessagesByIDs(peerID peer.ID, msgIDs []string) ([]*pb.ContactMessage, error) {

	ctx := context.Background()

	var msgs []*pb.ContactMessage
	for _, msgID := range msgIDs {
		val, err := m.Get(ctx, contactDsKey.MsgLogKey(peerID, msgID))
		if err != nil {
			return nil, err
		}

		var msg pb.ContactMessage
		if err := proto.Unmarshal(val, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}
