package admin

// 群管理相关协议

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	gevent "github.com/jianbo-zh/dchat/service/event"
	"github.com/jianbo-zh/dchat/service/group/datastore"
	"github.com/jianbo-zh/dchat/service/group/protocol/admin/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/libp2p/go-msgio/pbio"
)

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/admin.proto=./pb pb/admin.proto

var log = logging.Logger("message")

var StreamTimeout = 1 * time.Minute

const (
	ID = "/dchat/group/admin/1.0.0"

	ServiceName = "group.admin"
	maxMsgSize  = 4 * 1024 // 4K
)

type GroupAdminService struct {
	host host.Host

	datastore datastore.GroupIface

	emitters struct {
		evtSendAdminLog event.Emitter
	}
}

func NewGroupAdminService(h host.Host, ds datastore.GroupIface, eventBus event.Bus) *GroupAdminService {
	gsvc := &GroupAdminService{
		host:      h,
		datastore: ds,
	}

	h.SetStreamHandler(ID, gsvc.Handler)

	var err error
	if gsvc.emitters.evtSendAdminLog, err = eventBus.Emitter(&gevent.EvtSendAdminLog{}); err != nil {
		log.Errorf("set emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(gevent.EvtSendAdminLog), new(gevent.EvtRecvAdminLog)}, eventbus.Name("adminlog"))
	if err != nil {
		log.Warnf("subscription failed. group admin server error: %v", err)

	} else {
		go gsvc.background(context.Background(), sub)
	}

	return gsvc
}

func (grp *GroupAdminService) background(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}
			ev := e.(gevent.EvtRecvAdminLog)
			// 接收消息日志
			err := grp.datastore.LogAdminOperation(ctx, grp.host.ID(), datastore.GroupID(ev.MsgData.GroupId), ev.MsgData)
			if err != nil {
				log.Errorf("log admin operation error: %v", err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (group *GroupAdminService) Handler(s network.Stream) {
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

	var msg pb.AdminLog
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		s.Reset()
		return
	}

	s.SetReadDeadline(time.Time{})

	err := group.datastore.LogAdminOperation(context.Background(), group.host.ID(), datastore.GroupID(msg.GroupId), &msg)
	if err != nil {
		log.Errorf("log admin operation error: %v", err)
		s.Reset()
		return
	}
}

func (group *GroupAdminService) CreateGroup(ctx context.Context, name string, memberIDs []peer.ID) (string, error) {

	groupID := datastore.GroupID(uuid.NewString())
	peerID := group.host.ID().String()

	// 创建组
	lamportTime, err := group.datastore.TickLamportTime(ctx, groupID)
	if err != nil {
		return string(groupID), err
	}
	createLog := pb.AdminLog{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "create", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.AdminLog_CREATE,
		Payload:    []byte(groupID),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err = group.datastore.LogAdminOperation(ctx, group.host.ID(), groupID, &createLog); err != nil {
		return string(groupID), err
	}

	if err := group.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_CREATE,
		MsgData: &createLog,
	}); err != nil {
		return "", err
	}

	// 设置名称
	lamportTime, err = group.datastore.TickLamportTime(ctx, groupID)
	if err != nil {
		return string(groupID), err
	}
	nameLog := pb.AdminLog{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "name", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.AdminLog_NAME,
		Payload:    []byte(name),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}
	if err = group.datastore.LogAdminOperation(ctx, group.host.ID(), groupID, &nameLog); err != nil {
		return string(groupID), err
	}

	if err := group.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_NAME,
		MsgData: &nameLog,
	}); err != nil {
		return "", err
	}

	// 设置成员
	for _, memberID := range memberIDs {
		lamportTime, err = group.datastore.TickLamportTime(ctx, groupID)
		if err != nil {
			return string(groupID), err
		}

		memberLog := pb.AdminLog{
			Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "member", peerID),
			GroupId:    string(groupID),
			PeerId:     peerID,
			Type:       pb.AdminLog_MEMBER,
			MemberId:   memberID.String(),
			Operate:    pb.AdminLog_AGREE,
			Payload:    []byte(""),
			Timestamp:  int32(time.Now().Unix()),
			Lamportime: lamportTime,
			Signature:  []byte(""),
		}

		if err = group.datastore.LogAdminOperation(ctx, group.host.ID(), groupID, &memberLog); err != nil {
			return string(groupID), err
		}

		// todo: 要单独给人发消息
	}

	return string(groupID), nil
}

func (group *GroupAdminService) DisbandGroup(ctx context.Context, groupID0 string) error {

	groupID := datastore.GroupID(groupID0)
	peerID := group.host.ID().String()

	// 创建组
	lamportTime, err := group.datastore.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.AdminLog{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "disband", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.AdminLog_DISBAND,
		Payload:    []byte(groupID),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err = group.datastore.LogAdminOperation(ctx, group.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = group.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_DISBAND,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (group *GroupAdminService) ListGroups(ctx context.Context) ([]datastore.Group, error) {
	return group.datastore.ListGroups(ctx)
}

func (group *GroupAdminService) GroupName(ctx context.Context, groupID string) (string, error) {

	if remark, err := group.datastore.GroupRemark(ctx, datastore.GroupID(groupID)); err != nil {
		return "", err

	} else if remark != "" {
		return remark, nil
	}

	return group.datastore.GroupName(ctx, datastore.GroupID(groupID))
}

func (group *GroupAdminService) GroupNotice(ctx context.Context, groupID string) (string, error) {
	return group.datastore.GroupNotice(ctx, datastore.GroupID(groupID))
}

func (group *GroupAdminService) SetGroupName(ctx context.Context, groupID0 string, name string) error {

	groupID := datastore.GroupID(groupID0)
	peerID := group.host.ID().String()

	lamportTime, err := group.datastore.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.AdminLog{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "name", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.AdminLog_NAME,
		Payload:    []byte(name),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err := group.datastore.LogAdminOperation(ctx, group.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = group.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_NAME,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (group *GroupAdminService) SetGroupNotice(ctx context.Context, groupID0 string, notice string) error {

	groupID := datastore.GroupID(groupID0)
	peerID := group.host.ID().String()

	lamportTime, err := group.datastore.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.AdminLog{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "notice", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.AdminLog_NOTICE,
		Payload:    []byte(notice),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err := group.datastore.LogAdminOperation(ctx, group.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = group.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_NOTICE,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (group *GroupAdminService) SetGroupRemark(ctx context.Context, groupID string, remark string) error {
	// 只是更新本地
	return group.datastore.SetGroupRemark(ctx, datastore.GroupID(groupID), remark)
}

func (group *GroupAdminService) InviteMember(ctx context.Context, groupID0 string, peerID0 peer.ID) error {

	groupID := datastore.GroupID(groupID0)
	peerID := group.host.ID().String()

	lamportTime, err := group.datastore.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	return group.datastore.LogAdminOperation(ctx, group.host.ID(), groupID, &pb.AdminLog{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "member", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.AdminLog_MEMBER,
		Operate:    pb.AdminLog_INVITE,
		MemberId:   peerID0.String(),
		Payload:    []byte(""),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	})
}

func (group *GroupAdminService) ApplyMember(ctx context.Context, groupID string, memberID peer.ID, lamportime uint64) error {

	peerID := group.host.ID().String()

	stream, err := group.host.NewStream(ctx, memberID, ID)
	if err != nil {
		return err
	}

	pw := pbio.NewDelimitedWriter(stream)

	pmsg := pb.AdminLog{
		Id:         fmt.Sprintf("%d_%s_%s", lamportime, "member", peerID),
		GroupId:    groupID,
		PeerId:     peerID,
		Type:       pb.AdminLog_MEMBER,
		Operate:    pb.AdminLog_APPLY,
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

func (group *GroupAdminService) ReviewMember(ctx context.Context, groupID0 string, memberID0 peer.ID, isAgree bool) error {
	groupID := datastore.GroupID(groupID0)
	peerID := group.host.ID().String()

	lamportTime, err := group.datastore.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	operate := pb.AdminLog_AGREE
	if !isAgree {
		operate = pb.AdminLog_REJECTED
	}

	pbmsg := pb.AdminLog{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "member", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.AdminLog_MEMBER,
		Operate:    operate,
		MemberId:   memberID0.String(),
		Payload:    []byte(""),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err := group.datastore.LogAdminOperation(ctx, group.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = group.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_MEMBER,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (group *GroupAdminService) RemoveMember(ctx context.Context, groupID0 string, memberID0 peer.ID) error {
	groupID := datastore.GroupID(groupID0)
	peerID := group.host.ID().String()

	lamportTime, err := group.datastore.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.AdminLog{
		Id:         fmt.Sprintf("%d_%s_%s", lamportTime, "member", peerID),
		GroupId:    string(groupID),
		PeerId:     peerID,
		Type:       pb.AdminLog_MEMBER,
		Operate:    pb.AdminLog_REMOVE,
		MemberId:   memberID0.String(),
		Payload:    []byte(""),
		Timestamp:  int32(time.Now().Unix()),
		Lamportime: lamportTime,
		Signature:  []byte(""),
	}

	if err := group.datastore.LogAdminOperation(ctx, group.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = group.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_MEMBER,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (group *GroupAdminService) ListMembers(ctx context.Context, groupID0 string) ([]Member, error) {

	groupID := datastore.GroupID(groupID0)
	memberLogs, err := group.datastore.GroupMemberLogs(ctx, groupID)
	if err != nil {
		return nil, err
	}

	oks := make(map[string]struct{})
	mmap := make(map[string]pb.AdminLog_Operate)

	for _, pbmsg := range memberLogs {
		if state, exists := mmap[pbmsg.MemberId]; !exists {
			mmap[pbmsg.MemberId] = pbmsg.Operate

		} else {
			if pbmsg.Operate == state {
				continue
			}

			switch state {
			case pb.AdminLog_REMOVE, pb.AdminLog_REJECTED:
				continue
			case pb.AdminLog_AGREE:
				if pbmsg.Operate == pb.AdminLog_APPLY {
					oks[pbmsg.MemberId] = struct{}{}
				}
			case pb.AdminLog_APPLY:
				if pbmsg.Operate == pb.AdminLog_AGREE {
					oks[pbmsg.MemberId] = struct{}{}
				}
			default:
				mmap[pbmsg.MemberId] = pbmsg.Operate
			}
		}
	}

	var members []Member
	for memberID := range oks {
		peerID, err := peer.Decode(memberID)
		if err != nil {
			return nil, err
		}
		members = append(members, Member{
			PeerID: peerID,
		})
	}

	return members, nil
}
