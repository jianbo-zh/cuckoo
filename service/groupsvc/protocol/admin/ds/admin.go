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
	"github.com/jianbo-zh/dchat/internal/datastore"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/admin/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

// group 相关存储操作及接口

var _ AdminIface = (*AdminDs)(nil)

var adminDsKey = &datastore.GroupDsKey{}

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

	if err = batch.Put(ctx, adminDsKey.AdminLogKey(groupID, msg.Id), bs); err != nil {
		return err
	}

	headID, err := a.Get(ctx, adminDsKey.AdminLogHeadKey(groupID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}

	if len(headID) == 0 || msg.Id < string(headID) {
		if err = batch.Put(ctx, adminDsKey.AdminLogHeadKey(groupID), []byte(msg.Id)); err != nil {
			return err
		}
	}

	tailID, err := a.Get(ctx, adminDsKey.AdminLogTailKey(groupID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	if len(tailID) == 0 || msg.Id > string(tailID) {
		if err = batch.Put(ctx, adminDsKey.AdminLogTailKey(groupID), []byte(msg.Id)); err != nil {
			return err
		}
	}

	switch msg.Type {
	case pb.Log_CREATE:
		if err = batch.Put(ctx, adminDsKey.CreatorKey(groupID), []byte(msg.PeerId)); err != nil {
			return err
		}
		if err = batch.Put(ctx, adminDsKey.StateKey(groupID), []byte("normal")); err != nil {
			return err
		}
		if err = batch.Put(ctx, adminDsKey.SessionKey(groupID), []byte(strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
			return err
		}
	case pb.Log_DISBAND:
		if err = batch.Put(ctx, adminDsKey.StateKey(groupID), []byte("disband")); err != nil {
			return err
		}
	case pb.Log_NAME:
		if err = batch.Put(ctx, adminDsKey.NameKey(groupID), msg.Payload); err != nil {
			return err
		}
	case pb.Log_AVATAR:
		if err = batch.Put(ctx, adminDsKey.AvatarKey(groupID), msg.Payload); err != nil {
			return err
		}
	case pb.Log_NOTICE:
		if err = batch.Put(ctx, adminDsKey.NoticeKey(groupID), msg.Payload); err != nil {
			return err
		}
	case pb.Log_MEMBER:
		switch msg.Operate {
		case pb.Log_REMOVE:
			if hostID.String() == msg.MemberId {
				// 自己被移除了，则更新状态
				if err = batch.Put(ctx, adminDsKey.StateKey(groupID), []byte("kicked")); err != nil {
					return err
				}
			}
		case pb.Log_AGREE, pb.Log_APPLY:
			if hostID.String() == msg.MemberId {
				// 新加入
				if err = batch.Put(ctx, adminDsKey.StateKey(groupID), []byte("normal")); err != nil {
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

	if err = batch.Put(ctx, adminDsKey.AdminLogKey(groupID, msg.Id), bs); err != nil {
		return err
	}

	headID, err := a.Get(ctx, adminDsKey.AdminLogHeadKey(groupID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}

	if len(headID) == 0 || msg.Id < string(headID) {
		if err = batch.Put(ctx, adminDsKey.AdminLogHeadKey(groupID), []byte(msg.Id)); err != nil {
			return err
		}
	}
	tailID, err := a.Get(ctx, adminDsKey.AdminLogTailKey(groupID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	if len(tailID) == 0 || msg.Id > string(tailID) {
		if err = batch.Put(ctx, adminDsKey.AdminLogTailKey(groupID), []byte(msg.Id)); err != nil {
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

	if err = batch.Put(ctx, adminDsKey.NameKey(groupID), []byte(name)); err != nil {
		return fmt.Errorf("batch.Put error: %w", err)
	}

	if err = batch.Put(ctx, adminDsKey.AvatarKey(groupID), []byte(avatar)); err != nil {
		return fmt.Errorf("batch.Put error: %w", err)
	}

	if err = batch.Put(ctx, adminDsKey.StateKey(groupID), []byte("normal")); err != nil {
		return fmt.Errorf("batch.Put error: %w", err)
	}

	if err := batch.Put(ctx, adminDsKey.SessionKey(groupID), []byte(strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
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
			Prefix: adminDsKey.AdminPrefix(groupID),
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

		if err = a.Delete(ctx, ds.NewKey(entry.Key)); err != nil {
			return fmt.Errorf("a.Delete error: %w", err)
		}
	}

	if err = a.Delete(ctx, adminDsKey.SessionKey(groupID)); err != nil {
		return fmt.Errorf("a.Delete error: %w", err)
	}

	return nil
}

func (a *AdminDs) GetGroupIDs(ctx context.Context) ([]string, error) {

	results, err := a.Query(ctx, query.Query{
		Prefix: adminDsKey.SessionPrefix(),
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var groupIDs []string
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("result.Error: %w", err)
		}

		groupIDs = append(groupIDs, strings.TrimPrefix(result.Key, adminDsKey.SessionPrefix()))
	}

	return groupIDs, nil
}

func (a *AdminDs) GetGroup(ctx context.Context, groupID string) (*Group, error) {

	nameBs, err := a.Get(ctx, adminDsKey.NameKey(groupID))
	if err != nil {
		return nil, fmt.Errorf("a.Get error: %w", err)
	}

	avatarBs, err := a.Get(ctx, adminDsKey.AvatarKey(groupID))
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
		Prefix: adminDsKey.SessionPrefix(),
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

		namebs, err := a.Get(ctx, adminDsKey.LocalNameKey(groupID))
		if err != nil && !errors.Is(err, ds.ErrNotFound) {
			return nil, fmt.Errorf("a.Get group local name error: %w", err)
		}
		if len(namebs) == 0 {
			namebs, err = a.Get(ctx, adminDsKey.NameKey(groupID))
			if err != nil {
				return nil, fmt.Errorf("a.Get group name error: %w", err)
			}
		}

		avatarbs, err := a.Get(ctx, adminDsKey.LocalAvatarKey(groupID))
		if err != nil && !errors.Is(err, ds.ErrNotFound) {
			return nil, fmt.Errorf("a.Get group local avatar error: %w", err)
		}
		if len(avatarbs) == 0 {
			avatarbs, err = a.Get(ctx, adminDsKey.AvatarKey(groupID))
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

	tbs, err := a.Get(ctx, adminDsKey.NameKey(groupID))
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

	tbs, err := a.Get(ctx, adminDsKey.LocalNameKey(groupID))
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

	tbs, err := a.Get(ctx, adminDsKey.AvatarKey(groupID))
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

	tbs, err := a.Get(ctx, adminDsKey.LocalAvatarKey(groupID))
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

	tbs, err := a.Get(ctx, adminDsKey.NoticeKey(groupID))
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
	return a.Put(ctx, adminDsKey.LocalNameKey(groupID), []byte(name))
}

func (a *AdminDs) SetGroupLocalAvatar(ctx context.Context, groupID GroupID, avatar string) error {
	return a.Put(ctx, adminDsKey.LocalAvatarKey(groupID), []byte(avatar))
}

func (a *AdminDs) GroupMemberLogs(ctx context.Context, groupID GroupID) ([]*pb.Log, error) {

	results, err := a.Query(ctx, query.Query{
		Prefix:  adminDsKey.AdminPrefix(groupID),
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
