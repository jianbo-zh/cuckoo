package groupadminds

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
	pb "github.com/jianbo-zh/dchat/protobuf/pb/grouppb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

// group 相关存储操作及接口

var _ AdminIface = (*AdminDs)(nil)

var adminDsKey = &datastore.GroupDsKey{}

type LogRecorder struct {
	logHead   string
	logTail   string
	logLength int
}

type AdminDs struct {
	ds.Batching

	lamportMutex sync.Mutex

	logMutex   sync.Mutex
	logRecords map[string]*LogRecorder
}

func AdminWrap(d ds.Batching) *AdminDs {
	return &AdminDs{
		Batching:   d,
		logRecords: make(map[string]*LogRecorder, 0),
	}
}

func (a *AdminDs) GetLog(ctx context.Context, groupID string, logID string) (*pb.GroupLog, error) {
	value, err := a.Get(ctx, adminDsKey.AdminLogKey(groupID, logID))
	if err != nil {
		return nil, fmt.Errorf("ds get admin log error: %w", err)
	}

	var log pb.GroupLog
	if err = proto.Unmarshal(value, &log); err != nil {
		return nil, fmt.Errorf("proto unmarshal error: %w", err)
	}

	return &log, nil
}

// SaveLog 保存日志
func (a *AdminDs) SaveLog(ctx context.Context, msg *pb.GroupLog) error {

	bs, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("bs proto.marshal error: %w", err)
	}

	if err = a.Put(ctx, adminDsKey.AdminLogKey(msg.GroupId, msg.Id), bs); err != nil {
		return err
	}

	return nil
}

// GetLogHead 获取日志头ID
func (a *AdminDs) GetLogHead(ctx context.Context, groupID string) (string, error) {
	pblog, err := a.getOldestLog(ctx, groupID, "")
	if err != nil {
		return "", fmt.Errorf("getOldestLog error: %w", err)
	}

	return pblog.Id, nil
}

// GetLogTail 获取日志尾ID
func (a *AdminDs) GetLogTail(ctx context.Context, groupID string) (string, error) {
	pblog, err := a.getLatestLog(ctx, groupID, "")
	if err != nil {
		return "", fmt.Errorf("getOldestLog error: %w", err)
	}

	return pblog.Id, nil
}

// GetLogLength 获取日志长度
func (a *AdminDs) GetLogLength(ctx context.Context, groupID string) (int, error) {
	result, err := a.Query(ctx, query.Query{
		Prefix:   adminDsKey.AdminLogPrefix(groupID),
		Orders:   []query.Order{query.OrderByKey{}},
		KeysOnly: true,
	})
	if err != nil {
		return 0, fmt.Errorf("ds query error: %w", err)
	}

	var num int
	for entry := range result.Next() {
		if entry.Error != nil {
			return 0, fmt.Errorf("ds result next error: %w", entry.Error)
		}
		num++
	}

	return num, nil
}

// GetState 获取群组状态
func (a *AdminDs) GetState(ctx context.Context, groupID string) (string, error) {

	value, err := a.Get(ctx, adminDsKey.StateKey(groupID))
	if err != nil {
		return "", fmt.Errorf("ds get state error: %w", err)
	}

	return string(value), nil
}

// SetState 设置群组状态
func (a *AdminDs) SetState(ctx context.Context, groupID string, state string) error {

	if err := a.Put(ctx, adminDsKey.StateKey(groupID), []byte(state)); err != nil {
		return fmt.Errorf("ds set state error: %w", err)
	}

	return nil
}

// GetName 获取组名
func (a *AdminDs) GetName(ctx context.Context, groupID string) (string, error) {

	value, err := a.Get(ctx, adminDsKey.NameKey(groupID))
	if err != nil {
		return "", fmt.Errorf("ds get name error: %w", err)
	}

	return string(value), nil
}

// SetName 设置组名
func (a *AdminDs) SetName(ctx context.Context, groupID string, name string) error {

	if err := a.Put(ctx, adminDsKey.NameKey(groupID), []byte(name)); err != nil {
		return fmt.Errorf("ds set name error: %w", err)
	}

	return nil
}

// GetName 获取组别名
func (a *AdminDs) GetAlias(ctx context.Context, groupID string) (string, error) {

	value, err := a.Get(ctx, adminDsKey.AliasKey(groupID))
	if err != nil {
		return "", fmt.Errorf("ds get alias error: %w", err)
	}

	return string(value), nil
}

// SetName 设置组别名
func (a *AdminDs) SetAlias(ctx context.Context, groupID string, alias string) error {

	if err := a.Put(ctx, adminDsKey.AliasKey(groupID), []byte(alias)); err != nil {
		return fmt.Errorf("ds set alias error: %w", err)
	}

	return nil
}

// GetName 获取组头像
func (a *AdminDs) GetAvatar(ctx context.Context, groupID string) (string, error) {

	value, err := a.Get(ctx, adminDsKey.AvatarKey(groupID))
	if err != nil {
		return "", fmt.Errorf("ds get avatar error: %w", err)
	}

	return string(value), nil
}

// SetName 设置组头像
func (a *AdminDs) SetAvatar(ctx context.Context, groupID string, avatar string) error {

	if err := a.Put(ctx, adminDsKey.AvatarKey(groupID), []byte(avatar)); err != nil {
		return fmt.Errorf("ds set avatar error: %w", err)
	}

	return nil
}

// GetName 获取组通知
func (a *AdminDs) GetNotice(ctx context.Context, groupID string) (string, error) {

	value, err := a.Get(ctx, adminDsKey.NoticeKey(groupID))
	if err != nil {
		return "", fmt.Errorf("ds get notice error: %w", err)
	}

	return string(value), nil
}

// GetAutoJoinGroup 获取入群是否免审核
func (a *AdminDs) GetAutoJoinGroup(ctx context.Context, groupID string) (bool, error) {

	value, err := a.Get(ctx, adminDsKey.AutoJoinGroupKey(groupID))
	if err != nil {
		return false, fmt.Errorf("ds get notice error: %w", err)
	}

	return string(value) == "true", nil
}

// GetDepositAddress 获取群组的寄存节点
func (a *AdminDs) GetDepositAddress(ctx context.Context, groupID string) (peer.ID, error) {

	value, err := a.Get(ctx, adminDsKey.DepositAddressKey(groupID))
	if err != nil {
		return peer.ID(""), fmt.Errorf("ds get notice error: %w", err)
	}

	return peer.ID(value), nil
}

// GetName 获取组创建时间
func (a *AdminDs) GetCreateTime(ctx context.Context, groupID string) (int64, error) {

	value, err := a.Get(ctx, adminDsKey.CreateTimeKey(groupID))
	if err != nil {
		return 0, fmt.Errorf("ds get notice error: %w", err)
	}

	createTime, err := strconv.ParseInt(string(value), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse create time error: %w", err)
	}

	return createTime, nil
}

// GetCreator 获取群组创建者
func (a *AdminDs) GetCreator(ctx context.Context, groupID string) (peer.ID, error) {

	value, err := a.Get(ctx, adminDsKey.CreatorKey(groupID))
	if err != nil {
		return "", fmt.Errorf("ds get creator error: %w", err)
	}

	return peer.ID(value), nil
}

// GetSessions 获取会话列表
func (a *AdminDs) GetGroupIDs(ctx context.Context) ([]string, error) {
	results, err := a.Query(ctx, query.Query{
		Prefix: adminDsKey.ListPrefix(),
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var groupIDs []string
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("ds result error: %w", err)
		}

		groupIDs = append(groupIDs, strings.TrimPrefix(result.Key, adminDsKey.ListPrefix()))
	}

	return groupIDs, nil
}

func (a *AdminDs) GetMembers(ctx context.Context, groupID string) ([]*pb.GroupMember, error) {

	value, err := a.Get(ctx, adminDsKey.MembersKey(groupID))
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("ds get member error: %w", err)
	}

	var pbmsg pb.GroupMembers
	if err = proto.Unmarshal(value, &pbmsg); err != nil {
		return nil, fmt.Errorf("proto.Unmarshal error: %w", err)
	}

	fmt.Println("size", len(value))
	fmt.Println("pbmsg: ", pbmsg.String())
	fmt.Println("pbmember: ", pbmsg.Members)

	return pbmsg.Members, nil
}

// SetListID 设置会话
func (a *AdminDs) SetListID(ctx context.Context, groupID string) error {
	if err := a.Put(ctx, adminDsKey.ListKey(groupID), []byte(strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
		return fmt.Errorf("ds set session error: %w", err)
	}
	return nil
}

// DeleteListID 删除会话
func (a *AdminDs) DeleteListID(ctx context.Context, groupID string) error {
	if err := a.Delete(ctx, adminDsKey.ListKey(groupID)); err != nil {
		return fmt.Errorf("ds delete session key error: %w", err)
	}
	return nil
}

func (a *AdminDs) DeleteGroup(ctx context.Context, groupID string) error {

	if err := a.Delete(ctx, adminDsKey.ListKey(groupID)); err != nil {
		return fmt.Errorf("ds delete session key error: %w", err)
	}

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

	return nil
}
