package groupmsgds

import (
	"context"
	"strings"

	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/service/groupsvc/datastore/filter"
	pb "github.com/jianbo-zh/dchat/service/groupsvc/protobuf/pb/grouppb"
	"google.golang.org/protobuf/proto"
)

func (m *MessageDs) GetRangeMessages(groupID string, startID string, endID string) ([]*pb.GroupMessage, error) {
	results, err := m.Query(context.Background(), query.Query{
		Prefix:  adminDsKey.MsgLogPrefix(groupID),
		Filters: []query.Filter{filter.NewIDRangeFilter(startID, endID)},
		Orders:  []query.Order{query.OrderByKey{}},
	})
	if err != nil {
		return nil, err
	}

	var msgs []*pb.GroupMessage

	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		var msg pb.GroupMessage
		if err := proto.Unmarshal(result.Entry.Value, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (m *MessageDs) GetRangeIDs(groupID string, startID string, endID string) ([]string, error) {

	results, err := m.Query(context.Background(), query.Query{
		Prefix:   adminDsKey.MsgLogPrefix(groupID),
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

func (m *MessageDs) GetMessagesByIDs(groupID string, msgIDs []string) ([]*pb.GroupMessage, error) {

	ctx := context.Background()

	var msgs []*pb.GroupMessage

	for _, msgID := range msgIDs {
		val, err := m.Get(ctx, adminDsKey.MsgLogKey(groupID, msgID))
		if err != nil {
			return nil, err
		}

		var msg pb.GroupMessage
		if err := proto.Unmarshal(val, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}
