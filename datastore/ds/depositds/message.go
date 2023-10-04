package depositds

import (
	"context"
	"errors"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/datastore/filter"
	"github.com/jianbo-zh/dchat/internal/datastore"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/depositpb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

var _ DepositMessageIface = (*DepositPeerDataStore)(nil)

var depositDsKey = &datastore.DepositDsKey{}

type DepositPeerDataStore struct {
	ipfsds.Batching
}

func DepositPeerWrap(d ipfsds.Batching) *DepositPeerDataStore {
	return &DepositPeerDataStore{Batching: d}
}

func (pds *DepositPeerDataStore) SaveContactMessage(msg *pb.DepositContactMessage) error {

	msg.Id = contactMsgID(peer.ID(msg.FromPeerId))
	key := depositDsKey.PeerMsgLogKey(peer.ID(msg.ToPeerId), msg.Id)

	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return pds.Put(context.Background(), key, bs)
}

func (pds *DepositPeerDataStore) GetContactMessages(peerID peer.ID, startID string, limit int) (msgs []*pb.DepositContactMessage, err error) {
	results, err := pds.Query(context.Background(), query.Query{
		Prefix:  depositDsKey.PeerMsgLogPrefix(peerID),
		Filters: []query.Filter{filter.NewFromKeyFilter(startID)},
		Orders:  []query.Order{query.OrderByKey{}},
		Limit:   limit,
	})

	if err != nil {
		return nil, err
	}

	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		var msg pb.DepositContactMessage
		if err := proto.Unmarshal(result.Entry.Value, &msg); err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (pds *DepositPeerDataStore) SaveGroupMessage(msg *pb.DepositGroupMessage) error {

	msg.Id = groupMsgID(msg.GroupId)
	key := depositDsKey.GroupMsgLogKey(msg.GroupId, msg.Id)

	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return pds.Put(context.Background(), key, bs)
}

func (pds *DepositPeerDataStore) GetGroupMessages(groupID string, startID string, limit int) (msgs []*pb.DepositGroupMessage, err error) {
	results, err := pds.Query(context.Background(), query.Query{
		Prefix:  depositDsKey.GroupMsgLogPrefix(groupID),
		Filters: []query.Filter{filter.NewFromKeyFilter(startID)},
		Orders:  []query.Order{query.OrderByKey{}},
		Limit:   limit,
	})

	if err != nil {
		return nil, err
	}

	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		var msg pb.DepositGroupMessage
		if err := proto.Unmarshal(result.Entry.Value, &msg); err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}

	return msgs, nil
}

func (pds *DepositPeerDataStore) SetContactLastID(peerID peer.ID, depositID string) error {
	return pds.Put(context.Background(), depositDsKey.PeerLastIDKey(peerID), []byte(depositID))
}

func (pds *DepositPeerDataStore) GetContactLastID(peerID peer.ID) (string, error) {
	ackbs, err := pds.Get(context.Background(), depositDsKey.PeerLastIDKey(peerID))
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return "", err
	}

	return string(ackbs), nil
}

func (pds *DepositPeerDataStore) SetGroupLastID(groupID string, depositID string) error {
	return pds.Put(context.Background(), depositDsKey.GroupLastIDKey(groupID), []byte(depositID))
}

func (pds *DepositPeerDataStore) GetGroupLastID(groupID string) (string, error) {
	ackbs, err := pds.Get(context.Background(), depositDsKey.GroupLastIDKey(groupID))
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return "", err
	}

	return string(ackbs), nil
}

func contactMsgID(peerID peer.ID) string {
	return fmt.Sprintf("%019d_%s", time.Now().Unix(), peerID.String())
}

func groupMsgID(groupID string) string {
	return fmt.Sprintf("%019d_%s", time.Now().Unix(), groupID)
}
