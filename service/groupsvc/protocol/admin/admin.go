package admin

// 群管理相关协议

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/admin/ds"
	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/admin/pb"
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
	ID      = "/dchat/group/admin/1.0.0"
	SYNC_ID = "/dchat/group/syncmsg/1.0.0"

	ServiceName = "group.admin"
	maxMsgSize  = 4 * 1024 // 4K
)

type AdminService struct {
	host host.Host

	data ds.AdminIface

	emitters struct {
		evtSendAdminLog    event.Emitter
		evtInviteJoinGroup event.Emitter
	}

	groupConns map[string]map[peer.ID]struct{}
}

func NewAdminService(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus) (*AdminService, error) {
	var err error

	admsvc := &AdminService{
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

func (a *AdminService) Handler(s network.Stream) {
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

	err := a.data.SaveLog(context.Background(), a.host.ID(), ds.GroupID(msg.GroupId), &msg)
	if err != nil {
		log.Errorf("log admin operation error: %v", err)
		s.Reset()
		return
	}
}

func (a *AdminService) subscribeHandler(ctx context.Context, sub event.Subscription) {
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

func (a *AdminService) JoinGroup(ctx context.Context, groupID string, groupName string, groupAvatar string, lamptime uint64) error {
	err := a.data.MergeLamportTime(ctx, groupID, lamptime)
	if err != nil {
		return fmt.Errorf("a.data.MergeLamportTime error: %w", err)
	}

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return fmt.Errorf("a.data.TickLamportTime error: %w", err)
	}

	hostID := a.host.ID()

	memberLog := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "member", hostID.String()),
		GroupId:    groupID,
		PeerId:     hostID.String(),
		Type:       pb.Log_MEMBER,
		MemberId:   hostID.String(),
		Operate:    pb.Log_APPLY,
		Payload:    []byte(""),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err = a.data.JoinGroupSaveLog(ctx, hostID, groupID, &memberLog); err != nil {
		return fmt.Errorf("a.data.JoinGroupSaveLog error: %w", err)
	}

	if err = a.data.JoinGroup(ctx, groupID, groupName, groupAvatar); err != nil {
		return fmt.Errorf("a.data.JoinGroup error: %w", err)
	}

	return nil
}

func (a *AdminService) CreateGroup(ctx context.Context, name string, avatar string, memberIDs []peer.ID) (string, error) {

	groupID := ds.GroupID(uuid.NewString())
	peerID := a.host.ID().String()

	// 创建组
	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return string(groupID), err
	}
	createLog := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "create", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_CREATE,
		Payload:    []byte(groupID),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err = a.data.SaveLog(ctx, a.host.ID(), groupID, &createLog); err != nil {
		return string(groupID), err
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
		return string(groupID), err
	}
	nameLog := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "name", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_NAME,
		Payload:    []byte(name),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}
	if err = a.data.SaveLog(ctx, a.host.ID(), groupID, &nameLog); err != nil {
		return string(groupID), err
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
		return string(groupID), err
	}
	avatarLog := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "avatar", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_AVATAR,
		Payload:    []byte(avatar),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}
	if err = a.data.SaveLog(ctx, a.host.ID(), groupID, &avatarLog); err != nil {
		return string(groupID), err
	}
	if err := a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_AVATAR,
		MsgData: &nameLog,
	}); err != nil {
		return "", err
	}

	// 设置成员
	for _, memberID := range memberIDs {
		lamportTime, err = a.data.TickLamportTime(ctx, groupID)
		if err != nil {
			return string(groupID), err
		}

		memberLog := pb.Log{
			Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "member", peerID),
			GroupId:    string(groupID),
			PeerId:     peerID,
			Type:       pb.Log_MEMBER,
			MemberId:   memberID.String(),
			Operate:    pb.Log_AGREE,
			Payload:    []byte(""),
			Timestamp:  int32(time.Now().Unix()),
			Lamportime: lamportTime,
			Signature:  []byte(""),
		}

		if err = a.data.SaveLog(ctx, a.host.ID(), groupID, &memberLog); err != nil {
			return string(groupID), err
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
		return string(groupID), fmt.Errorf("evtInviteJoinGroup.Emit error: %w", err)
	}

	return string(groupID), nil
}

func (a *AdminService) DisbandGroup(ctx context.Context, groupID0 string) error {

	groupID := ds.GroupID(groupID0)
	peerID := a.host.ID().String()

	// 创建组
	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "disband", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_DISBAND,
		Payload:    []byte(groupID),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err = a.data.SaveLog(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_DISBAND,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AdminService) ExitGroup(ctx context.Context, groupID string) error {

	peerID := a.host.ID().String()

	// 创建组
	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "exit", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_DISBAND,
		Payload:    []byte(groupID),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err = a.data.SaveLog(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_DISBAND,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AdminService) DeleteGroup(ctx context.Context, groupID string) error {

	fmt.Println("delete group ", groupID)

	peerID := a.host.ID().String()

	// 创建组
	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "exit", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_DISBAND,
		Payload:    []byte(groupID),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err = a.data.SaveLog(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_DISBAND,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	if err = a.data.DeleteGroup(ctx, groupID); err != nil {
		return fmt.Errorf("data.DeleteGroup error: %w", err)
	}

	return nil
}

func (a *AdminService) ListGroups(ctx context.Context) ([]ds.Group, error) {
	return a.data.ListGroups(ctx)
}

func (a *AdminService) GetGroupIDs(ctx context.Context) ([]string, error) {
	return a.data.GetGroupIDs(ctx)
}

func (a *AdminService) GetGroup(ctx context.Context, groupID string) (*ds.Group, error) {
	return a.data.GetGroup(ctx, groupID)
}

func (a *AdminService) GroupName(ctx context.Context, groupID string) (string, error) {

	if name, err := a.data.GroupLocalName(ctx, ds.GroupID(groupID)); err != nil {
		return "", err

	} else if name != "" {
		return name, nil
	}

	return a.data.GroupName(ctx, ds.GroupID(groupID))
}

func (a *AdminService) GroupAvatar(ctx context.Context, groupID string) (string, error) {

	if avatar, err := a.data.GroupLocalAvatar(ctx, ds.GroupID(groupID)); err != nil {
		return "", err

	} else if avatar != "" {
		return avatar, nil
	}

	return a.data.GroupAvatar(ctx, ds.GroupID(groupID))
}

func (a *AdminService) GroupNotice(ctx context.Context, groupID string) (string, error) {
	return a.data.GroupNotice(ctx, ds.GroupID(groupID))
}

func (a *AdminService) SetGroupName(ctx context.Context, groupID0 string, name string) error {

	groupID := ds.GroupID(groupID0)
	peerID := a.host.ID().String()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "name", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_NAME,
		Payload:    []byte(name),
		Timestamp:  int32(time.Now().Unix()),
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

func (a *AdminService) SetGroupLocalName(ctx context.Context, groupID0 string, name string) error {
	err := a.data.SetGroupLocalName(ctx, groupID0, name)
	if err != nil {
		return fmt.Errorf("data.SetGroupLocalName error: %w", err)
	}

	return nil
}

func (a *AdminService) SetGroupAvatar(ctx context.Context, groupID0 string, avatar string) error {

	groupID := ds.GroupID(groupID0)
	peerID := a.host.ID().String()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "avatar", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_AVATAR,
		Payload:    []byte(avatar),
		Timestamp:  int32(time.Now().Unix()),
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

func (a *AdminService) SetGroupLocalAvatar(ctx context.Context, groupID0 string, avatar string) error {
	err := a.data.SetGroupLocalAvatar(ctx, groupID0, avatar)
	if err != nil {
		return fmt.Errorf("data.SetGroupLocalAvatar error: %w", err)
	}

	return nil
}

func (a *AdminService) SetGroupNotice(ctx context.Context, groupID0 string, notice string) error {

	groupID := ds.GroupID(groupID0)
	peerID := a.host.ID().String()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "notice", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_NOTICE,
		Payload:    []byte(notice),
		Timestamp:  int32(time.Now().Unix()),
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

func (a *AdminService) InviteMember(ctx context.Context, groupID0 string, peerID0 peer.ID) error {

	groupID := ds.GroupID(groupID0)
	peerID := a.host.ID().String()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	return a.data.SaveLog(ctx, a.host.ID(), groupID, &pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "member", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_MEMBER,
		Operate:    pb.Log_INVITE,
		MemberId:   peerID0.String(),
		Payload:    []byte(""),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	})
}

func (a *AdminService) ApplyMember(ctx context.Context, groupID string, memberID peer.ID, lamportime uint64) error {

	peerID := a.host.ID().String()

	stream, err := a.host.NewStream(ctx, memberID, ID)
	if err != nil {
		return err
	}

	pw := pbio.NewDelimitedWriter(stream)

	pmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportime, "member", peerID),
		GroupId:    groupID,
		PeerId:     peerID,
		Type:       pb.Log_MEMBER,
		Operate:    pb.Log_APPLY,
		MemberId:   peerID,
		Payload:    []byte(groupID),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportime,
		Signature:  []byte{},
	}

	if err := pw.WriteMsg(&pmsg); err != nil {
		return err
	}

	return nil
}

func (a *AdminService) ReviewMember(ctx context.Context, groupID0 string, memberID0 peer.ID, isAgree bool) error {
	groupID := ds.GroupID(groupID0)
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
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "member", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_MEMBER,
		Operate:    operate,
		MemberId:   memberID0.String(),
		Payload:    []byte(""),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
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

func (a *AdminService) RemoveMember(ctx context.Context, groupID0 string, memberID0 peer.ID) error {
	groupID := ds.GroupID(groupID0)
	peerID := a.host.ID().String()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "member", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.Log_MEMBER,
		Operate:    pb.Log_REMOVE,
		MemberId:   memberID0.String(),
		Payload:    []byte(""),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
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

func (a *AdminService) ListMembers(ctx context.Context, groupID0 string) ([]peer.ID, error) {

	groupID := ds.GroupID(groupID0)
	memberLogs, err := a.data.GroupMemberLogs(ctx, groupID)
	if err != nil {
		return nil, err
	}

	oks := make(map[string]struct{})
	mmap := make(map[string]pb.Log_Operate)

	for _, pbmsg := range memberLogs {
		if state, exists := mmap[pbmsg.MemberId]; !exists {
			mmap[pbmsg.MemberId] = pbmsg.Operate

		} else {
			if pbmsg.Operate == state {
				continue
			}

			switch state {
			case pb.Log_REMOVE, pb.Log_REJECTED:
				continue
			case pb.Log_AGREE:
				if pbmsg.Operate == pb.Log_APPLY {
					oks[pbmsg.MemberId] = struct{}{}
				}
			case pb.Log_APPLY:
				if pbmsg.Operate == pb.Log_AGREE {
					oks[pbmsg.MemberId] = struct{}{}
				}
			default:
				mmap[pbmsg.MemberId] = pbmsg.Operate
			}
		}
	}

	var memberIDs []peer.ID
	for memberID := range oks {
		peerID, err := peer.Decode(memberID)
		if err != nil {
			return nil, err
		}
		memberIDs = append(memberIDs, peerID)
	}

	return memberIDs, nil
}
