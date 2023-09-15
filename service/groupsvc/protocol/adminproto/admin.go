package adminproto

// 群管理相关协议

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/protocol"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/adminproto/ds"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/adminproto/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/libp2p/go-msgio/pbio"
)

var log = logging.Logger("message")

var StreamTimeout = 1 * time.Minute

const (
	ID      = protocol.GroupAdminID_v100
	SYNC_ID = protocol.GroupAdminSyncID_v100

	ServiceName = "group.admin"
	maxMsgSize  = 4 * 1024 // 4K
)

type AdminProto struct {
	host host.Host

	data ds.AdminIface

	emitters struct {
		evtSendAdminLog    event.Emitter
		evtInviteJoinGroup event.Emitter
	}

	groupConns map[string]map[peer.ID]struct{}
}

func NewAdminProto(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus) (*AdminProto, error) {
	var err error

	admsvc := &AdminProto{
		host:       lhost,
		data:       ds.AdminWrap(ids),
		groupConns: make(map[string]map[peer.ID]struct{}),
	}

	lhost.SetStreamHandler(ID, admsvc.Handler)
	lhost.SetStreamHandler(SYNC_ID, admsvc.Handler)

	if admsvc.emitters.evtSendAdminLog, err = eventBus.Emitter(&gevent.EvtSendAdminLog{}); err != nil {
		return nil, fmt.Errorf("set emitter error: %v", err)
	}

	if admsvc.emitters.evtInviteJoinGroup, err = eventBus.Emitter(&gevent.EvtInviteJoinGroup{}); err != nil {
		return nil, fmt.Errorf("set emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(gevent.EvtGroupConnectChange)}, eventbus.Name("adminlog"))
	if err != nil {
		return nil, fmt.Errorf("subscription failed. group admin server error: %v", err)

	} else {
		go admsvc.subscribeHandler(context.Background(), sub)
	}

	return admsvc, nil
}

func (a *AdminProto) Handler(s network.Stream) {
	fmt.Println("handler....")
	if err := s.Scope().SetService(ServiceName); err != nil {
		log.Errorf("failed to attaching stream to identify service: %v", err)
		s.Reset()
		return
	}
	defer s.Close()

	rd := pbio.NewDelimitedReader(s, maxMsgSize)
	defer rd.Close()

	s.SetDeadline(time.Now().Add(StreamTimeout))

	var msg pb.Log
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		s.Reset()
		return
	}

	s.SetReadDeadline(time.Time{})

	err := a.data.SaveLog(context.Background(), a.host.ID(), msg.GroupId, &msg)
	if err != nil {
		log.Errorf("log admin operation error: %v", err)
		s.Reset()
		return
	}
}

func (a *AdminProto) subscribeHandler(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}

			evt := e.(gevent.EvtGroupConnectChange)

			if !evt.IsConnected { // 断开连接
				delete(a.groupConns[evt.GroupID], evt.PeerID)
			}

			// 新建连接
			if _, exists := a.groupConns[evt.GroupID]; !exists {
				a.groupConns[evt.GroupID] = make(map[peer.ID]struct{})
			}
			a.groupConns[evt.GroupID][evt.PeerID] = struct{}{}

			// 启动同步
			a.sync(evt.GroupID, evt.PeerID)

		case <-ctx.Done():
			return
		}
	}
}

func (a *AdminProto) AgreeJoinGroup(ctx context.Context, account *types.Account, group *types.Group, lamptime uint64) error {

	if err := a.data.MergeLamportTime(ctx, group.ID, lamptime); err != nil {
		return fmt.Errorf("a.data.MergeLamportTime error: %w", err)
	}

	lamportTime, err := a.data.TickLamportTime(ctx, group.ID)
	if err != nil {
		return fmt.Errorf("a.data.TickLamportTime error: %w", err)
	}

	hostID := a.host.ID()

	memberLog := pb.Log{
		Id:      fmt.Sprintf("%d_%s_%s", lamportTime, "member", account.ID.String()),
		GroupId: group.ID,
		PeerId:  account.ID.String(),
		LogType: pb.Log_MEMBER,
		Member: &pb.Log_Member{
			Id:     []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MemberOperate: pb.Log_APPLY,
		Payload:       []byte(""),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamportTime,
		Signature:     []byte(""),
	}

	if err = a.data.JoinGroupSaveLog(ctx, hostID, group.ID, &memberLog); err != nil {
		return fmt.Errorf("a.data.JoinGroupSaveLog error: %w", err)
	}

	if err = a.data.JoinGroup(ctx, group); err != nil {
		return fmt.Errorf("a.data.JoinGroup error: %w", err)
	}

	return nil
}

func (a *AdminProto) CreateGroup(ctx context.Context, name string, avatar string, members []*pb.Log_Member) (string, error) {

	groupID := uuid.NewString()
	peerID := a.host.ID().String()

	// 创建组
	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return groupID, err
	}
	createLog := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "create", peerID),
		GroupId:    groupID,
		PeerId:     peerID,
		LogType:    pb.Log_CREATE,
		Payload:    []byte(groupID),
		CreateTime: time.Now().Unix(),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err = a.data.SaveLog(ctx, a.host.ID(), groupID, &createLog); err != nil {
		return groupID, err
	}

	if err := a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_CREATE,
		MsgData: &createLog,
	}); err != nil {
		return "", err
	}

	// 设置名称
	lamportTime, err = a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return groupID, err
	}
	nameLog := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "name", peerID),
		GroupId:    groupID,
		PeerId:     peerID,
		LogType:    pb.Log_NAME,
		Payload:    []byte(name),
		CreateTime: time.Now().Unix(),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}
	if err = a.data.SaveLog(ctx, a.host.ID(), groupID, &nameLog); err != nil {
		return groupID, err
	}
	if err := a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_NAME,
		MsgData: &nameLog,
	}); err != nil {
		return "", err
	}

	// 设置头像
	lamportTime, err = a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return groupID, err
	}
	avatarLog := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "avatar", peerID),
		GroupId:    groupID,
		PeerId:     peerID,
		LogType:    pb.Log_AVATAR,
		Payload:    []byte(avatar),
		CreateTime: time.Now().Unix(),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}
	if err = a.data.SaveLog(ctx, a.host.ID(), groupID, &avatarLog); err != nil {
		return groupID, err
	}
	if err := a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_AVATAR,
		MsgData: &nameLog,
	}); err != nil {
		return "", err
	}

	// 设置成员
	var memberIDs []peer.ID
	for _, member := range members {
		memberIDs = append(memberIDs, peer.ID(member.Id))

		lamportTime, err = a.data.TickLamportTime(ctx, groupID)
		if err != nil {
			return groupID, err
		}

		memberLog := pb.Log{
			Id:            fmt.Sprintf("%d_%s_%s", lamportTime, "member", peerID),
			GroupId:       groupID,
			PeerId:        peerID,
			LogType:       pb.Log_MEMBER,
			Member:        member,
			MemberOperate: pb.Log_AGREE,
			Payload:       []byte(""),
			CreateTime:    time.Now().Unix(),
			Lamportime:    lamportTime,
			Signature:     []byte(""),
		}

		if err = a.data.SaveLog(ctx, a.host.ID(), groupID, &memberLog); err != nil {
			return groupID, err
		}

		if err := a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
			MsgType: pb.Log_MEMBER,
			MsgData: &memberLog,
		}); err != nil {
			return "", err
		}

	}

	// 发送邀请消息
	if err = a.emitters.evtInviteJoinGroup.Emit(gevent.EvtInviteJoinGroup{
		PeerIDs:       memberIDs,
		GroupID:       groupID,
		GroupName:     name,
		GroupAvatar:   avatar,
		GroupLamptime: lamportTime,
	}); err != nil {
		return groupID, fmt.Errorf("evtInviteJoinGroup.Emit error: %w", err)
	}

	return groupID, nil
}

// ExitGroup 退出群
func (a *AdminProto) ExitGroup(ctx context.Context, account *types.Account, groupID string) error {

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:      fmt.Sprintf("%d_%s_%s", lamportTime, "exit", account.ID.String()),
		GroupId: groupID,
		PeerId:  account.ID.String(),
		LogType: pb.Log_MEMBER,
		Member: &pb.Log_Member{
			Id:     []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MemberOperate: pb.Log_EXIT,
		Payload:       []byte{},
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamportTime,
		Signature:     []byte(""),
	}

	if err = a.data.SaveLog(ctx, account.ID, groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_MEMBER,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteGroup 删除群
func (a *AdminProto) DeleteGroup(ctx context.Context, account *types.Account, groupID string) error {

	state, err := a.data.GetState(ctx, groupID)
	if err != nil {
		return fmt.Errorf("proto get state error: %w", err)
	}

	if state == types.GroupStateNormal {
		// 正常状态则先退出群
		lamportTime, err := a.data.TickLamportTime(ctx, groupID)
		if err != nil {
			return err
		}

		pbmsg := pb.Log{
			Id:      fmt.Sprintf("%d_%s_%s", lamportTime, "exit", account.ID.String()),
			GroupId: groupID,
			PeerId:  account.ID.String(),
			LogType: pb.Log_MEMBER,
			Member: &pb.Log_Member{
				Id:     []byte(account.ID),
				Name:   account.Name,
				Avatar: account.Avatar,
			},
			MemberOperate: pb.Log_EXIT,
			Payload:       []byte{},
			CreateTime:    time.Now().Unix(),
			Lamportime:    lamportTime,
			Signature:     []byte(""),
		}

		if err = a.data.SaveLog(ctx, account.ID, groupID, &pbmsg); err != nil {
			return err
		}

		err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
			MsgType: pb.Log_MEMBER,
			MsgData: &pbmsg,
		})
		if err != nil {
			return err
		}
	}

	if err = a.data.DeleteGroup(ctx, groupID); err != nil {
		return fmt.Errorf("data.DeleteGroup error: %w", err)
	}

	return nil
}

// DisbandGroup 解散群（所有人退出）
func (a *AdminProto) DisbandGroup(ctx context.Context, account *types.Account, groupID string) error {

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, types.GroupStateDisband, account.ID.String()),
		GroupId:    groupID,
		PeerId:     account.ID.String(),
		LogType:    pb.Log_DISBAND,
		Payload:    []byte(groupID),
		CreateTime: time.Now().Unix(),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err = a.data.SaveLog(ctx, account.ID, groupID, &pbmsg); err != nil {
		return err
	}

	if err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_DISBAND,
		MsgData: &pbmsg,
	}); err != nil {
		return err
	}

	return nil
}

func (a *AdminProto) GetGroups(ctx context.Context) ([]ds.Group, error) {
	return a.data.GetGroups(ctx)
}

func (a *AdminProto) GetGroupIDs(ctx context.Context) ([]string, error) {
	return a.data.GetGroupIDs(ctx)
}

func (a *AdminProto) GetGroup(ctx context.Context, groupID string) (*ds.Group, error) {
	return a.data.GetGroup(ctx, groupID)
}

// todo: ??
func (a *AdminProto) GetGroupDetail(ctx context.Context, groupID string) (*ds.Group, error) {
	return a.data.GetGroup(ctx, groupID)
}

func (a *AdminProto) GroupName(ctx context.Context, groupID string) (string, error) {

	if name, err := a.data.GroupLocalName(ctx, groupID); err != nil {
		return "", err

	} else if name != "" {
		return name, nil
	}

	return a.data.GroupName(ctx, groupID)
}

func (a *AdminProto) GroupAvatar(ctx context.Context, groupID string) (string, error) {

	if avatar, err := a.data.GroupLocalAvatar(ctx, groupID); err != nil {
		return "", err

	} else if avatar != "" {
		return avatar, nil
	}

	return a.data.GroupAvatar(ctx, groupID)
}

func (a *AdminProto) GroupNotice(ctx context.Context, groupID string) (string, error) {
	return a.data.GroupNotice(ctx, groupID)
}

func (a *AdminProto) SetGroupName(ctx context.Context, groupID string, name string) error {

	peerID := a.host.ID().String()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "name", peerID),
		GroupId:    groupID,
		PeerId:     peerID,
		LogType:    pb.Log_NAME,
		Payload:    []byte(name),
		CreateTime: time.Now().Unix(),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err := a.data.SaveLog(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_NAME,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AdminProto) SetGroupLocalName(ctx context.Context, groupID string, name string) error {
	err := a.data.SetGroupLocalName(ctx, groupID, name)
	if err != nil {
		return fmt.Errorf("data.SetGroupLocalName error: %w", err)
	}

	return nil
}

func (a *AdminProto) SetGroupAvatar(ctx context.Context, groupID string, avatar string) error {

	peerID := a.host.ID().String()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "avatar", peerID),
		GroupId:    groupID,
		PeerId:     peerID,
		LogType:    pb.Log_AVATAR,
		Payload:    []byte(avatar),
		CreateTime: time.Now().Unix(),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err := a.data.SaveLog(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_AVATAR,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AdminProto) SetGroupLocalAvatar(ctx context.Context, groupID string, avatar string) error {
	err := a.data.SetGroupLocalAvatar(ctx, groupID, avatar)
	if err != nil {
		return fmt.Errorf("data.SetGroupLocalAvatar error: %w", err)
	}

	return nil
}

func (a *AdminProto) SetGroupAutoJoin(ctx context.Context, groupID string, isAutoJoin bool) error {
	return nil
}

func (a *AdminProto) SetGroupNotice(ctx context.Context, groupID string, notice string) error {
	peerID := a.host.ID().String()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "notice", peerID),
		GroupId:    groupID,
		PeerId:     peerID,
		LogType:    pb.Log_NOTICE,
		Payload:    []byte(notice),
		CreateTime: time.Now().Unix(),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err := a.data.SaveLog(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_NOTICE,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AdminProto) ApplyMember(ctx context.Context, groupID string, memberID peer.ID, lamportime uint64) error {

	peerID := a.host.ID().String()

	stream, err := a.host.NewStream(ctx, memberID, ID)
	if err != nil {
		return err
	}

	pw := pbio.NewDelimitedWriter(stream)

	pmsg := pb.Log{
		Id:            fmt.Sprintf("%d_%s_%s", lamportime, "member", peerID),
		GroupId:       groupID,
		PeerId:        peerID,
		LogType:       pb.Log_MEMBER,
		MemberOperate: pb.Log_APPLY,
		Member: &pb.Log_Member{
			Id:     []byte(peerID),
			Name:   "",
			Avatar: "",
		},
		Payload:    []byte(groupID),
		CreateTime: time.Now().Unix(),
		Lamportime: lamportime,
		Signature:  []byte{},
	}

	if err := pw.WriteMsg(&pmsg); err != nil {
		return err
	}

	return nil
}

func (a *AdminProto) ReviewJoinGroup(ctx context.Context, groupID string, member *types.Peer, isAgree bool) error {
	peerID := a.host.ID().String()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	operate := pb.Log_AGREE
	if !isAgree {
		operate = pb.Log_REJECTED
	}

	pbmsg := pb.Log{
		Id:      fmt.Sprintf("%d_%s_%s", lamportTime, "member", peerID),
		GroupId: groupID,
		PeerId:  peerID,
		LogType: pb.Log_MEMBER,
		Member: &pb.Log_Member{
			Id:     []byte(member.ID),
			Name:   member.Name,
			Avatar: member.Avatar,
		},
		MemberOperate: operate,
		Payload:       []byte(""),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamportTime,
		Signature:     []byte(""),
	}

	if err := a.data.SaveLog(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_MEMBER,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AdminProto) RemoveMember(ctx context.Context, groupID string, memberID peer.ID) error {
	hostID := a.host.ID()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	// todo: get member
	member := types.GroupMember{
		ID:     memberID,
		Name:   "",
		Avatar: "",
	}

	pbmsg := pb.Log{
		Id:      fmt.Sprintf("%d_%s_%s", lamportTime, "member", hostID.String()),
		GroupId: groupID,
		PeerId:  hostID.String(),
		LogType: pb.Log_MEMBER,
		Member: &pb.Log_Member{
			Id:     []byte(member.ID),
			Name:   member.Name,
			Avatar: member.Avatar,
		},
		MemberOperate: pb.Log_REMOVE,
		Payload:       []byte(""),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamportTime,
		Signature:     []byte(""),
	}

	if err := a.data.SaveLog(ctx, hostID, groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_MEMBER,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AdminProto) GetMemberIDs(ctx context.Context, groupID string) ([]peer.ID, error) {
	members, err := a.getGroupMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("ds get group members error: %w", err)
	}

	memberIDs := make([]peer.ID, len(members))

	for i, member := range members {
		memberIDs[i] = peer.ID(member.Id)
	}

	return memberIDs, nil
}

func (a *AdminProto) GetGroupMembers(ctx context.Context, groupID string) ([]*pb.Log_Member, error) {
	return a.getGroupMembers(ctx, groupID)
}

func (a *AdminProto) getGroupMembers(ctx context.Context, groupID string) ([]*pb.Log_Member, error) {

	memberLogs, err := a.data.GroupMemberLogs(ctx, groupID)
	if err != nil {
		return nil, err
	}

	oks := make(map[string]*pb.Log_Member)
	mmap := make(map[string]pb.Log_MemberOperate)

	for _, pbmsg := range memberLogs {
		if state, exists := mmap[string(pbmsg.Member.Id)]; !exists {
			mmap[string(pbmsg.Member.Id)] = pbmsg.MemberOperate

		} else {
			if pbmsg.MemberOperate == state {
				continue
			}

			switch state {
			case pb.Log_REMOVE, pb.Log_REJECTED:
				continue
			case pb.Log_AGREE:
				if pbmsg.MemberOperate == pb.Log_APPLY {
					oks[string(pbmsg.Member.Id)] = pbmsg.Member
				}
			case pb.Log_APPLY:
				if pbmsg.MemberOperate == pb.Log_AGREE {
					oks[string(pbmsg.Member.Id)] = pbmsg.Member
				}
			default:
				mmap[string(pbmsg.Member.Id)] = pbmsg.MemberOperate
			}
		}
	}

	var members []*pb.Log_Member
	for _, member := range oks {
		members = append(members, member)
	}

	return members, nil
}
