package admin

// 群管理相关协议

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/service/group/protocol/admin/ds"
	"github.com/jianbo-zh/dchat/service/group/protocol/admin/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/libp2p/go-msgio/pbio"
)

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/group_admin.proto=./pb pb/group_admin.proto

var log = logging.Logger("message")

var StreamTimeout = 1 * time.Minute

const (
	ID = "/dchat/group/admin/1.0.0"

	ServiceName = "group.admin"
	maxMsgSize  = 4 * 1024 // 4K
)

type AdminService struct {
	host host.Host

	data ds.AdminIface

	emitters struct {
		evtSendAdminLog event.Emitter
	}
}

func NewAdminService(h host.Host, ids ipfsds.Batching, eventBus event.Bus) (*AdminService, error) {
	var err error

	admsvc := &AdminService{
		host: h,
		data: ds.AdminWrap(ids),
	}

	h.SetStreamHandler(ID, admsvc.Handler)

	if admsvc.emitters.evtSendAdminLog, err = eventBus.Emitter(&gevent.EvtSendAdminLog{}); err != nil {
		return nil, fmt.Errorf("set emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(gevent.EvtRecvAdminLog)}, eventbus.Name("adminlog"))
	if err != nil {
		return nil, fmt.Errorf("subscription failed. group admin server error: %v", err)

	} else {
		go admsvc.handleSubscribe(context.Background(), sub)
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

	var msg pb.AdminLog
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %w", err)
		s.Reset()
		return
	}

	s.SetReadDeadline(time.Time{})

	err := a.data.LogAdminOperation(context.Background(), a.host.ID(), ds.GroupID(msg.GroupId), &msg)
	if err != nil {
		log.Errorf("log admin operation error: %v", err)
		s.Reset()
		return
	}
}

func (a *AdminService) handleSubscribe(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}
			ev := e.(gevent.EvtRecvAdminLog)
			// 接收消息日志
			err := a.data.LogAdminOperation(ctx, a.host.ID(), ds.GroupID(ev.MsgData.GroupId), ev.MsgData)
			if err != nil {
				log.Errorf("log admin operation error: %v", err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (a *AdminService) CreateGroup(ctx context.Context, name string, memberIDs []peer.ID) (string, error) {

	groupID := ds.GroupID(uuid.NewString())
	peerID := a.host.ID().String()

	// 创建组
	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
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

	if err = a.data.LogAdminOperation(ctx, a.host.ID(), groupID, &createLog); err != nil {
		return string(groupID), err
	}

	if err := a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_CREATE,
		MsgData: &createLog,
	}); err != nil {
		return "", err
	}

	// 设置名称
	lamportTime, err = a.data.TickLamportTime(ctx, groupID)
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
	if err = a.data.LogAdminOperation(ctx, a.host.ID(), groupID, &nameLog); err != nil {
		return string(groupID), err
	}

	if err := a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_NAME,
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

		if err = a.data.LogAdminOperation(ctx, a.host.ID(), groupID, &memberLog); err != nil {
			return string(groupID), err
		}

		// todo: 要单独给人发消息
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

	if err = a.data.LogAdminOperation(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_DISBAND,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AdminService) ListGroups(ctx context.Context) ([]ds.Group, error) {
	return a.data.ListGroups(ctx)
}

func (a *AdminService) GroupName(ctx context.Context, groupID string) (string, error) {

	if remark, err := a.data.GroupRemark(ctx, ds.GroupID(groupID)); err != nil {
		return "", err

	} else if remark != "" {
		return remark, nil
	}

	return a.data.GroupName(ctx, ds.GroupID(groupID))
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

	if err := a.data.LogAdminOperation(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_NAME,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
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

	if err := a.data.LogAdminOperation(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_NOTICE,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AdminService) SetGroupRemark(ctx context.Context, groupID string, remark string) error {
	// 只是更新本地
	return a.data.SetGroupRemark(ctx, ds.GroupID(groupID), remark)
}

func (a *AdminService) InviteMember(ctx context.Context, groupID0 string, peerID0 peer.ID) error {

	groupID := ds.GroupID(groupID0)
	peerID := a.host.ID().String()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
	if err != nil {
		return err
	}

	return a.data.LogAdminOperation(ctx, a.host.ID(), groupID, &pb.AdminLog{
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

func (a *AdminService) ApplyMember(ctx context.Context, groupID string, memberID peer.ID, lamportime uint64) error {

	peerID := a.host.ID().String()

	stream, err := a.host.NewStream(ctx, memberID, ID)
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

func (a *AdminService) ReviewMember(ctx context.Context, groupID0 string, memberID0 peer.ID, isAgree bool) error {
	groupID := ds.GroupID(groupID0)
	peerID := a.host.ID().String()

	lamportTime, err := a.data.TickLamportTime(ctx, groupID)
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

	if err := a.data.LogAdminOperation(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_MEMBER,
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

	if err := a.data.LogAdminOperation(ctx, a.host.ID(), groupID, &pbmsg); err != nil {
		return err
	}

	err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.AdminLog_MEMBER,
		MsgData: &pbmsg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AdminService) ListMembers(ctx context.Context, groupID0 string) ([]ds.Member, error) {

	groupID := ds.GroupID(groupID0)
	memberLogs, err := a.data.GroupMemberLogs(ctx, groupID)
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

	var members []ds.Member
	for memberID := range oks {
		peerID, err := peer.Decode(memberID)
		if err != nil {
			return nil, err
		}
		members = append(members, ds.Member{
			PeerID: peerID,
		})
	}

	return members, nil
}
