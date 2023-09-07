package ds

import (
	"context"
	"strings"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/message/pb"
	"google.golang.org/protobuf/proto"
)

func (m *MessageDs) GetRangeMessages(groupID string, startID string, endID string) ([]*pb.Message, error) {
	results, err := m.Query(context.Background(), query.Query{
		Prefix:  "/dchat/group/" + groupID + "/message/logs/",
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

func (m *MessageDs) GetRangeIDs(groupID string, startID string, endID string) ([]string, error) {

	results, err := m.Query(context.Background(), query.Query{
		Prefix:   "/dchat/group/" + groupID + "/message/logs/",
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

func (m *MessageDs) GetMessagesByIDs(groupID string, msgIDs []string) ([]*pb.Message, error) {

	ctx := context.Background()
	prefix := "/dchat/group/" + groupID + "/message/logs/"

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
