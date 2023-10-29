package groupadminproto

import (
	"context"
	"errors"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/service/groupsvc/protobuf/pb/grouppb"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (a *AdminProto) receiveLog(ctx context.Context, pblog *pb.GroupLog, fromPeerIDs ...peer.ID) error {

	log.Debugln("receiveLog: ")

	oldState, err := a.data.GetState(ctx, pblog.GroupId)

	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return fmt.Errorf("data.GetState error: %w", err)
	}

	isUpdated, err := a.saveLog(ctx, pblog)
	if err != nil {
		return fmt.Errorf("a.saveLog error: %w", err)
	}

	if isUpdated {
		if err := a.broadcastMessage(pblog.GroupId, pblog, fromPeerIDs...); err != nil {
			return fmt.Errorf("a.broadcastMessage error: %w", err)
		}

		if err := a.checkGroupChange(ctx, pblog.GroupId, pblog.LogType == pb.GroupLog_MEMBER, oldState); err != nil {
			return fmt.Errorf(".checkGroupChange error: %w", err)
		}
	}

	return nil
}

func (a *AdminProto) saveLog(ctx context.Context, pblog *pb.GroupLog) (isUpdated bool, err error) {

	// 检查是否已经存在
	if _, err := a.data.GetLog(ctx, pblog.GroupId, pblog.Id); err == nil {
		return false, nil

	} else if !errors.Is(err, ipfsds.ErrNotFound) {
		return false, fmt.Errorf("data.GetLog error: %w", err)
	}

	// 保存日志
	if err := a.data.SaveLog(ctx, pblog); err != nil {
		return false, fmt.Errorf("data.SaveLogs error: %w", err)
	}

	// 更新缓存
	switch pblog.LogType {
	case pb.GroupLog_CREATE:
		if err := a.data.UpdateCreator(ctx, pblog.GroupId); err != nil {
			return true, fmt.Errorf("data.UpdateCreator error: %w", err)
		}
	case pb.GroupLog_NAME:
		if err := a.data.UpdateName(ctx, pblog.GroupId); err != nil {
			return true, fmt.Errorf("data.UpdateName error: %w", err)
		}
	case pb.GroupLog_AVATAR:
		if err := a.data.UpdateAvatar(ctx, pblog.GroupId); err != nil {
			return true, fmt.Errorf("data.UpdateAvatar error: %w", err)
		}

	case pb.GroupLog_NOTICE:
		if err := a.data.UpdateNotice(ctx, pblog.GroupId); err != nil {
			return true, fmt.Errorf("data.UpdateNotice error: %w", err)
		}

	case pb.GroupLog_AUTO_JOIN_GROUP:
		if err := a.data.UpdateAutoJoinGroup(ctx, pblog.GroupId); err != nil {
			return true, fmt.Errorf("data.UpdateAutoJoinGroup error: %w", err)
		}

	case pb.GroupLog_DEPOSIT_PEER_ID:
		if err := a.data.UpdateDepositAddress(ctx, pblog.GroupId); err != nil {
			return true, fmt.Errorf("data.UpdateDepositAddress error: %w", err)
		}

	case pb.GroupLog_MEMBER:
		if err := a.data.UpdateMembers(ctx, pblog.GroupId, a.host.ID()); err != nil {
			return true, fmt.Errorf("data.UpdateMembers error: %w", err)
		}

	case pb.GroupLog_DISBAND:
		if err := a.data.UpdateDisband(ctx, pblog.GroupId); err != nil {
			return true, fmt.Errorf("data.UpdateDisband error: %w", err)
		}

	default:
		return true, fmt.Errorf("unsupport group log type")
	}

	return true, nil
}

// checkGroupChange 检查群组变化
func (a *AdminProto) checkGroupChange(ctx context.Context, groupID string, isMemberOperate bool, oldState string) error {

	log.Debugln("checkGroupChange: ")

	curState, err := a.data.GetState(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return fmt.Errorf("data.GetState error: %w", err)
	}

	if curState != oldState {
		switch curState {
		case mytype.GroupStateExit, mytype.GroupStateDisband:
			// 触发断开连接
			a.emitters.evtGroupsChange.Emit(myevent.EvtGroupsChange{
				DeleteGroupID: groupID,
			})

		case mytype.GroupStateNormal:
			// 触发增加连接
			a.emitters.evtGroupsChange.Emit(myevent.EvtGroupsChange{
				AddGroupID: groupID,
			})
		}

	} else if curState == mytype.GroupStateNormal && isMemberOperate {
		log.Debugln("emit: EvtGroupMemberChange")
		// 正常连接，并且是成员操作，则更新连接成员信息
		// 触发增加连接
		a.emitters.evtGroupMemberChange.Emit(myevent.EvtGroupMemberChange{
			GroupID: groupID,
		})
	}

	return nil
}
