package ds

import (
	"context"
	"encoding/json"
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

func (a *AdminDs) GetLog(ctx context.Context, groupID string, logID string) (*pb.Log, error) {
	value, err := a.Get(ctx, adminDsKey.AdminLogKey(groupID, logID))
	if err != nil {
		return nil, fmt.Errorf("ds get admin log error: %w", err)
	}

	var log *pb.Log
	if err = proto.Unmarshal(value, log); err != nil {
		return nil, fmt.Errorf("proto unmarshal error: %w", err)
	}

	return log, nil
}

func (a *AdminDs) SaveLog(ctx context.Context, msg *pb.Log) error {

	_, err := a.Get(ctx, adminDsKey.AdminLogKey(msg.GroupId, msg.Id))
	if err == nil {
		return nil

	} else if !errors.Is(err, ds.ErrNotFound) {
		return fmt.Errorf("ds get key error: %w", err)
	}

	// 保存日志
	bs, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("bs proto.marshal error: %w", err)
	}
	if err = a.Put(ctx, adminDsKey.AdminLogKey(msg.GroupId, msg.Id), bs); err != nil {
		return err
	}

	// 更新统计
	logRecorder, err := a.getLogRecorder(ctx, msg.GroupId)
	if err != nil {
		return fmt.Errorf("ds get log recorder error: %w", err)
	}

	if logRecorder.logHead == "" || msg.Id < logRecorder.logHead {
		logRecorder.logHead = msg.Id
	}
	if msg.Id > logRecorder.logTail {
		logRecorder.logTail = msg.Id
	}
	logRecorder.logLength++

	// 更新缓存
	err = a.syncLogCache(ctx, msg.GroupId, msg)
	if err != nil {
		return fmt.Errorf("update log cache error: %w", err)
	}

	return nil
}

// GetLogHead
func (a *AdminDs) GetLogHead(ctx context.Context, groupID string) (string, error) {

	logrecorder, err := a.getLogRecorder(ctx, groupID)
	if err != nil {
		return "", fmt.Errorf("ds get log recorder error: %w", err)
	}

	return logrecorder.logHead, nil
}

// GetLogTail
func (a *AdminDs) GetLogTail(ctx context.Context, groupID string) (string, error) {

	logrecorder, err := a.getLogRecorder(ctx, groupID)
	if err != nil {
		return "", fmt.Errorf("ds get log recorder error: %w", err)
	}

	return logrecorder.logTail, nil
}

// GetLogLength
func (a *AdminDs) GetLogLength(ctx context.Context, groupID string) (int, error) {

	logrecorder, err := a.getLogRecorder(ctx, groupID)
	if err != nil {
		return 0, fmt.Errorf("ds get log recorder error: %w", err)
	}

	return logrecorder.logLength, nil
}

// GetState 获取群组状态
func (a *AdminDs) GetState(ctx context.Context, groupID string) (string, error) {

	if err := a.checkInitCache(ctx, groupID); err != nil {
		return "", fmt.Errorf("check init cache error: %w", err)
	}

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

	if err := a.checkInitCache(ctx, groupID); err != nil {
		return "", fmt.Errorf("check init cache error: %w", err)
	}

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

	if err := a.checkInitCache(ctx, groupID); err != nil {
		return "", fmt.Errorf("check init cache error: %w", err)
	}

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

	if err := a.checkInitCache(ctx, groupID); err != nil {
		return "", fmt.Errorf("check init cache error: %w", err)
	}

	value, err := a.Get(ctx, adminDsKey.NoticeKey(groupID))
	if err != nil {
		return "", fmt.Errorf("ds get notice error: %w", err)
	}

	return string(value), nil
}

// GetAutoJoinGroup 获取入群是否免审核
func (a *AdminDs) GetAutoJoinGroup(ctx context.Context, groupID string) (bool, error) {

	if err := a.checkInitCache(ctx, groupID); err != nil {
		return false, fmt.Errorf("check init cache error: %w", err)
	}

	value, err := a.Get(ctx, adminDsKey.AutoJoinGroupKey(groupID))
	if err != nil {
		return false, fmt.Errorf("ds get notice error: %w", err)
	}

	return string(value) == "true", nil
}

// GetDepositAddress 获取群组的寄存节点
func (a *AdminDs) GetDepositAddress(ctx context.Context, groupID string) (peer.ID, error) {

	if err := a.checkInitCache(ctx, groupID); err != nil {
		return peer.ID(""), fmt.Errorf("check init cache error: %w", err)
	}

	value, err := a.Get(ctx, adminDsKey.DepositAddressKey(groupID))
	if err != nil {
		return peer.ID(""), fmt.Errorf("ds get notice error: %w", err)
	}

	return peer.ID(value), nil
}

// GetName 获取组创建时间
func (a *AdminDs) GetCreateTime(ctx context.Context, groupID string) (int64, error) {

	if err := a.checkInitCache(ctx, groupID); err != nil {
		return 0, fmt.Errorf("check init cache error: %w", err)
	}

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

	if err := a.checkInitCache(ctx, groupID); err != nil {
		return "", fmt.Errorf("check init cache error: %w", err)
	}

	value, err := a.Get(ctx, adminDsKey.CreatorKey(groupID))
	if err != nil {
		return "", fmt.Errorf("ds get creator error: %w", err)
	}

	return peer.ID(value), nil
}

// GetSessions 获取会话列表
func (a *AdminDs) GetSessionIDs(ctx context.Context) ([]string, error) {
	results, err := a.Query(ctx, query.Query{
		Prefix: adminDsKey.SessionPrefix(),
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var groupIDs []string
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("ds result error: %w", err)
		}

		groupIDs = append(groupIDs, strings.TrimPrefix(result.Key, adminDsKey.SessionPrefix()))
	}

	return groupIDs, nil
}

func (a *AdminDs) GetMembers(ctx context.Context, groupID string) ([]types.GroupMember, error) {

	if err := a.checkInitCache(ctx, groupID); err != nil {
		return nil, fmt.Errorf("check init cache error: %w", err)
	}

	value, err := a.Get(ctx, adminDsKey.MembersKey(groupID))
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("ds get member error: %w", err)
	}

	var members []types.GroupMember
	if err = json.Unmarshal(value, &members); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	return members, nil
}

func (a *AdminDs) GetAgreeMembers(ctx context.Context, groupID string) ([]types.GroupMember, error) {

	if err := a.checkInitCache(ctx, groupID); err != nil {
		return nil, fmt.Errorf("check init cache error: %w", err)
	}

	value, err := a.Get(ctx, adminDsKey.AgreeMembersKey(groupID))
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("ds get member error: %w", err)
	}

	var members []types.GroupMember
	if err = json.Unmarshal(value, &members); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	return members, nil
}

// SetSession 设置会话
func (a *AdminDs) SetSession(ctx context.Context, groupID string) error {
	if err := a.Put(ctx, adminDsKey.SessionKey(groupID), []byte(strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
		return fmt.Errorf("ds set session error: %w", err)
	}
	return nil
}

// DeleteSession 删除会话
func (a *AdminDs) DeleteSession(ctx context.Context, groupID string) error {
	if err := a.Delete(ctx, adminDsKey.SessionKey(groupID)); err != nil {
		return fmt.Errorf("ds delete session key error: %w", err)
	}
	return nil
}

func (a *AdminDs) DeleteGroup(ctx context.Context, groupID string) error {

	if err := a.Delete(ctx, adminDsKey.SessionKey(groupID)); err != nil {
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

func (a *AdminDs) getLogRecorder(ctx context.Context, groupID string) (*LogRecorder, error) {

	if err := a.checkInitCache(ctx, groupID); err != nil {
		return nil, fmt.Errorf("check init cache error: %w", err)
	}

	return a.logRecords[groupID], nil
}

func (a *AdminDs) checkInitCache(ctx context.Context, groupID string) error {
	if _, exists := a.logRecords[groupID]; exists {
		return nil
	}

	a.logMutex.Lock()
	defer a.logMutex.Unlock()

	return a.initLogCache(ctx, groupID)
}

// uploadLogCache 更新群组记录器
func (a *AdminDs) initLogCache(ctx context.Context, groupID string) error {
	if _, exists := a.logRecords[groupID]; exists {
		return nil
	}

	result, err := a.Query(ctx, query.Query{
		Prefix: adminDsKey.AdminLogPrefix(groupID),
		Orders: []query.Order{query.OrderByKey{}}, // asc
	})
	if err != nil {
		return fmt.Errorf("ds query error: %w", err)
	}

	var logHead, logTail string
	var logLength int
	var logCreator []byte
	var logCreateTime []byte
	var logName, logAvatar, logNotice []byte
	var logAutoJoinGroup, depositPeerID []byte

	formalMap := make(map[peer.ID]bool)
	operateMap := make(map[peer.ID]pb.Log_MemberOperate)
	allMemberMap := make(map[peer.ID]*pb.Log_Member)

	for entry := range result.Next() {
		if entry.Error != nil {
			return fmt.Errorf("ds result next error: %w", entry.Error)
		}

		if logHead == "" {
			logHead = strings.TrimPrefix(string(entry.Key), adminDsKey.AdminLogPrefix(groupID))
		}

		logTail = strings.TrimPrefix(string(entry.Key), adminDsKey.AdminLogPrefix(groupID))
		logLength++

		var pblog pb.Log
		if err = proto.Unmarshal(entry.Value, &pblog); err != nil {
			return fmt.Errorf("ds proto unmarshal error: %w", err)
		}

		fmt.Println("------ ", pblog.String())

		switch pblog.LogType {
		case pb.Log_CREATE:
			logCreator = pblog.PeerId
			logCreateTime = []byte(strconv.FormatInt(pblog.CreateTime, 10))
		case pb.Log_NAME:
			logName = pblog.Payload
		case pb.Log_AVATAR:
			logAvatar = pblog.Payload
		case pb.Log_NOTICE:
			logNotice = pblog.Payload
		case pb.Log_AUTO_JOIN_GROUP:
			logAutoJoinGroup = pblog.Payload // "true" or "false"
		case pb.Log_DEPOSIT_PEER_ID:
			depositPeerID = pblog.Payload
		case pb.Log_MEMBER:
			if _, exists := allMemberMap[peer.ID(pblog.Member.Id)]; !exists {
				allMemberMap[peer.ID(pblog.Member.Id)] = pblog.Member
			}

			if pblog.MemberOperate == pb.Log_CREATOR {
				formalMap[peer.ID(pblog.Member.Id)] = true
				continue
			}

			if lastState, exists := operateMap[peer.ID(pblog.Member.Id)]; !exists {
				operateMap[peer.ID(pblog.Member.Id)] = pblog.MemberOperate

			} else {
				if pblog.MemberOperate == lastState {
					continue
				}

				switch lastState {

				case pb.Log_REMOVE, pb.Log_EXIT:
					delete(operateMap, peer.ID(pblog.Member.Id))
					delete(formalMap, peer.ID(pblog.Member.Id))

				case pb.Log_REJECTED:
					delete(operateMap, peer.ID(pblog.Member.Id))

				case pb.Log_AGREE:
					if pblog.MemberOperate == pb.Log_APPLY {
						formalMap[peer.ID(pblog.Member.Id)] = true
					}

				case pb.Log_APPLY:
					if pblog.MemberOperate == pb.Log_AGREE {
						formalMap[peer.ID(pblog.Member.Id)] = true
					}

				default:
					operateMap[peer.ID(pblog.Member.Id)] = pblog.MemberOperate
				}
			}
		default:
			// other nothing to do
		}
	}

	if len(logCreator) > 0 {
		if err = a.Put(ctx, adminDsKey.CreatorKey(groupID), logCreator); err != nil {
			return fmt.Errorf("ds put creator key error: %w", err)
		}
	}

	if len(logCreateTime) > 0 {
		if err = a.Put(ctx, adminDsKey.CreateTimeKey(groupID), logCreateTime); err != nil {
			return fmt.Errorf("ds put create time key error: %w", err)
		}
	}

	if len(logName) > 0 {
		if err = a.Put(ctx, adminDsKey.NameKey(groupID), logName); err != nil {
			return fmt.Errorf("ds put name key error: %w", err)
		}
	}

	if len(logAvatar) > 0 {
		if err = a.Put(ctx, adminDsKey.AvatarKey(groupID), logAvatar); err != nil {
			return fmt.Errorf("ds put avatar key error: %w", err)
		}
	}
	if len(logNotice) > 0 {
		if err = a.Put(ctx, adminDsKey.NoticeKey(groupID), logNotice); err != nil {
			return fmt.Errorf("ds put notice key error: %w", err)
		}
	}
	if len(logAutoJoinGroup) > 0 {
		if err = a.Put(ctx, adminDsKey.AutoJoinGroupKey(groupID), logAutoJoinGroup); err != nil {
			return fmt.Errorf("ds put auto join group key error: %w", err)
		}
	}
	if len(depositPeerID) > 0 {
		if err = a.Put(ctx, adminDsKey.DepositAddressKey(groupID), depositPeerID); err != nil {
			return fmt.Errorf("ds put deposit peer id key error: %w", err)
		}
	}

	// 正式成员
	var members []types.GroupMember
	for memberID := range formalMap {
		members = append(members, types.GroupMember{
			ID:     memberID,
			Name:   allMemberMap[memberID].Name,
			Avatar: allMemberMap[memberID].Avatar,
		})
	}

	bsMembers, err := json.Marshal(members)
	if err != nil {
		return fmt.Errorf("json marshal member error: %w", err)
	}
	if err = a.Put(ctx, adminDsKey.MembersKey(groupID), bsMembers); err != nil {
		return fmt.Errorf("ds put member key error: %w", err)
	}

	// 允许成员
	agreeMembers := members
	for memberID, operate := range operateMap {
		if operate == pb.Log_AGREE {
			if _, exists := formalMap[memberID]; !exists {
				agreeMembers = append(agreeMembers, types.GroupMember{
					ID:     memberID,
					Name:   allMemberMap[memberID].Name,
					Avatar: allMemberMap[memberID].Avatar,
				})
			}
		}
	}
	bsAgreeMembers, err := json.Marshal(agreeMembers)
	if err != nil {
		return fmt.Errorf("json marshal member error: %w", err)
	}
	if err = a.Put(ctx, adminDsKey.AgreeMembersKey(groupID), bsAgreeMembers); err != nil {
		return fmt.Errorf("ds put agree member key error: %w", err)
	}

	a.logRecords[groupID] = &LogRecorder{
		logHead:   logHead,
		logTail:   logTail,
		logLength: logLength,
	}
	return nil
}

func (a *AdminDs) syncLogCache(ctx context.Context, groupID string, synclog *pb.Log) error {
	switch synclog.LogType {
	case pb.Log_CREATE:
		if err := a.Put(ctx, adminDsKey.CreatorKey(groupID), synclog.PeerId); err != nil {
			return fmt.Errorf("ds put creator key error: %w", err)
		}
		if err := a.Put(ctx, adminDsKey.CreateTimeKey(groupID), []byte(strconv.FormatInt(synclog.CreateTime, 10))); err != nil {
			return fmt.Errorf("ds put create time key error: %w", err)
		}
	case pb.Log_NAME:
		result, err := a.Query(ctx, query.Query{
			Prefix: adminDsKey.AdminLogPrefix(groupID),
			Orders: []query.Order{query.OrderByKeyDescending{}}, // desc, limit 1
			Filters: []query.Filter{GroupContainFilter{
				Keywords: KwName,
			}},
			Limit: 1,
		})
		if err != nil {
			return fmt.Errorf("ds query error: %w", err)
		}

		var name []byte
		for entry := range result.Next() {
			if entry.Error != nil {
				return fmt.Errorf("ds result next error: %w", entry.Error)
			}

			var pblog pb.Log
			if err := proto.Unmarshal(entry.Value, &pblog); err != nil {
				return fmt.Errorf("proto unmarshal error: %w", err)
			}

			name = pblog.Payload
		}

		if err = a.Put(ctx, adminDsKey.NameKey(groupID), name); err != nil {
			return fmt.Errorf("ds save group name error: %w", err)
		}

	case pb.Log_AVATAR:
		result, err := a.Query(ctx, query.Query{
			Prefix: adminDsKey.AdminLogPrefix(groupID),
			Orders: []query.Order{query.OrderByKeyDescending{}}, // desc, limit 1
			Filters: []query.Filter{GroupContainFilter{
				Keywords: KwAvatar,
			}},
			Limit: 1,
		})
		if err != nil {
			return fmt.Errorf("ds query error: %w", err)
		}

		var avatar []byte
		for entry := range result.Next() {
			if entry.Error != nil {
				return fmt.Errorf("ds result next error: %w", entry.Error)
			}

			var pblog pb.Log
			if err := proto.Unmarshal(entry.Value, &pblog); err != nil {
				return fmt.Errorf("proto unmarshal error: %w", err)
			}

			avatar = pblog.Payload
		}

		if err = a.Put(ctx, adminDsKey.AvatarKey(groupID), avatar); err != nil {
			return fmt.Errorf("ds save group avatar error: %w", err)
		}

	case pb.Log_NOTICE:
		result, err := a.Query(ctx, query.Query{
			Prefix: adminDsKey.AdminLogPrefix(groupID),
			Orders: []query.Order{query.OrderByKeyDescending{}}, // desc, limit 1
			Filters: []query.Filter{GroupContainFilter{
				Keywords: KwNotice,
			}},
			Limit: 1,
		})
		if err != nil {
			return fmt.Errorf("ds query error: %w", err)
		}

		var notice []byte
		for entry := range result.Next() {
			if entry.Error != nil {
				return fmt.Errorf("ds result next error: %w", entry.Error)
			}

			var pblog pb.Log
			if err := proto.Unmarshal(entry.Value, &pblog); err != nil {
				return fmt.Errorf("proto unmarshal error: %w", err)
			}

			notice = pblog.Payload
		}

		if err = a.Put(ctx, adminDsKey.NoticeKey(groupID), notice); err != nil {
			return fmt.Errorf("ds save group notice error: %w", err)
		}
	case pb.Log_AUTO_JOIN_GROUP:
		result, err := a.Query(ctx, query.Query{
			Prefix: adminDsKey.AdminLogPrefix(groupID),
			Orders: []query.Order{query.OrderByKeyDescending{}}, // desc, limit 1
			Filters: []query.Filter{GroupContainFilter{
				Keywords: KwAutoJoin,
			}},
			Limit: 1,
		})
		if err != nil {
			return fmt.Errorf("ds query error: %w", err)
		}

		var autoJoinGroup []byte
		for entry := range result.Next() {
			if entry.Error != nil {
				return fmt.Errorf("ds result next error: %w", entry.Error)
			}

			var pblog pb.Log
			if err := proto.Unmarshal(entry.Value, &pblog); err != nil {
				return fmt.Errorf("proto unmarshal error: %w", err)
			}

			autoJoinGroup = pblog.Payload
		}

		if err = a.Put(ctx, adminDsKey.AutoJoinGroupKey(groupID), autoJoinGroup); err != nil {
			return fmt.Errorf("ds save group notice error: %w", err)
		}

	case pb.Log_DEPOSIT_PEER_ID:
		result, err := a.Query(ctx, query.Query{
			Prefix: adminDsKey.AdminLogPrefix(groupID),
			Orders: []query.Order{query.OrderByKeyDescending{}}, // desc, limit 1
			Filters: []query.Filter{GroupContainFilter{
				Keywords: KwDepositPeer,
			}},
			Limit: 1,
		})
		if err != nil {
			return fmt.Errorf("ds query error: %w", err)
		}

		var depositPeerID []byte
		for entry := range result.Next() {
			if entry.Error != nil {
				return fmt.Errorf("ds result next error: %w", entry.Error)
			}

			var pblog pb.Log
			if err := proto.Unmarshal(entry.Value, &pblog); err != nil {
				return fmt.Errorf("proto unmarshal error: %w", err)
			}

			depositPeerID = pblog.Payload
		}

		if err = a.Put(ctx, adminDsKey.DepositAddressKey(groupID), depositPeerID); err != nil {
			return fmt.Errorf("ds save group deposit peer id error: %w", err)
		}

	case pb.Log_MEMBER:
		result, err := a.Query(ctx, query.Query{
			Prefix: adminDsKey.AdminLogPrefix(groupID),
			Orders: []query.Order{query.OrderByKey{}}, // asc
			Filters: []query.Filter{GroupContainFilter{
				Keywords: KwMember,
			}},
		})
		if err != nil {
			return fmt.Errorf("ds query error: %w", err)
		}

		formalMap := make(map[peer.ID]bool)
		operateMap := make(map[peer.ID]pb.Log_MemberOperate)
		allMemberMap := make(map[peer.ID]*pb.Log_Member)

		for entry := range result.Next() {
			if entry.Error != nil {
				return fmt.Errorf("ds result next error: %w", entry.Error)
			}

			var pblog pb.Log
			if err := proto.Unmarshal(entry.Value, &pblog); err != nil {
				return fmt.Errorf("proto unmarshal error: %w", err)
			}

			if _, exists := allMemberMap[peer.ID(pblog.Member.Id)]; !exists {
				allMemberMap[peer.ID(pblog.Member.Id)] = pblog.Member
			}

			if pblog.MemberOperate == pb.Log_CREATOR {
				formalMap[peer.ID(pblog.Member.Id)] = true
				continue
			}

			if lastState, exists := operateMap[peer.ID(pblog.Member.Id)]; !exists {
				operateMap[peer.ID(pblog.Member.Id)] = pblog.MemberOperate

			} else {
				if pblog.MemberOperate == lastState {
					continue
				}

				switch lastState {

				case pb.Log_REMOVE, pb.Log_EXIT:
					delete(operateMap, peer.ID(pblog.Member.Id))
					delete(formalMap, peer.ID(pblog.Member.Id))

				case pb.Log_REJECTED:
					delete(operateMap, peer.ID(pblog.Member.Id))

				case pb.Log_AGREE:
					if pblog.MemberOperate == pb.Log_APPLY {
						formalMap[peer.ID(pblog.Member.Id)] = true
					}

				case pb.Log_APPLY:
					if pblog.MemberOperate == pb.Log_AGREE {
						formalMap[peer.ID(pblog.Member.Id)] = true
					}

				default:
					operateMap[peer.ID(pblog.Member.Id)] = pblog.MemberOperate
				}
			}
		}

		// 正式成员
		var members []types.GroupMember
		for memberID := range formalMap {
			members = append(members, types.GroupMember{
				ID:     memberID,
				Name:   allMemberMap[memberID].Name,
				Avatar: allMemberMap[memberID].Avatar,
			})
		}

		bsMembers, err := json.Marshal(members)
		if err != nil {
			return fmt.Errorf("json marshal member error: %w", err)
		}
		if err = a.Put(ctx, adminDsKey.MembersKey(groupID), bsMembers); err != nil {
			return fmt.Errorf("ds put member key error: %w", err)
		}

		// 允许成员
		agreeMembers := members
		for memberID, operate := range operateMap {
			if operate == pb.Log_AGREE {
				if _, exists := formalMap[memberID]; !exists {
					agreeMembers = append(agreeMembers, types.GroupMember{
						ID:     memberID,
						Name:   allMemberMap[memberID].Name,
						Avatar: allMemberMap[memberID].Avatar,
					})
				}
			}
		}
		bsAgreeMembers, err := json.Marshal(agreeMembers)
		if err != nil {
			return fmt.Errorf("json marshal member error: %w", err)
		}
		if err = a.Put(ctx, adminDsKey.AgreeMembersKey(groupID), bsAgreeMembers); err != nil {
			return fmt.Errorf("ds put agree member key error: %w", err)
		}

	case pb.Log_DISBAND:
		if err := a.Put(ctx, adminDsKey.StateKey(groupID), []byte("disband")); err != nil {
			return fmt.Errorf("ds save group state error: %w", err)
		}
	}

	return nil
}
