package ds

import (
	"context"
	"errors"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/service/deposit/protocol/peer/pb"
	"github.com/libp2p/go-libp2p/core/peer"
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

func (pds *DepositPeerDataStore) SaveDepositMessage(msg *pb.OfflineMessage) error {
	prefix := "/dchat/deposit/peer/" + peer.ID(msg.ToPeerId).String() + "/message/logs/"

	msg.Id = msgID(msg.DepositTime, peer.ID(msg.FromPeerId))

	key := ipfsds.NewKey(prefix + msg.Id)

	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return pds.Put(context.Background(), key, bs)
}

func (pds *DepositPeerDataStore) GetDepositMessages(peerID peer.ID, offset int, limit int, startTime int64, lastID string) (msgs []*pb.OfflineMessage, err error) {
	prefix := "/dchat/deposit/peer/" + peerID.String() + "/message/logs/"
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

		var msg pb.OfflineMessage
		if err := proto.Unmarshal(result.Entry.Value, &msg); err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (pds *DepositPeerDataStore) SetLastAckID(peerID peer.ID, ackID string) error {
	key := ipfsds.NewKey("/dchat/deposit/peer/" + peerID.String() + "/ackid")
	return pds.Put(context.Background(), key, []byte(ackID))
}

func (pds *DepositPeerDataStore) GetLastAckID(peerID peer.ID) (string, error) {
	key := ipfsds.NewKey("/dchat/deposit/peer/" + peerID.String() + "/ackid")
	ackbs, err := pds.Get(context.Background(), key)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return "", err
	}

	return string(ackbs), nil
}

func msgID(timestamp int64, peerID peer.ID) string {
	return fmt.Sprintf("%019d_%s", timestamp, peerID.String())
}
