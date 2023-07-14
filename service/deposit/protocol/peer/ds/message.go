package ds

import (
	"context"
	"errors"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/service/deposit/protocol/peer/pb"
	"google.golang.org/protobuf/proto"
)

var NSPrefix = []string{"dchat", "deposit", "peer"}

var _ DepositMessageIface = (*DepositPeerDataStore)(nil)

type DepositPeerDataStore struct {
	ipfsds.Batching
}

func DepositPeerWrap(d ipfsds.Batching) *DepositPeerDataStore {
	return &DepositPeerDataStore{Batching: d}
}

func (pds *DepositPeerDataStore) SaveDepositMessage(msg *pb.DepositMessage) error {
	prefix := "/dchat/deposit/peer/" + msg.ToPeerId + "/message/logs/"

	msg.Id = fmt.Sprintf("%d_%s", msg.DepositTime, msg.FromPeerId)

	key := ipfsds.NewKey(prefix + msg.Id)

	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return pds.Put(context.Background(), key, bs)
}

func (pds *DepositPeerDataStore) GetDepositMessages(peerID string, offset int, limit int, startTime int64, lastID string) (msgs []*pb.DepositMessage, err error) {
	prefix := "/dchat/deposit/peer/" + peerID + "/message/logs/"
	results, err := pds.Query(context.Background(), query.Query{
		Prefix:  prefix,
		Filters: []query.Filter{TimePrefixFilter{StartTime: startTime, Sep: "_"}},
		Orders:  []query.Order{query.OrderByKey{}},
		Offset:  offset,
		Limit:   limit,
	})

	if err != nil {
		return nil, err
	}

	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		var msg pb.DepositMessage
		if err := proto.Unmarshal(result.Entry.Value, &msg); err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (pds *DepositPeerDataStore) SetLastAckID(peerID string, ackID string) error {
	key := ipfsds.NewKey("/dchat/deposit/peer/" + peerID + "/ackid")
	return pds.Put(context.Background(), key, []byte(ackID))
}

func (pds *DepositPeerDataStore) GetLastAckID(peerID string) (string, error) {
	key := ipfsds.NewKey("/dchat/deposit/peer/" + peerID + "/ackid")
	ackbs, err := pds.Get(context.Background(), key)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return "", err
	}

	return string(ackbs), nil
}
