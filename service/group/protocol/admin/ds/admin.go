package ds

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/service/group/protocol/admin/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-varint"
	"google.golang.org/protobuf/proto"
)

// group 相关存储操作及接口

var _ AdminIface = (*AdminDs)(nil)

type AdminDs struct {
	ds.Batching

	lamportMutex sync.Mutex
}

func AdminWrap(d ds.Batching) *AdminDs {
	return &AdminDs{Batching: d}
}

func (a *AdminDs) GetLamportTime(ctx context.Context, groupID GroupID) (uint64, error) {
	a.lamportMutex.Lock()
	defer a.lamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "admin", "lamportime"})

	tbs, err := a.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return varint.ReadUvarint(bytes.NewReader(tbs))
}

func (a *AdminDs) SetLamportTime(ctx context.Context, groupID GroupID, lamptime uint64) error {
	a.lamportMutex.Lock()
	defer a.lamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "lamportime"})

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime)

	return a.Put(ctx, key, buff[:len])
}

func (a *AdminDs) TickLamportTime(ctx context.Context, groupID GroupID) (uint64, error) {
	a.lamportMutex.Lock()
	defer a.lamportMutex.Unlock()

	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "message", "lamportime"})

	lamptime := uint64(0)

	if tbs, err := a.Get(ctx, key); err != nil {

		if !errors.Is(err, ds.ErrNotFound) {
			return 0, err
		}

	} else if lamptime, err = varint.ReadUvarint(bytes.NewReader(tbs)); err != nil {
		return 0, err
	}

	buff := make([]byte, varint.MaxLenUvarint63)
	len := varint.PutUvarint(buff, lamptime+1)

	if err := a.Put(ctx, key, buff[:len]); err != nil {
		return 0, err
	}

	return lamptime + 1, nil
}

func (a *AdminDs) LogAdminOperation(ctx context.Context, peerID peer.ID, groupID GroupID, msg *pb.AdminLog) error {

	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	batch, err := a.Batch(ctx)
	if err != nil {
		return err
	}

	msgPrefix := []string{"dchat", "group", string(groupID)}

	msgKey := ds.KeyWithNamespaces(append(msgPrefix, "admin", "logs", msg.Id))
	if err = batch.Put(ctx, msgKey, bs); err != nil {
		return err
	}

	headKey := ds.KeyWithNamespaces(append(msgPrefix, "admin", "head"))
	head, err := a.Get(ctx, headKey)
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
	case pb.AdminLog_CREATE:
		createKey := ds.KeyWithNamespaces(append(msgPrefix, "creator"))
		if err = batch.Put(ctx, createKey, []byte(msg.PeerId)); err != nil {
			return err
		}
		stateKey := ds.KeyWithNamespaces(append(msgPrefix, "state"))
		if err = batch.Put(ctx, stateKey, []byte("create")); err != nil {
			return err
		}
	case pb.AdminLog_DISBAND:
		stateKey := ds.KeyWithNamespaces(append(msgPrefix, "state"))
		if err = batch.Put(ctx, stateKey, []byte("disband")); err != nil {
			return err
		}
	case pb.AdminLog_NAME:
		nameKey := ds.KeyWithNamespaces(append(msgPrefix, "name"))
		if err = batch.Put(ctx, nameKey, msg.Payload); err != nil {
			return err
		}
	case pb.AdminLog_NOTICE:
		noticeKey := ds.KeyWithNamespaces(append(msgPrefix, "notice"))
		if err = batch.Put(ctx, noticeKey, msg.Payload); err != nil {
			return err
		}
	case pb.AdminLog_MEMBER:
		switch msg.Operate {
		case pb.AdminLog_REMOVE:
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

func (a *AdminDs) ListGroups(ctx context.Context) ([]Group, error) {

	results, err := a.Query(ctx, query.Query{
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

func (a *AdminDs) GroupName(ctx context.Context, groupID GroupID) (string, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "name"})

	tbs, err := a.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			// todo: 从日志里面读取
			return "", nil
		}
		return "", err
	}

	return string(tbs), nil
}

func (a *AdminDs) GroupRemark(ctx context.Context, groupID GroupID) (string, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "remark"})

	tbs, err := a.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			// todo: 从日志里面读取
			return "", nil
		}
		return "", err
	}

	return string(tbs), nil
}

func (a *AdminDs) GroupNotice(ctx context.Context, groupID GroupID) (string, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "notice"})

	tbs, err := a.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			// todo: 从日志里面读取
			return "", nil
		}
		return "", err
	}

	return string(tbs), nil
}

func (a *AdminDs) SetGroupRemark(ctx context.Context, groupID GroupID, remark string) error {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "remark"})
	return a.Put(ctx, key, []byte(remark))
}

func (a *AdminDs) GroupMemberLogs(ctx context.Context, groupID GroupID) ([]*pb.AdminLog, error) {

	results, err := a.Query(ctx, query.Query{
		Prefix:  "/dchat/group/" + string(groupID),
		Orders:  []query.Order{GroupOrderByKeyDescending{}},
		Filters: []query.Filter{GroupMemberFilter{}},
	})
	if err != nil {
		return nil, err
	}

	var memberLogs []*pb.AdminLog

	for result := range results.Next() {
		if result.Error != nil {
			return nil, err
		}

		var pbmsg pb.AdminLog
		if err := proto.Unmarshal(result.Entry.Value, &pbmsg); err != nil {
			return nil, err
		}

		memberLogs = append(memberLogs, &pbmsg)
	}

	return memberLogs, nil
}
