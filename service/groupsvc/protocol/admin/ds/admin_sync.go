package ds

import (
	"context"
	"errors"
	"strings"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/admin/pb"
	"google.golang.org/protobuf/proto"
)

func (a *AdminDs) GetRangeMessages(groupID string, startID string, endID string) ([]*pb.Log, error) {
	results, err := a.Query(context.Background(), query.Query{
		Prefix:  "/dchat/group/" + groupID + "/admin/logs/",
		Filters: []query.Filter{NewIDRangeFilter(startID, endID)},
		Orders:  []query.Order{query.OrderByKey{}},
	})
	if err != nil {
		return nil, err
	}

	var msgs []*pb.Log

	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		var msg pb.Log
		if err := proto.Unmarshal(result.Entry.Value, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (a *AdminDs) GetRangeIDs(groupID string, startID string, endID string) ([]string, error) {

	results, err := a.Query(context.Background(), query.Query{
		Prefix:   "/dchat/group/" + groupID + "/admin/logs/",
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

func (a *AdminDs) GetMessagesByIDs(groupID string, msgIDs []string) ([]*pb.Log, error) {

	ctx := context.Background()
	prefix := "/dchat/group/" + groupID + "/admin/logs/"

	var msgs []*pb.Log

	for _, msgID := range msgIDs {
		val, err := a.Get(ctx, ds.NewKey(prefix+msgID))
		if err != nil {
			return nil, err
		}

		var msg pb.Log
		if err := proto.Unmarshal(val, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (a *AdminDs) GetMessageHead(ctx context.Context, groupID GroupID) (string, error) {
	headKey := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "admin", "head"})
	head, err := a.Get(ctx, headKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", err
	}

	return string(head), nil
}
func (a *AdminDs) GetMessageTail(ctx context.Context, groupID GroupID) (string, error) {
	tailKey := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "admin", "tail"})
	tail, err := a.Get(ctx, tailKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", err
	}

	return string(tail), nil
}

func (a *AdminDs) GetMessageLength(context.Context, GroupID) (int32, error) {
	return 0, nil
}
