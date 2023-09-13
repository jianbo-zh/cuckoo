package ds

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/admin/pb"
	"github.com/libp2p/go-libp2p/core/peer"
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

func (a *AdminDs) SaveLog(ctx context.Context, hostID peer.ID, groupID GroupID, msg *pb.Log) error {

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
	headID, err := a.Get(ctx, headKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}

	if len(headID) == 0 || msg.Id < string(headID) {
		if err = batch.Put(ctx, headKey, []byte(msg.Id)); err != nil {
			return err
		}
	}

	tailKey := ds.KeyWithNamespaces(append(msgPrefix, "admin", "tail"))
	tailID, err := a.Get(ctx, tailKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	if len(tailID) == 0 || msg.Id > string(tailID) {
		if err = batch.Put(ctx, tailKey, []byte(msg.Id)); err != nil {
			return err
		}
	}

	switch msg.Type {
	case pb.Log_CREATE:
		createKey := ds.KeyWithNamespaces(append(msgPrefix, "creator"))
		if err = batch.Put(ctx, createKey, []byte(msg.PeerId)); err != nil {
			return err
		}
		stateKey := ds.KeyWithNamespaces(append(msgPrefix, "state"))
		if err = batch.Put(ctx, stateKey, []byte("normal")); err != nil {
			return err
		}
		groupListKey := ds.KeyWithNamespaces([]string{"dchat", "group", "list", groupID})
		if err = batch.Put(ctx, groupListKey, []byte(strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
			return err
		}
	case pb.Log_DISBAND:
		stateKey := ds.KeyWithNamespaces(append(msgPrefix, "state"))
		if err = batch.Put(ctx, stateKey, []byte("disband")); err != nil {
			return err
		}
	case pb.Log_NAME:
		nameKey := ds.KeyWithNamespaces(append(msgPrefix, "name"))
		if err = batch.Put(ctx, nameKey, msg.Payload); err != nil {
			return err
		}
	case pb.Log_AVATAR:
		avatarKey := ds.KeyWithNamespaces(append(msgPrefix, "avatar"))
		if err = batch.Put(ctx, avatarKey, msg.Payload); err != nil {
			return err
		}
	case pb.Log_NOTICE:
		noticeKey := ds.KeyWithNamespaces(append(msgPrefix, "notice"))
		if err = batch.Put(ctx, noticeKey, msg.Payload); err != nil {
			return err
		}
	case pb.Log_MEMBER:
		switch msg.Operate {
		case pb.Log_REMOVE:
			if hostID.String() == msg.MemberId {
				// 自己被移除了，则更新状态
				stateKey := ds.KeyWithNamespaces(append(msgPrefix, "state"))
				if err = batch.Put(ctx, stateKey, []byte("kicked")); err != nil {
					return err
				}
			}
		case pb.Log_AGREE, pb.Log_APPLY:
			if hostID.String() == msg.MemberId {
				// 新加入
				stateKey := ds.KeyWithNamespaces(append(msgPrefix, "state"))
				if err = batch.Put(ctx, stateKey, []byte("normal")); err != nil {
					return err
				}
			}
		}
	}

	return batch.Commit(ctx)
}

func (a *AdminDs) JoinGroupSaveLog(ctx context.Context, hostID peer.ID, groupID GroupID, msg *pb.Log) error {

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
	headID, err := a.Get(ctx, headKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}

	if len(headID) == 0 || msg.Id < string(headID) {
		if err = batch.Put(ctx, headKey, []byte(msg.Id)); err != nil {
			return err
		}
	}

	tailKey := ds.KeyWithNamespaces(append(msgPrefix, "admin", "tail"))
	tailID, err := a.Get(ctx, tailKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	if len(tailID) == 0 || msg.Id > string(tailID) {
		if err = batch.Put(ctx, tailKey, []byte(msg.Id)); err != nil {
			return err
		}
	}

	return batch.Commit(ctx)
}

func (a *AdminDs) JoinGroup(ctx context.Context, groupID string, name string, avatar string) error {
	batch, err := a.Batch(ctx)
	if err != nil {
		return err
	}

	msgPrefix := []string{"dchat", "group", string(groupID)}
	nameKey := ds.KeyWithNamespaces(append(msgPrefix, "name"))
	if err = batch.Put(ctx, nameKey, []byte(name)); err != nil {
		return fmt.Errorf("batch.Put error: %w", err)
	}

	avatarKey := ds.KeyWithNamespaces(append(msgPrefix, "avatar"))
	if err = batch.Put(ctx, avatarKey, []byte(avatar)); err != nil {
		return fmt.Errorf("batch.Put error: %w", err)
	}

	stateKey := ds.KeyWithNamespaces(append(msgPrefix, "state"))
	if err = batch.Put(ctx, stateKey, []byte("normal")); err != nil {
		return fmt.Errorf("batch.Put error: %w", err)
	}

	groupListKey := ds.KeyWithNamespaces([]string{"dchat", "group", "list", groupID})
	if err := batch.Put(ctx, groupListKey, []byte(strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
		return fmt.Errorf("batch.Put error: %w", err)
	}

	if err = batch.Commit(ctx); err != nil {
		return fmt.Errorf("batch.Commit error: %w", err)
	}

	return nil
}

func (a *AdminDs) DeleteGroup(ctx context.Context, groupID string) error {

	result, err := a.Query(ctx, query.Query{
		Filters: []query.Filter{query.FilterKeyPrefix{
			Prefix: "/dchat/group/" + groupID + "/",
		}},
	})
	if err != nil {
		return fmt.Errorf("a.query error: %w", err)
	}

	fmt.Println("a.query result")

	for entry := range result.Next() {
		if entry.Error != nil {
			return fmt.Errorf("entry.Error error: %w", entry.Error)
		}

		fmt.Println("group key ", entry.Key)

		if err = a.Delete(ctx, ds.NewKey(entry.Key)); err != nil {
			return fmt.Errorf("a.Delete error: %w", err)
		}
	}

	if err = a.Delete(ctx, ds.NewKey("/dchat/group/list/"+groupID)); err != nil {
		return fmt.Errorf("a.Delete error: %w", err)
	}

	return nil
}

func (a *AdminDs) GetGroupIDs(ctx context.Context) ([]string, error) {

	msgPrefix := "/dchat/group/list/"
	results, err := a.Query(ctx, query.Query{
		Prefix: msgPrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var groupIDs []string
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("result.Error: %w", err)
		}

		groupIDs = append(groupIDs, strings.TrimPrefix(result.Key, msgPrefix))
	}

	return groupIDs, nil
}

func (a *AdminDs) GetGroup(ctx context.Context, groupID string) (*Group, error) {

	msgPrefix := []string{"dchat", "group", string(groupID)}
	nameKey := ds.KeyWithNamespaces(append(msgPrefix, "name"))
	nameBs, err := a.Get(ctx, nameKey)
	if err != nil {
		return nil, fmt.Errorf("a.Get error: %w", err)
	}

	avatarKey := ds.KeyWithNamespaces(append(msgPrefix, "avatar"))
	avatarBs, err := a.Get(ctx, avatarKey)
	if err != nil {
		return nil, fmt.Errorf("a.Get error: %w", err)
	}

	return &Group{
		ID:     groupID,
		Name:   string(nameBs),
		Avatar: string(avatarBs),
	}, nil
}

func (a *AdminDs) ListGroups(ctx context.Context) ([]Group, error) {

	results, err := a.Query(ctx, query.Query{
		Prefix: "/dchat/group/list",
		Orders: []query.Order{GroupOrderByValueTimeDescending{}},
	})
	if err != nil {
		return nil, err
	}

	var groups []Group
	for result := range results.Next() {
		if result.Error != nil {
			return nil, err
		}

		fields := strings.Split(strings.Trim(result.Entry.Key, "/"), "/")
		groupID := fields[len(fields)-1]

		groupPrefix := []string{"dchat", "group", string(groupID)}
		namebs, err := a.Get(ctx, ds.KeyWithNamespaces(append(groupPrefix, "localname")))
		if err != nil && !errors.Is(err, ds.ErrNotFound) {
			return nil, fmt.Errorf("a.Get group local name error: %w", err)
		}
		if len(namebs) == 0 {
			namebs, err = a.Get(ctx, ds.KeyWithNamespaces(append(groupPrefix, "name")))
			if err != nil {
				return nil, fmt.Errorf("a.Get group name error: %w", err)
			}
		}

		avatarbs, err := a.Get(ctx, ds.KeyWithNamespaces(append(groupPrefix, "localavatar")))
		if err != nil && !errors.Is(err, ds.ErrNotFound) {
			return nil, fmt.Errorf("a.Get group local avatar error: %w", err)
		}
		if len(avatarbs) == 0 {
			avatarbs, err = a.Get(ctx, ds.KeyWithNamespaces(append(groupPrefix, "avatar")))
			if err != nil {
				return nil, fmt.Errorf("a.Get group avatar error: %w", err)
			}
		}

		groups = append(groups, Group{
			ID:     groupID,
			Name:   string(namebs),
			Avatar: string(avatarbs),
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

func (a *AdminDs) GroupLocalName(ctx context.Context, groupID GroupID) (string, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "localname"})

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

func (a *AdminDs) GroupAvatar(ctx context.Context, groupID GroupID) (string, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "avatar"})

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

func (a *AdminDs) GroupLocalAvatar(ctx context.Context, groupID GroupID) (string, error) {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "localavatar"})

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

func (a *AdminDs) SetGroupLocalName(ctx context.Context, groupID GroupID, name string) error {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "localname"})
	return a.Put(ctx, key, []byte(name))
}

func (a *AdminDs) SetGroupLocalAvatar(ctx context.Context, groupID GroupID, avatar string) error {
	key := ds.KeyWithNamespaces([]string{"dchat", "group", string(groupID), "localavatar"})
	return a.Put(ctx, key, []byte(avatar))
}

func (a *AdminDs) GroupMemberLogs(ctx context.Context, groupID GroupID) ([]*pb.Log, error) {

	results, err := a.Query(ctx, query.Query{
		Prefix:  "/dchat/group/" + string(groupID),
		Orders:  []query.Order{GroupOrderByKeyDescending{}},
		Filters: []query.Filter{GroupMemberFilter{}},
	})
	if err != nil {
		return nil, err
	}

	var memberLogs []*pb.Log

	for result := range results.Next() {
		if result.Error != nil {
			return nil, err
		}

		var pbmsg pb.Log
		if err := proto.Unmarshal(result.Entry.Value, &pbmsg); err != nil {
			return nil, err
		}

		memberLogs = append(memberLogs, &pbmsg)
	}

	return memberLogs, nil
}
