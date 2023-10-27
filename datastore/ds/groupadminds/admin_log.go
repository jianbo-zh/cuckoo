package groupadminds

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/datastore/filter"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/grouppb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

// UpdateCreator 更新创建者
func (a *AdminDs) UpdateCreator(ctx context.Context, groupID string) error {

	pblog, err := a.getOldestLog(ctx, groupID, KwCreator)
	if err != nil {
		return fmt.Errorf("getOldestLog error: %w", err)
	}

	if err := a.Put(ctx, adminDsKey.CreatorKey(groupID), pblog.PeerId); err != nil {
		return fmt.Errorf("ds put creator key error: %w", err)
	}

	if err := a.Put(ctx, adminDsKey.CreateTimeKey(groupID), []byte(strconv.FormatInt(pblog.CreateTime, 10))); err != nil {
		return fmt.Errorf("ds put create time key error: %w", err)
	}

	if err := a.Put(ctx, adminDsKey.StateKey(groupID), []byte(mytype.GroupStateNormal)); err != nil {
		return fmt.Errorf("ds put state key error: %w", err)
	}

	return nil
}

// UpdateName 更新群组名称
func (a *AdminDs) UpdateName(ctx context.Context, groupID string) error {

	pblog, err := a.getLatestLog(ctx, groupID, KwName)
	if err != nil {
		return fmt.Errorf("getOldestLog error: %w", err)
	}

	if err := a.Put(ctx, adminDsKey.NameKey(groupID), pblog.Payload); err != nil {
		return fmt.Errorf("ds put name key error: %w", err)
	}

	return nil
}

// UpdateAvatar 更新头像
func (a *AdminDs) UpdateAvatar(ctx context.Context, groupID string) error {

	pblog, err := a.getLatestLog(ctx, groupID, KwAvatar)
	if err != nil {
		return fmt.Errorf("getOldestLog error: %w", err)
	}

	if err := a.Put(ctx, adminDsKey.AvatarKey(groupID), pblog.Payload); err != nil {
		return fmt.Errorf("ds put avatar key error: %w", err)
	}

	return nil
}

// UpdateNotice 更新通知
func (a *AdminDs) UpdateNotice(ctx context.Context, groupID string) error {

	pblog, err := a.getLatestLog(ctx, groupID, KwNotice)
	if err != nil {
		return fmt.Errorf("getOldestLog error: %w", err)
	}

	if err := a.Put(ctx, adminDsKey.NoticeKey(groupID), pblog.Payload); err != nil {
		return fmt.Errorf("ds put notice key error: %w", err)
	}

	return nil
}

// UpdateAutoJoinGroup 更新自动加入组
func (a *AdminDs) UpdateAutoJoinGroup(ctx context.Context, groupID string) error {

	pblog, err := a.getLatestLog(ctx, groupID, KwAutoJoin)
	if err != nil {
		return fmt.Errorf("getOldestLog error: %w", err)
	}

	if err := a.Put(ctx, adminDsKey.AutoJoinGroupKey(groupID), pblog.Payload); err != nil {
		return fmt.Errorf("ds put auto join group key error: %w", err)
	}

	return nil
}

// UpdateDepositAddress 更新寄存地址
func (a *AdminDs) UpdateDepositAddress(ctx context.Context, groupID string) error {

	pblog, err := a.getLatestLog(ctx, groupID, KwDepositAddress)
	if err != nil {
		return fmt.Errorf("getOldestLog error: %w", err)
	}

	if err := a.Put(ctx, adminDsKey.DepositAddressKey(groupID), pblog.Payload); err != nil {
		return fmt.Errorf("ds put deposit address error: %w", err)
	}

	return nil
}

func (a *AdminDs) UpdateMembers(ctx context.Context, groupID string, hostID peer.ID) error {
	pblogs, err := a.getGroupLogs(ctx, groupID, KwMember)
	if err != nil {
		return fmt.Errorf("getGroupLogs error: %w", err)
	}

	refuseLogs := make(map[peer.ID]string)
	allPeers := make(map[peer.ID]*pb.GroupMember)

	// 多条件判断： 0-agree, 1-apply
	var peers = make(map[peer.ID][2]bool)
	for _, pblog := range pblogs {

		memberID := peer.ID(pblog.Member.Id)
		if _, exists := peers[memberID]; !exists {
			peers[memberID] = [2]bool{false, false}
		}

		allPeers[memberID] = pblog.Member

		peerCond := peers[memberID]

		switch pblog.MemberOperate {
		case pb.GroupLog_CREATOR:
			peerCond[0] = true
			peerCond[1] = true

		case pb.GroupLog_APPLY:
			peerCond[1] = true

		case pb.GroupLog_AGREE:
			peerCond[0] = true

		case pb.GroupLog_REJECTED:
			peerCond[0] = false

		case pb.GroupLog_REMOVE, pb.GroupLog_EXIT:
			peerCond[0] = false
			peerCond[1] = false
			// 移除或退出日志
			refuseLogs[memberID] = pblog.Id
		default:
			return fmt.Errorf("unsupport member operate")
		}

		// 记得还回去，数组是值拷贝
		peers[memberID] = peerCond
	}

	var members pb.GroupMembers
	var agreePeerIds pb.GroupAgreePeerIds
	var refusePeers pb.GroupRefusePeers

	groupState := mytype.GroupStateUnknown

	for peerID, cond := range peers {

		if cond[0] && cond[1] {
			members.Members = append(members.Members, allPeers[peerID])
			agreePeerIds.PeerIds = append(agreePeerIds.PeerIds, []byte(peerID)) // 正常会员必然是可以连接会员

			if peerID == hostID {
				groupState = mytype.GroupStateNormal
			}

		} else if cond[0] {
			agreePeerIds.PeerIds = append(agreePeerIds.PeerIds, []byte(peerID))

			if peerID == hostID {
				groupState = mytype.GroupStateAgree
			}

		} else if cond[1] {
			if peerID == hostID {
				groupState = mytype.GroupStateApply
			}

		} else {
			refusePeers.Peers = append(refusePeers.Peers, &pb.GroupRefuseLog{
				PeerId: []byte(peerID),
				LogId:  refuseLogs[peerID],
			})

			if peerID == hostID {
				groupState = mytype.GroupStateExit
			}
		}
	}

	// 正式成员
	if bs, err := proto.Marshal(&members); err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)

	} else {
		if err = a.Put(ctx, adminDsKey.MembersKey(groupID), bs); err != nil {
			return fmt.Errorf("ds put member key error: %w", err)
		}
	}

	// 允许连接成员
	if bs, err := proto.Marshal(&agreePeerIds); err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)

	} else {
		if err = a.Put(ctx, adminDsKey.AgreePeersKey(groupID), bs); err != nil {
			return fmt.Errorf("ds put agree member key error: %w", err)
		}
	}

	// 拒绝连接成员
	if bs, err := proto.Marshal(&refusePeers); err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)

	} else {
		if err = a.Put(ctx, adminDsKey.RefusePeersKey(groupID), bs); err != nil {
			return fmt.Errorf("ds put refuse member key error: %w", err)
		}
	}

	// 群组状态判断
	if pblog, err := a.getLatestLog(ctx, groupID, KwDisband); err != nil {
		return fmt.Errorf("get disband error: %w", err)

	} else if pblog.Id == "" {
		// not disband, so can set
		if groupState != "" {
			if err := a.Put(ctx, adminDsKey.StateKey(groupID), []byte(groupState)); err != nil {
				return fmt.Errorf("ds put state key error: %w", err)
			}
		}
	}

	return nil
}

// UpdateDisband 更新解散
func (a *AdminDs) UpdateDisband(ctx context.Context, groupID string) error {

	if _, err := a.getLatestLog(ctx, groupID, KwDisband); err != nil {
		return fmt.Errorf("getLatestLog error: %w", err)
	}

	// found, so set
	if err := a.Put(ctx, adminDsKey.StateKey(groupID), []byte(mytype.GroupStateDisband)); err != nil {
		return fmt.Errorf("ds put state key error: %w", err)
	}

	return nil
}

// GetMemberIDs 获取群组成员IDs
func (a *AdminDs) GetMemberIDs(ctx context.Context, groupID string) ([]peer.ID, error) { // 正式成员IDs

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

	var memberIDs []peer.ID
	for _, member := range pbmsg.Members {
		memberIDs = append(memberIDs, peer.ID(member.Id))
	}

	return memberIDs, nil
}

// GetAgreePeerIDs 获取允许连接的PeerIDs列表
func (a *AdminDs) GetAgreePeerIDs(ctx context.Context, groupID string) ([]peer.ID, error) { // 所有审核通过的成员IDs
	value, err := a.Get(ctx, adminDsKey.AgreePeersKey(groupID))
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("ds get agree peers error: %w", err)
	}

	var pbmsg pb.GroupAgreePeerIds
	if err = proto.Unmarshal(value, &pbmsg); err != nil {
		return nil, fmt.Errorf("proto.Unmarshal error: %w", err)
	}

	var peerIDs []peer.ID
	for _, peerid := range pbmsg.PeerIds {
		peerIDs = append(peerIDs, peer.ID(peerid))
	}

	return peerIDs, nil
}

// GetRefusePeerLogs 获取拒绝连接的PeerLog
func (a *AdminDs) GetRefusePeerLogs(ctx context.Context, groupID string) (map[peer.ID]string, error) { // 所有拒绝连接的成员
	value, err := a.Get(ctx, adminDsKey.RefusePeersKey(groupID))
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("ds get refuse peers error: %w", err)
	}

	var pbmsg pb.GroupRefusePeers
	if err = proto.Unmarshal(value, &pbmsg); err != nil {
		return nil, fmt.Errorf("proto.Unmarshal error: %w", err)
	}

	refusePeerLogs := make(map[peer.ID]string)
	for _, log := range pbmsg.Peers {
		refusePeerLogs[peer.ID(log.PeerId)] = log.LogId
	}

	return refusePeerLogs, nil
}

// getLatestLog 最新日志
func (a *AdminDs) getLatestLog(ctx context.Context, groupID string, keywords string) (*pb.GroupLog, error) {
	var filters []query.Filter
	if keywords != "" {
		filters = []query.Filter{filter.GroupContainFilter{
			Keywords: keywords,
		}}
	}
	result, err := a.Query(ctx, query.Query{
		Prefix:  adminDsKey.AdminLogPrefix(groupID),
		Orders:  []query.Order{query.OrderByKeyDescending{}}, // desc, limit 1
		Filters: filters,
		Limit:   1,
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var pblog pb.GroupLog
	for entry := range result.Next() {
		if entry.Error != nil {
			return nil, fmt.Errorf("ds result next error: %w", entry.Error)
		}
		pblog.Reset()
		if err := proto.Unmarshal(entry.Value, &pblog); err != nil {
			return nil, fmt.Errorf("proto unmarshal error: %w", err)
		}
	}
	return &pblog, nil
}

// getOldestLog 最早的日子
func (a *AdminDs) getOldestLog(ctx context.Context, groupID string, keywords string) (*pb.GroupLog, error) {
	var filters []query.Filter
	if keywords != "" {
		filters = []query.Filter{filter.GroupContainFilter{
			Keywords: keywords,
		}}
	}
	result, err := a.Query(ctx, query.Query{
		Prefix:  adminDsKey.AdminLogPrefix(groupID),
		Orders:  []query.Order{query.OrderByKey{}}, // asc, limit 1
		Filters: filters,
		Limit:   1,
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var pblog pb.GroupLog
	for entry := range result.Next() {
		if entry.Error != nil {
			return nil, fmt.Errorf("ds result next error: %w", entry.Error)
		}
		pblog.Reset()
		if err := proto.Unmarshal(entry.Value, &pblog); err != nil {
			return nil, fmt.Errorf("proto unmarshal error: %w", err)
		}
	}
	return &pblog, nil
}

// getGroupLogs 获取一组日志
func (a *AdminDs) getGroupLogs(ctx context.Context, groupID string, keywords string) ([]*pb.GroupLog, error) {
	var filters []query.Filter
	if keywords != "" {
		filters = []query.Filter{filter.GroupContainFilter{
			Keywords: keywords,
		}}
	}
	result, err := a.Query(ctx, query.Query{
		Prefix:  adminDsKey.AdminLogPrefix(groupID),
		Orders:  []query.Order{query.OrderByKey{}}, // asc
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var pblogs []*pb.GroupLog
	for entry := range result.Next() {
		if entry.Error != nil {
			return nil, fmt.Errorf("ds result next error: %w", entry.Error)
		}

		var pblog pb.GroupLog
		if err := proto.Unmarshal(entry.Value, &pblog); err != nil {
			return nil, fmt.Errorf("proto unmarshal error: %w", err)
		}

		pblogs = append(pblogs, &pblog)
	}

	return pblogs, nil
}
