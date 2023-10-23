package groupadminproto

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/grouppb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

// CreateGroup 创建群
func (a *AdminProto) CreateGroup(ctx context.Context, account *mytype.Account, name string, avatar string, content string, members []mytype.Contact) (*mytype.Group, error) {

	hostID := a.host.ID()
	groupID := uuid.NewString()

	// 创建组
	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return nil, err
	}
	createLog := pb.GroupLog{
		Id:      logIdCreate(lamptime, hostID),
		GroupId: groupID,
		PeerId:  []byte(hostID),
		LogType: pb.GroupLog_CREATE,
		Member: &pb.GroupMember{
			Id:     []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MemberOperate: pb.GroupLog_NONE,
		Payload:       []byte(hostID),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if _, err = a.saveLog(ctx, &createLog); err != nil {
		return nil, err
	}

	// 设置名称
	lamptime, err = a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return nil, err
	}
	nameLog := pb.GroupLog{
		Id:            logIdName(lamptime, hostID),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.GroupLog_NAME,
		Member:        nil,
		MemberOperate: pb.GroupLog_NONE,
		Payload:       []byte(name),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}
	if _, err = a.saveLog(ctx, &nameLog); err != nil {
		return nil, err
	}

	// 设置头像
	lamptime, err = a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return nil, err
	}
	avatarLog := pb.GroupLog{
		Id:            logIdAvatar(lamptime, hostID),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.GroupLog_AVATAR,
		Member:        nil,
		MemberOperate: pb.GroupLog_NONE,
		Payload:       []byte(avatar),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}
	if _, err = a.saveLog(ctx, &avatarLog); err != nil {
		return nil, err
	}

	var memberIDs []peer.ID
	memberIDs = append(memberIDs, account.ID)

	// 设置创建者
	lamptime, err = a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return nil, err
	}

	memberLog := pb.GroupLog{
		Id:      logIdMember(lamptime, hostID),
		GroupId: groupID,
		PeerId:  []byte(hostID),
		LogType: pb.GroupLog_MEMBER,
		Member: &pb.GroupMember{
			Id:     []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MemberOperate: pb.GroupLog_CREATOR,
		Payload:       []byte(hostID),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if _, err = a.saveLog(ctx, &memberLog); err != nil {
		return nil, err
	}

	// 设置其他成员
	var groupInvites []myevent.GroupInvite
	for _, member := range members {
		memberIDs = append(memberIDs, member.ID)

		lamptime, err = a.data.TickLamptime(ctx, groupID)
		if err != nil {
			return nil, err
		}

		memberLog := pb.GroupLog{
			Id:      logIdMember(lamptime, hostID),
			GroupId: groupID,
			PeerId:  []byte(hostID),
			LogType: pb.GroupLog_MEMBER,
			Member: &pb.GroupMember{
				Id:     []byte(member.ID),
				Name:   member.Name,
				Avatar: member.Avatar,
			},
			MemberOperate: pb.GroupLog_AGREE,
			Payload:       []byte(member.ID),
			CreateTime:    time.Now().Unix(),
			Lamportime:    lamptime,
			Signature:     []byte(""),
		}

		if _, err = a.saveLog(ctx, &memberLog); err != nil {
			return nil, err
		}

		bs, err := proto.Marshal(&memberLog)
		if err != nil {
			return nil, fmt.Errorf("bs proto.marshal error: %w", err)
		}

		groupInvites = append(groupInvites, myevent.GroupInvite{
			PeerID:         member.ID,
			DepositAddress: member.DepositAddress,
			GroupLog:       bs,
		})
	}

	if err = a.data.SetState(ctx, groupID, mytype.GroupStateNormal); err != nil {
		return nil, fmt.Errorf("ds set group state error: %w", err)
	}

	if err = a.data.SetListID(ctx, groupID); err != nil {
		return nil, fmt.Errorf("ds set group session error: %w", err)
	}

	sessionID := mytype.GroupSessionID(groupID)
	if err = a.sessionData.SetSessionID(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("svc set session id error: %w", err)
	}

	// 触发新增群组
	if err = a.emitters.evtGroupsChange.Emit(myevent.EvtGroupsChange{
		AddGroups: []myevent.Groups{
			{
				GroupID:     groupID,
				PeerIDs:     []peer.ID{account.ID},
				AcptPeerIDs: memberIDs,
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("emite add group error: %w", err)
	}

	// 发送邀请消息
	if err = a.emitters.evtInviteJoinGroup.Emit(myevent.EvtInviteJoinGroup{
		GroupID:     groupID,
		GroupName:   name,
		GroupAvatar: avatar,
		Content:     content,
		Invites:     groupInvites,
	}); err != nil {
		return nil, fmt.Errorf("emit invite join group error: %w", err)
	}

	return &mytype.Group{
		ID:     groupID,
		Name:   name,
		Avatar: avatar,
	}, nil
}

// InviteJoinGroup 邀请进群
func (a *AdminProto) InviteJoinGroup(ctx context.Context, account *mytype.Account, groupID string, contacts []mytype.Contact, content string) error {

	hostID := a.host.ID()

	creatorID, err := a.data.GetCreator(ctx, groupID)
	if err != nil {
		return fmt.Errorf("data.GetCreator error: %w", err)
	}

	if hostID != creatorID {
		return fmt.Errorf("must creator can invite")
	}

	var groupInvites []myevent.GroupInvite
	for _, contact := range contacts {
		lamptime, err := a.data.TickLamptime(ctx, groupID)
		if err != nil {
			return fmt.Errorf("data.TickLamptime error: %w", err)
		}

		pbmsg := pb.GroupLog{
			Id:      logIdMember(lamptime, hostID),
			GroupId: groupID,
			PeerId:  []byte(hostID),
			LogType: pb.GroupLog_MEMBER,
			Member: &pb.GroupMember{
				Id:     []byte(contact.ID),
				Name:   contact.Name,
				Avatar: contact.Avatar,
			},
			MemberOperate: pb.GroupLog_AGREE,
			Payload:       []byte(contact.ID),
			CreateTime:    time.Now().Unix(),
			Lamportime:    lamptime,
			Signature:     []byte(""),
		}

		if _, err = a.saveLog(ctx, &pbmsg); err != nil {
			return fmt.Errorf("data.SaveLog error: %w", err)
		}

		bs, err := proto.Marshal(&pbmsg)
		if err != nil {
			return fmt.Errorf("bs proto.marshal error: %w", err)
		}

		groupInvites = append(groupInvites, myevent.GroupInvite{
			PeerID:         contact.ID,
			DepositAddress: contact.DepositAddress,
			GroupLog:       bs,
		})

		if err = a.broadcastMessage(ctx, groupID, &pbmsg); err != nil {
			return fmt.Errorf("broadcast msg error: %w", err)
		}
	}

	// 触发网络改变
	if err := a.emitGroupMemberChange(ctx, groupID); err != nil {
		return fmt.Errorf("emitGroupMemberChange error: %w", err)
	}

	// 发送邀请信息
	group, err := a.GetGroup(ctx, groupID)
	if err != nil {
		return fmt.Errorf("get group error: %w", err)
	}
	if err = a.emitters.evtInviteJoinGroup.Emit(myevent.EvtInviteJoinGroup{
		GroupID:     group.ID,
		GroupName:   group.Name,
		GroupAvatar: group.Avatar,
		Content:     content,
		Invites:     groupInvites,
	}); err != nil {
		return fmt.Errorf("emit invite join group error: %w", err)
	}

	return nil
}

// AgreeJoinGroup 同意加入群
func (a *AdminProto) AgreeJoinGroup(ctx context.Context, account *mytype.Account, group *mytype.Group, inviteLogBs []byte) error {

	var inviteLog pb.GroupLog
	if err := proto.Unmarshal(inviteLogBs, &inviteLog); err != nil {
		return fmt.Errorf("proto.Unmarshal error: %w", err)
	}

	if err := a.data.MergeLamptime(ctx, group.ID, inviteLog.Lamportime); err != nil {
		return fmt.Errorf("data merge lamptime error: %w", err)
	}

	if _, err := a.saveLog(ctx, &inviteLog); err != nil {
		return fmt.Errorf("data.SaveLog error: %w", err)
	}

	lamptime, err := a.data.TickLamptime(ctx, group.ID)
	if err != nil {
		return fmt.Errorf("data tick lamptime error: %w", err)
	}

	applyLog := &pb.GroupLog{
		Id:      logIdMember(lamptime, account.ID),
		GroupId: group.ID,
		PeerId:  []byte(account.ID),
		LogType: pb.GroupLog_MEMBER,
		Member: &pb.GroupMember{
			Id:     []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MemberOperate: pb.GroupLog_APPLY,
		Payload:       []byte(""),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if _, err = a.saveLog(ctx, applyLog); err != nil {
		return fmt.Errorf("save log error: %w", err)
	}

	if err = a.data.SetName(ctx, group.ID, group.Name); err != nil {
		return fmt.Errorf("ds set group name error: %w", err)
	}

	if err = a.data.SetAvatar(ctx, group.ID, group.Avatar); err != nil {
		return fmt.Errorf("ds set group avatar error: %w", err)
	}

	if err = a.data.SetState(ctx, group.ID, mytype.GroupStateNormal); err != nil {
		return fmt.Errorf("ds set group state error: %w", err)
	}

	if err = a.data.SetListID(ctx, group.ID); err != nil {
		return fmt.Errorf("ds set group session error: %w", err)
	}

	sessionID := mytype.GroupSessionID(group.ID)
	if err = a.sessionData.SetSessionID(ctx, sessionID.String()); err != nil {
		return fmt.Errorf("svc set session id error: %w", err)
	}

	if err = a.emitters.evtGroupsChange.Emit(myevent.EvtGroupsChange{
		AddGroups: []myevent.Groups{
			{
				GroupID:     group.ID,
				PeerIDs:     []peer.ID{account.ID},
				AcptPeerIDs: []peer.ID{account.ID},
			},
		},
	}); err != nil {
		return fmt.Errorf("emit groups change error: %w", err)
	}

	return nil
}

// ExitGroup 退出群
func (a *AdminProto) ExitGroup(ctx context.Context, account *mytype.Account, groupID string) error {

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.GroupLog{
		Id:      logIdExit(lamptime, account.ID),
		GroupId: groupID,
		PeerId:  []byte(account.ID),
		LogType: pb.GroupLog_MEMBER,
		Member: &pb.GroupMember{
			Id:     []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MemberOperate: pb.GroupLog_EXIT,
		Payload:       []byte{},
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if _, err = a.saveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.data.SetState(ctx, groupID, mytype.GroupStateExit); err != nil {
		return fmt.Errorf("set state exit error: %w", err)
	}

	if err = a.broadcastMessage(ctx, groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	// 触发退出群组
	if err = a.emitters.evtGroupsChange.Emit(myevent.EvtGroupsChange{
		DeleteGroups: []string{groupID},
	}); err != nil {
		return fmt.Errorf("emite add group error: %w", err)
	}

	return nil
}

// DeleteGroup 删除群
func (a *AdminProto) DeleteGroup(ctx context.Context, account *mytype.Account, groupID string) error {

	state, err := a.data.GetState(ctx, groupID)
	if err != nil {
		return fmt.Errorf("proto get state error: %w", err)
	}

	if state == mytype.GroupStateNormal {
		// 正常状态则先退出群
		lamptime, err := a.data.TickLamptime(ctx, groupID)
		if err != nil {
			return err
		}

		pbmsg := pb.GroupLog{
			Id:      logIdExit(lamptime, account.ID),
			GroupId: groupID,
			PeerId:  []byte(account.ID),
			LogType: pb.GroupLog_MEMBER,
			Member: &pb.GroupMember{
				Id:     []byte(account.ID),
				Name:   account.Name,
				Avatar: account.Avatar,
			},
			MemberOperate: pb.GroupLog_EXIT,
			Payload:       []byte{},
			CreateTime:    time.Now().Unix(),
			Lamportime:    lamptime,
			Signature:     []byte(""),
		}

		if _, err = a.saveLog(ctx, &pbmsg); err != nil {
			return fmt.Errorf("data save log error: %w", err)
		}

		if err = a.broadcastMessage(ctx, groupID, &pbmsg); err != nil {
			return fmt.Errorf("broadcast msg error: %w", err)
		}

		// 触发退出群组
		if err = a.emitters.evtGroupsChange.Emit(myevent.EvtGroupsChange{
			DeleteGroups: []string{groupID},
		}); err != nil {
			return fmt.Errorf("emite add group error: %w", err)
		}
	}

	if err = a.data.DeleteGroup(ctx, groupID); err != nil {
		return fmt.Errorf("data delete group error: %w", err)
	}

	return nil
}

// DisbandGroup 解散群
func (a *AdminProto) DisbandGroup(ctx context.Context, account *mytype.Account, groupID string) error {

	hostID := a.host.ID()

	creatorID, err := a.data.GetCreator(ctx, groupID)
	if err != nil {
		return fmt.Errorf("data.GetCreator error: %w", err)
	}

	if hostID != creatorID {
		return fmt.Errorf("must creator can disband group")
	}

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.GroupLog{
		Id:            logIdDisband(lamptime, account.ID),
		GroupId:       groupID,
		PeerId:        []byte(account.ID),
		LogType:       pb.GroupLog_DISBAND,
		Member:        nil,
		MemberOperate: pb.GroupLog_NONE,
		Payload:       []byte(groupID),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if _, err = a.saveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(ctx, groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	if err = a.emitters.evtGroupsChange.Emit(myevent.EvtGroupsChange{
		DeleteGroups: []string{groupID},
	}); err != nil {
		return fmt.Errorf("emite delete group error: %w", err)
	}

	return nil
}

// ReviewJoinGroup 入群审核
func (a *AdminProto) ReviewJoinGroup(ctx context.Context, groupID string, member *mytype.Peer, isAgree bool) error {
	hostID := a.host.ID()

	creator, err := a.data.GetCreator(ctx, groupID)
	if err != nil {
		return fmt.Errorf("data get creator error: %w", err)
	}

	if hostID != creator {
		return fmt.Errorf("no permission to review join group")
	}

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	operate := pb.GroupLog_AGREE
	if !isAgree {
		operate = pb.GroupLog_REJECTED
	}

	pbmsg := pb.GroupLog{
		Id:      logIdMember(lamptime, hostID),
		GroupId: groupID,
		PeerId:  []byte(hostID),
		LogType: pb.GroupLog_MEMBER,
		Member: &pb.GroupMember{
			Id:     []byte(member.ID),
			Name:   member.Name,
			Avatar: member.Avatar,
		},
		MemberOperate: operate,
		Payload:       []byte(""),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if _, err := a.saveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(ctx, groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	if err := a.emitGroupMemberChange(ctx, groupID); err != nil {
		return fmt.Errorf("emitGroupMemberChange error: %w", err)
	}

	return nil
}

// RemoveGroupMember 移除群成员
func (a *AdminProto) RemoveGroupMember(ctx context.Context, groupID string, removeMemberIDs []peer.ID) error {

	hostID := a.host.ID()

	creator, err := a.data.GetCreator(ctx, groupID)
	if err != nil {
		return fmt.Errorf("data get creator error: %w", err)
	}

	if hostID != creator {
		return fmt.Errorf("no permission to remove member")
	}

	members, err := a.data.GetMembers(ctx, groupID)
	if err != nil {
		return fmt.Errorf("ds get group members error: %w", err)
	}

	var removeMembers []*pb.GroupMember
	for _, member := range members {
		for _, memberID := range removeMemberIDs {
			if peer.ID(member.Id) == memberID {
				removeMembers = append(removeMembers, member)
				break
			}
		}
	}

	for _, member := range removeMembers {
		lamptime, err := a.data.TickLamptime(ctx, groupID)
		if err != nil {
			return err
		}

		pbmsg := pb.GroupLog{
			Id:      logIdMember(lamptime, hostID),
			GroupId: groupID,
			PeerId:  []byte(hostID),
			LogType: pb.GroupLog_MEMBER,
			Member: &pb.GroupMember{
				Id:     member.Id,
				Name:   member.Name,
				Avatar: member.Avatar,
			},
			MemberOperate: pb.GroupLog_REMOVE,
			Payload:       []byte(""),
			CreateTime:    time.Now().Unix(),
			Lamportime:    lamptime,
			Signature:     []byte(""),
		}

		if _, err := a.saveLog(ctx, &pbmsg); err != nil {
			return fmt.Errorf("data save log error: %w", err)
		}

		if err = a.broadcastMessage(ctx, groupID, &pbmsg); err != nil {
			return fmt.Errorf("broadcast msg error: %w", err)
		}
	}

	if err := a.emitGroupMemberChange(ctx, groupID); err != nil {
		return fmt.Errorf("emit group member change error: %w", err)
	}

	return nil
}

// GetGroupIDs 群IDs
func (a *AdminProto) GetGroupIDs(ctx context.Context) ([]string, error) {
	return a.data.GetGroupIDs(ctx)
}

// GetGroup 获取群
func (a *AdminProto) GetGroup(ctx context.Context, groupID string) (*mytype.Group, error) {
	name, err := a.data.GetName(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("data get name error: %w", err)
	}

	avatar, err := a.data.GetAvatar(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("data get avatar error: %w", err)
	}

	depositPeerID, err := a.data.GetDepositAddress(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("data get deposit peer error: %w", err)
	}

	return &mytype.Group{
		ID:             groupID,
		Name:           name,
		Avatar:         avatar,
		DepositAddress: depositPeerID,
	}, nil
}

// GetGroupDetail 群详情
func (a *AdminProto) GetGroupDetail(ctx context.Context, groupID string) (*mytype.GroupDetail, error) {
	creatorID, err := a.data.GetCreator(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("data get creator error: %w", err)
	}

	createTime, err := a.data.GetCreateTime(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("data get create time error: %w", err)
	}

	name, err := a.data.GetName(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("data get name error: %w", err)
	}

	avatar, err := a.data.GetAvatar(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("data get avatar error: %w", err)
	}

	notice, err := a.data.GetNotice(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("data get notice error: %w", err)
	}

	autoJoinGroup, err := a.data.GetAutoJoinGroup(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("data get auto join group error: %w", err)
	}

	depositPeerID, err := a.data.GetDepositAddress(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("data get deposit peer error: %w", err)
	}

	return &mytype.GroupDetail{
		ID:             groupID,
		CreatorID:      creatorID,
		Name:           name,
		Avatar:         avatar,
		Notice:         notice,
		AutoJoinGroup:  autoJoinGroup,
		DepositAddress: depositPeerID,
		CreateTime:     createTime,
	}, nil
}

// GetGroupMembers 获取群成员列表
func (a *AdminProto) GetGroupMembers(ctx context.Context, groupID string, keywords string, offset int, limit int) ([]mytype.GroupMember, error) {
	if offset < 0 || limit <= 0 {
		return nil, nil
	}

	members, err := a.data.GetMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get members error: %w", err)
	}

	var filters []*pb.GroupMember
	if keywords == "" {
		filters = members

	} else {
		for _, member := range members {
			if strings.Contains(member.Name, keywords) {
				filters = append(filters, member)
			}
		}
	}

	if len(filters) == 0 || offset >= len(filters) {
		return nil, nil
	}

	endOffset := offset + limit

	if endOffset > len(filters) {
		endOffset = len(filters)
	}

	var memberList []mytype.GroupMember
	for _, member := range filters[offset:endOffset] {
		memberList = append(memberList, mytype.GroupMember{
			ID:     peer.ID(member.Id),
			Name:   member.Name,
			Avatar: member.Avatar,
		})
	}

	return memberList, nil
}

// GetMemberIDs 群成员IDs
func (a *AdminProto) GetMemberIDs(ctx context.Context, groupID string) ([]peer.ID, error) {
	return a.data.GetMemberIDs(ctx, groupID)
}

// GetAgreePeerIDs 允许连接群的PeerIDs
func (a *AdminProto) GetAgreePeerIDs(ctx context.Context, groupID string) ([]peer.ID, error) {
	return a.data.GetAgreePeerIDs(ctx, groupID)
}

// SetGroupName 设置群名称
func (a *AdminProto) SetGroupName(ctx context.Context, groupID string, name string) error {

	hostID := a.host.ID()

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.GroupLog{
		Id:            logIdName(lamptime, hostID),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.GroupLog_NAME,
		Member:        nil,
		MemberOperate: pb.GroupLog_NONE,
		Payload:       []byte(name),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if _, err := a.saveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(ctx, groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	return nil
}

// SetGroupAlias 设置群昵称
func (a *AdminProto) SetGroupAlias(ctx context.Context, groupID string, alias string) error {
	return a.data.SetAlias(ctx, groupID, alias)
}

// SetGroupAvatar 设置群头像
func (a *AdminProto) SetGroupAvatar(ctx context.Context, groupID string, avatar string) error {

	hostID := a.host.ID()

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.GroupLog{
		Id:            logIdAvatar(lamptime, hostID),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.GroupLog_AVATAR,
		Member:        nil,
		MemberOperate: pb.GroupLog_NONE,
		Payload:       []byte(avatar),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if _, err := a.saveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(ctx, groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	return nil
}

// SetGroupAutoJoin 设置群自动加入
func (a *AdminProto) SetGroupAutoJoin(ctx context.Context, groupID string, isAutoJoin bool) error {

	hostID := a.host.ID()

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	autoJoin := "false"
	if isAutoJoin {
		autoJoin = "true"
	}

	pbmsg := pb.GroupLog{
		Id:            logIdAutojoin(lamptime, hostID),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.GroupLog_AUTO_JOIN_GROUP,
		Member:        nil,
		MemberOperate: pb.GroupLog_NONE,
		Payload:       []byte(autoJoin),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if _, err := a.saveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(ctx, groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	return nil
}

// SetGroupDepositAddress 设置群寄存地址
func (a *AdminProto) SetGroupDepositAddress(ctx context.Context, groupID string, depositPeerID peer.ID) error {

	hostID := a.host.ID()

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.GroupLog{
		Id:            logIdDepositPeer(lamptime, hostID),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.GroupLog_DEPOSIT_PEER_ID,
		Member:        nil,
		MemberOperate: pb.GroupLog_NONE,
		Payload:       []byte(depositPeerID),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if _, err := a.saveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(ctx, groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	return nil
}

// SetGroupNotice 设置群通知
func (a *AdminProto) SetGroupNotice(ctx context.Context, groupID string, notice string) error {
	hostID := a.host.ID()

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.GroupLog{
		Id:            logIdNotice(lamptime, hostID),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.GroupLog_NOTICE,
		Member:        nil,
		MemberOperate: pb.GroupLog_NONE,
		Payload:       []byte(notice),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if _, err := a.saveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(ctx, groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	return nil
}
