package datastore

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	admpb "github.com/jianbo-zh/dchat/service/group/protocol/admin/pb"
	msgpb "github.com/jianbo-zh/dchat/service/group/protocol/message/pb"
	netpb "github.com/jianbo-zh/dchat/service/group/protocol/network/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-varint"
	"google.golang.org/protobuf/proto"
)

type GroupID string

// group 相关存储操作及接口

type GroupIface interface {
	ds.Batching

	GetLamportTime(context.Context, GroupID) (uint64, error)
	SetLamportTime(context.Context, GroupID, uint64) error
	TickLamportTime(context.Context, GroupID) (uint64, error)

	LogAdminOperation(context.Context, peer.ID, GroupID, *admpb.AdminLog) error
	ListGroups(context.Context) ([]Group, error)

	GroupName(ctx context.Context, groupID GroupID) (string, error)
	GroupRemark(ctx context.Context, groupID GroupID) (string, error)
	GroupNotice(ctx context.Context, groupID GroupID) (string, error)
	SetGroupRemark(ctx context.Context, groupID GroupID, remark string) error
	GroupMemberLogs(ctx context.Context, groupID GroupID) ([]*admpb.AdminLog, error)

	SaveGroupMessage(ctx context.Context, groupID GroupID, msg *msgpb.GroupMsg) error
	GetGroupMessage(ctx context.Context, groupID GroupID, msgID string) (*msgpb.GroupMsg, error)

	CachePeerConnect(ctx context.Context, groupID GroupID, pbmsg *netpb.GroupConnect) error
	GetPeerConnect(ctx context.Context, groupID GroupID, peerID1 peer.ID, peerID2 peer.ID) (*netpb.GroupConnect, error)
	GetGroupConnects(ctx context.Context, groupID GroupID) ([]*netpb.GroupConnect, error)

	GetMessageLamportTime(context.Context, GroupID) (uint64, error)
	SetMessageLamportTime(context.Context, GroupID, uint64) error
	TickMessageLamportTime(context.Context, GroupID) (uint64, error)

	GetMessage(context.Context, GroupID, string) (*msgpb.GroupMsg, error)
	PutMessage(context.Context, GroupID, *msgpb.GroupMsg) error
	ListMessages(context.Context, GroupID) ([]*msgpb.GroupMsg, error)

	GetMessageHeadID(context.Context, GroupID) (string, error)
	GetMessageTailID(context.Context, GroupID) (string, error)
	GetMessageLength(context.Context, GroupID) (int32, error)

	GetNetworkLamportTime(context.Context, GroupID) (uint64, error)
	SetNetworkLamportTime(context.Context, GroupID, uint64) error
	TickNetworkLamportTime(context.Context, GroupID) (uint64, error)
}

var _ GroupIface = (*GroupDataStore)(nil)

type GroupDataStore struct {
	ds.Batching

	lamportMutex        sync.Mutex
	messageLamportMutex sync.Mutex
	networkLamportMutex sync.Mutex
}

func GroupWrap(d ds.Batching) *GroupDataStore {
	return &GroupDataStore{Batching: d}
}

func (gds *GroupDataStore) GetLamportTime(ctx context.Context, groupID GroupID) (uint64, error) {
	gds.lamportMutex.Lock()
	defer gds.lamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "admin", "lamportime"})

	tbs, err := gds.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return varint.ReadUvarint(bytes.NewReader(tbs))
}

func (gds *GroupDataStore) SetLamportTime(ctx context.Context, groupID GroupID, lamptime uint64) error {
	gds.lamportMutex.Lock()
	defer gds.lamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "lamportime"})

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime)

	return gds.Put(ctx, key, buff[:len])
}

func (gds *GroupDataStore) TickLamportTime(ctx context.Context, groupID GroupID) (uint64, error) {
	gds.lamportMutex.Lock()
	defer gds.lamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "lamportime"})

	lamptime := uint64(0)

	if tbs, err := gds.Get(ctx, key); err != nil {

		if !errors.Is(err, ds.ErrNotFound) {
			return 0, err
		}

	} else if lamptime, err = varint.ReadUvarint(bytes.NewReader(tbs)); err != nil {
		return 0, err
	}

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime+1)

	if err := gds.Put(ctx, key, buff[:len]); err != nil {
		return 0, err
	}

	return lamptime + 1, nil
}

func (gds *GroupDataStore) LogAdminOperation(ctx context.Context, peerID peer.ID, groupID GroupID, msg *admpb.AdminLog) error {

	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	batch, err := gds.Batch(ctx)
	if err != nil {
		return err
	}

	msgPrefix := []string{"dchat", "group", string(groupID)}

	msgKey := ds.KeyWithNamespaces(append(msgPrefix, "admin", "logs", msg.Id))
	if err = batch.Put(ctx, msgKey, bs); err != nil {
		return err
	}

	headKey := ds.KeyWithNamespaces(append(msgPrefix, "admin", "head"))
	head, err := gds.Get(ctx, headKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}

	if len(head) == 0 {
		if err = batch.Put(ctx, headKey, []byte(msg.Id)); err != nil {
			return err
		}
	}

	tailKey := ds.KeyWithNamespaces(append(msgPrefix, "admin", "tail"))
	if err = batch.Put(ctx, tailKey, []byte(msg.Id)); err != nil {
		return err
	}

	switch msg.Type {
	case admpb.AdminLog_CREATE:
		createKey := ds.KeyWithNamespaces(append(msgPrefix, "creator"))
		if err = batch.Put(ctx, createKey, []byte(msg.PeerId)); err != nil {
			return err
		}
		stateKey := ds.KeyWithNamespaces(append(msgPrefix, "state"))
		if err = batch.Put(ctx, stateKey, []byte("create")); err != nil {
			return err
		}
	case admpb.AdminLog_DISBAND:
		stateKey := ds.KeyWithNamespaces(append(msgPrefix, "state"))
		if err = batch.Put(ctx, stateKey, []byte("disband")); err != nil {
			return err
		}
	case admpb.AdminLog_NAME:
		nameKey := ds.KeyWithNamespaces(append(msgPrefix, "name"))
		if err = batch.Put(ctx, nameKey, msg.Payload); err != nil {
			return err
		}
	case admpb.AdminLog_NOTICE:
		noticeKey := ds.KeyWithNamespaces(append(msgPrefix, "notice"))
		if err = batch.Put(ctx, noticeKey, msg.Payload); err != nil {
			return err
		}
	case admpb.AdminLog_MEMBER:
		switch msg.Operate {
		case admpb.AdminLog_REMOVE:
			if peerID.String() == msg.MemberId {
				// 自己被移除了，则更新状态
				stateKey := ds.KeyWithNamespaces(append(msgPrefix, "state"))
				if err = batch.Put(ctx, stateKey, []byte("kicked")); err != nil {
					return err
				}
			}

		}
	}

	return batch.Commit(ctx)
}

func (gds *GroupDataStore) ListGroups(ctx context.Context) ([]Group, error) {

	results, err := gds.Query(ctx, query.Query{
		Prefix:  "/dchat/group/",
		Orders:  []query.Order{query.OrderByKeyDescending{}},
		Filters: []query.Filter{GroupFilter{}},
	})
	if err != nil {
		return nil, err
	}

	var groups []Group
	for result := range results.Next() {
		if result.Error != nil {
			return nil, err
		}

		keys := strings.Split(strings.Trim(result.Entry.Key, "/"), "/")

		groups = append(groups, Group{
			ID:   keys[2],
			Name: string(result.Entry.Value),
		})
	}

	return groups, nil
}

func (gds *GroupDataStore) GroupName(ctx context.Context, groupID GroupID) (string, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "name"})

	tbs, err := gds.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			// todo: 从日志里面读取
			return "", nil
		}
		return "", err
	}

	return string(tbs), nil
}

func (gds *GroupDataStore) GroupRemark(ctx context.Context, groupID GroupID) (string, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "remark"})

	tbs, err := gds.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			// todo: 从日志里面读取
			return "", nil
		}
		return "", err
	}

	return string(tbs), nil
}

func (gds *GroupDataStore) GroupNotice(ctx context.Context, groupID GroupID) (string, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "notice"})

	tbs, err := gds.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			// todo: 从日志里面读取
			return "", nil
		}
		return "", err
	}

	return string(tbs), nil
}

func (gds *GroupDataStore) SetGroupRemark(ctx context.Context, groupID GroupID, remark string) error {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "remark"})
	return gds.Put(ctx, key, []byte(remark))
}

func (gds *GroupDataStore) GroupMemberLogs(ctx context.Context, groupID GroupID) ([]*admpb.AdminLog, error) {

	results, err := gds.Query(ctx, query.Query{
		Prefix:  "/dchat/group/" + string(groupID),
		Orders:  []query.Order{GroupOrderByKeyDescending{}},
		Filters: []query.Filter{GroupMemberFilter{}},
	})
	if err != nil {
		return nil, err
	}

	var memberLogs []*admpb.AdminLog

	for result := range results.Next() {
		if result.Error != nil {
			return nil, err
		}

		var pbmsg admpb.AdminLog
		if err := proto.Unmarshal(result.Entry.Value, &pbmsg); err != nil {
			return nil, err
		}

		memberLogs = append(memberLogs, &pbmsg)
	}

	return memberLogs, nil
}

func (gds *GroupDataStore) SaveGroupMessage(ctx context.Context, groupID GroupID, msg *msgpb.GroupMsg) error {
	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	batch, err := gds.Batch(ctx)
	if err != nil {
		return err
	}

	msgPrefix := []string{"dchat", "group", string(groupID)}

	msgKey := ds.KeyWithNamespaces(append(msgPrefix, "message", "logs", msg.Id))
	if err = batch.Put(ctx, msgKey, bs); err != nil {
		return err
	}

	headKey := ds.KeyWithNamespaces(append(msgPrefix, "message", "head"))
	head, err := gds.Get(ctx, headKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}

	if len(head) == 0 {
		if err = batch.Put(ctx, headKey, []byte(msg.Id)); err != nil {
			return err
		}
	}

	tailKey := ds.KeyWithNamespaces(append(msgPrefix, "message", "tail"))
	if err = batch.Put(ctx, tailKey, []byte(msg.Id)); err != nil {
		return err
	}

	return batch.Commit(ctx)
}

func (gds *GroupDataStore) GetGroupMessage(ctx context.Context, groupID GroupID, msgID string) (*msgpb.GroupMsg, error) {

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "logs", msgID})
	bs, err := gds.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var msg msgpb.GroupMsg
	if err = proto.Unmarshal(bs, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}
