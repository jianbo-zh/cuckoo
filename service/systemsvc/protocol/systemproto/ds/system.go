package ds

import (
	"context"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/internal/datastore"
	"github.com/jianbo-zh/dchat/service/systemsvc/protocol/systemproto/pb"
	"google.golang.org/protobuf/proto"
)

var _ SystemIface = (*SystemDS)(nil)

var systemDsKey = &datastore.SystemDsKey{}

type SystemDS struct {
	ipfsds.Batching
}

func Wrap(b ipfsds.Batching) *SystemDS {
	return &SystemDS{Batching: b}
}

func (a *SystemDS) AddSystemMessage(ctx context.Context, msg *pb.SystemMsg) error {
	value, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return a.Put(ctx, systemDsKey.MsgLogKey(msg.Id), value)
}

func (a *SystemDS) GetSystemMessage(ctx context.Context, msgID string) (*pb.SystemMsg, error) {
	value, err := a.Get(ctx, systemDsKey.MsgLogKey(msgID))
	if err != nil {
		return nil, err
	}

	var msg pb.SystemMsg
	if err := proto.Unmarshal(value, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (a *SystemDS) UpdateSystemMessageState(ctx context.Context, msgID string, state string) error {
	value, err := a.Get(ctx, systemDsKey.MsgLogKey(msgID))
	if err != nil {
		return err
	}

	var msg pb.SystemMsg
	if err := proto.Unmarshal(value, &msg); err != nil {
		return err
	}

	msg.SystemState = state
	msg.UpdateTime = time.Now().Unix()

	value2, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	return a.Put(ctx, systemDsKey.MsgLogKey(msgID), value2)
}

func (a *SystemDS) GetSystemMessageList(ctx context.Context, offset int, limit int) ([]*pb.SystemMsg, error) {
	results, err := a.Query(context.Background(), query.Query{
		Prefix: systemDsKey.MsgLogPrefix(),
		Orders: []query.Order{query.OrderByKey{}},
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	var msgs []*pb.SystemMsg

	for entry := range results.Next() {
		if entry.Error != nil {
			return nil, entry.Error
		}

		var msg pb.SystemMsg
		if err := proto.Unmarshal(entry.Value, &msg); err != nil {
			return nil, err
		}

		msgs = append(msgs, &msg)
	}

	return msgs, nil
}
