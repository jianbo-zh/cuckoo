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
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/adminproto/pb"
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

func (a *AdminDs) SaveLog(ctx context.Context, hostID peer.ID, groupID string, msg *pb.Log) error {

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

	switch msg.LogType {
	case pb.Log_CREATE:
		if err = batch.Put(ctx, adminDsKey.CreatorKey(groupID), []byte(msg.PeerId)); err != nil {
			return err
		}
		if err = batch.Put(ctx, adminDsKey.StateKey(groupID), []byte(types.GroupStateNormal)); err != nil {
			return err
		}
		if err = batch.Put(ctx, adminDsKey.SessionKey(groupID), []byte(strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
			return err
		}
	case pb.Log_DISBAND:
		if err = batch.Put(ctx, adminDsKey.StateKey(groupID), []byte(types.GroupStateDisband)); err != nil {
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
		switch msg.MemberOperate {
		case pb.Log_REMOVE, pb.Log_EXIT:
			if hostID == peer.ID(msg.Member.Id) {
				if err = batch.Put(ctx, adminDsKey.StateKey(groupID), []byte(types.GroupStateExit)); err != nil {
					return err
				}
			}
		case pb.Log_AGREE, pb.Log_APPLY:
			if hostID == peer.ID(msg.Member.Id) {
				// todo: 操作和状态有问题？
				if err = batch.Put(ctx, adminDsKey.StateKey(groupID), []byte(types.GroupStateNormal)); err != nil {
					return err
				}
			}
		}
	}

	return batch.Commit(ctx)
}

func (a *AdminDs) GetState(ctx context.Context, groupID string) (string, error) {
	value, err := a.Get(ctx, adminDsKey.StateKey(groupID))
	if err != nil {
		return "", fmt.Errorf("ds get state error: %w", err)
	}

	return string(value), nil
}

func (a *AdminDs) JoinGroupSaveLog(ctx context.Context, hostID peer.ID, groupID string, msg *pb.Log) error {

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

func (a *AdminDs) JoinGroup(ctx context.Context, group *types.Group) error {
	batch, err := a.Batch(ctx)
	if err != nil {
		return err
	}

	if err = batch.Put(ctx, adminDsKey.NameKey(group.ID), []byte(group.Name)); err != nil {
		return fmt.Errorf("batch.Put error: %w", err)
	}

	if err = batch.Put(ctx, adminDsKey.AvatarKey(group.ID), []byte(group.Avatar)); err != nil {
		return fmt.Errorf("batch.Put error: %w", err)
	}

	if err = batch.Put(ctx, adminDsKey.StateKey(group.ID), []byte(types.GroupStateNormal)); err != nil {
		return fmt.Errorf("batch.Put error: %w", err)
	}

	if err := batch.Put(ctx, adminDsKey.SessionKey(group.ID), []byte(strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
		return fmt.Errorf("batch.Put error: %w", err)
	}

	if err = batch.Commit(ctx); err != nil {
		return fmt.Errorf("batch.Commit error: %w", err)
	}

	return nil
}

func (a *AdminDs) DeleteGroup(ctx context.Context, groupID string) error {

	result, err := a.Query(ctx, query.Query{
		Prefix: adminDsKey.AdminPrefix(groupID),
	})
	if err != nil {
		return fmt.Errorf("ds query error: %w", err)
	}

	for entry := range result.Next() {
		if entry.Error != nil {
			return fmt.Errorf("entry.Error error: %w", entry.Error)
		}

		if err = a.Delete(ctx, ds.NewKey(entry.Key)); err != nil {
			return fmt.Errorf("ds delete key error: %w", err)
		}
	}

	if err = a.Delete(ctx, adminDsKey.SessionKey(groupID)); err != nil {
		return fmt.Errorf("ds delete session key error: %w", err)
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

	return &Group{
		ID:     groupID,
		Name:   string(namebs),
		Avatar: string(avatarbs),
	}, nil
}

func (a *AdminDs) GetGroups(ctx context.Context) ([]Group, error) {

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

func (a *AdminDs) GroupName(ctx context.Context, groupID string) (string, error) {

	namebs, err := a.Get(ctx, adminDsKey.NameKey(groupID))
	if err != nil {
		return "", fmt.Errorf("a.Get group name error: %w", err)
	}

	return string(namebs), nil
}

func (a *AdminDs) GroupLocalName(ctx context.Context, groupID string) (string, error) {

	namebs, err := a.Get(ctx, adminDsKey.LocalNameKey(groupID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", fmt.Errorf("a.Get group local name error: %w", err)
	}

	return string(namebs), nil
}

func (a *AdminDs) GroupAvatar(ctx context.Context, groupID string) (string, error) {

	avatarbs, err := a.Get(ctx, adminDsKey.LocalAvatarKey(groupID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", fmt.Errorf("a.Get group local avatar error: %w", err)
	}

	return string(avatarbs), nil
}

func (a *AdminDs) GroupLocalAvatar(ctx context.Context, groupID string) (string, error) {

	avatarbs, err := a.Get(ctx, adminDsKey.LocalAvatarKey(groupID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", fmt.Errorf("ds group local avatar error: %w", err)
	}

	return string(avatarbs), nil
}

func (a *AdminDs) GroupNotice(ctx context.Context, groupID string) (string, error) {

	notice, err := a.Get(ctx, adminDsKey.NoticeKey(groupID))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return "", fmt.Errorf("ds group notice error: %w", err)
	}

	return string(notice), nil
}

func (a *AdminDs) SetGroupLocalName(ctx context.Context, groupID string, name string) error {
	return a.Put(ctx, adminDsKey.LocalNameKey(groupID), []byte(name))
}

func (a *AdminDs) SetGroupLocalAvatar(ctx context.Context, groupID string, avatar string) error {
	return a.Put(ctx, adminDsKey.LocalAvatarKey(groupID), []byte(avatar))
}

func (a *AdminDs) GroupMemberLogs(ctx context.Context, groupID string) ([]*pb.Log, error) {

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
