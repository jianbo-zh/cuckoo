package groupadminds

import (
	"context"
	"strings"

	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/service/groupsvc/datastore/filter"
	pb "github.com/jianbo-zh/dchat/service/groupsvc/protobuf/pb/grouppb"
	"google.golang.org/protobuf/proto"
)

func (a *AdminDs) GetRangeLogs(groupID string, startID string, endID string) ([]*pb.GroupLog, error) {
	results, err := a.Query(context.Background(), query.Query{
		Prefix:  adminDsKey.AdminLogPrefix(groupID),
		Filters: []query.Filter{filter.NewIDRangeFilter(startID, endID)},
		Orders:  []query.Order{query.OrderByKey{}},
	})
	if err != nil {
		return nil, err
	}

	var msgs []*pb.GroupLog

	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		var msg pb.GroupLog
		if err := proto.Unmarshal(result.Entry.Value, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (a *AdminDs) GetRangeLogIDs(groupID string, startID string, endID string) ([]string, error) {

	results, err := a.Query(context.Background(), query.Query{
		Prefix:   adminDsKey.AdminLogPrefix(groupID),
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

func (a *AdminDs) GetLogsByIDs(groupID string, msgIDs []string) ([]*pb.GroupLog, error) {

	ctx := context.Background()

	var msgs []*pb.GroupLog

	for _, msgID := range msgIDs {
		val, err := a.Get(ctx, adminDsKey.AdminLogKey(groupID, msgID))
		if err != nil {
			return nil, err
		}

		var msg pb.GroupLog
		if err := proto.Unmarshal(val, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}
