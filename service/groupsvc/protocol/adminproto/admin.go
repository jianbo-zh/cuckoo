package adminproto

// 群管理相关协议

import (
	"context"
	"fmt"
	"strings"
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
	"google.golang.org/protobuf/proto"
)

var log = logging.Logger("group-admin")

var StreamTimeout = 1 * time.Minute

const (
	ID      = protocol.GroupAdminID_v100
	SYNC_ID = protocol.GroupAdminSyncID_v100

	ServiceName = "group.admin"
	maxMsgSize  = 4 * 1024 // 4K
)

type GroupLogStats struct {
	Creator peer.ID
	Name    string
	Avatar  string
	Members []types.GroupMember
}

type AdminProto struct {
	host host.Host

	data ds.AdminIface

	emitters struct {
		evtSendAdminLog    event.Emitter
		evtInviteJoinGroup event.Emitter
		evtGroupsChange    event.Emitter
	}

	groupConns map[string]map[peer.ID]struct{}
	groupStats map[string]*GroupLogStats
}

func NewAdminProto(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus) (*AdminProto, error) {
	var err error

	admsvc := &AdminProto{
		host:       lhost,
		data:       ds.AdminWrap(ids),
		groupConns: make(map[string]map[peer.ID]struct{}),
		groupStats: make(map[string]*GroupLogStats),
	}

	lhost.SetStreamHandler(ID, admsvc.handler)
	lhost.SetStreamHandler(SYNC_ID, admsvc.syncHandler)

	if admsvc.emitters.evtSendAdminLog, err = eventBus.Emitter(&gevent.EvtSendAdminLog{}); err != nil {
		return nil, fmt.Errorf("set emitter error: %v", err)
	}

	if admsvc.emitters.evtInviteJoinGroup, err = eventBus.Emitter(&gevent.EvtInviteJoinGroup{}); err != nil {
		return nil, fmt.Errorf("set emitter error: %v", err)
	}

	if admsvc.emitters.evtGroupsChange, err = eventBus.Emitter(&gevent.EvtGroupsChange{}); err != nil {
		return nil, fmt.Errorf("set group change emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(gevent.EvtGroupConnectChange)}, eventbus.Name("adminlog"))
	if err != nil {
		return nil, fmt.Errorf("subscription failed. group admin server error: %v", err)

	} else {
		go admsvc.subscribeHandler(context.Background(), sub)
	}

	return admsvc, nil
}

func (m *AdminProto) syncHandler(stream network.Stream) {
	defer stream.Close()

	// 获取同步GroupID
	var syncmsg pb.SyncLog
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	if err := rd.ReadMsg(&syncmsg); err != nil {
		log.Errorf("pbio read sync init msg error: %v", err)
		return
	}

	groupID := syncmsg.GroupId

	summary, err := m.getMessageSummary(groupID)
	if err != nil {
		log.Errorf("get msg summary error: %v", err)
		return
	}

	bs, err := proto.Marshal(summary)
	if err != nil {
		log.Errorf("proto marshal summary error: %v", err)
		return
	}

	wt := pbio.NewDelimitedWriter(stream)
	if err = wt.WriteMsg(&pb.SyncLog{
		Type:    pb.SyncLog_SUMMARY,
		Payload: bs,
	}); err != nil {
		log.Errorf("pbio write sync summary error: %v", err)
		return
	}

	// 后台同步处理
	err = m.loopSync(groupID, stream, rd, wt)
	if err != nil {
		log.Errorf("loop sync error: %v", err)
	}
}

func (a *AdminProto) handler(stream network.Stream) {
	if err := stream.Scope().SetService(ServiceName); err != nil {
		log.Errorf("failed to attaching stream to identify service: %v", err)
		stream.Reset()
		return
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rd.Close()

	stream.SetDeadline(time.Now().Add(StreamTimeout))

	var msg pb.Log
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %v", err)
		stream.Reset()
		return
	}

	stream.SetReadDeadline(time.Time{})

	err := a.data.SaveLog(context.Background(), &msg)
	if err != nil {
		log.Errorf("save log error: %v", err)
		stream.Reset()
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

			// 启动同步: peerID小的主动发起
			if a.host.ID() < evt.PeerID {
				a.goSync(evt.GroupID, evt.PeerID)
			}

		case <-ctx.Done():
			return
		}
	}
}

// AgreeJoinGroup 同意加入群
func (a *AdminProto) AgreeJoinGroup(ctx context.Context, account *types.Account, group *types.Group, lamptime uint64) error {

	if err := a.data.MergeLamptime(ctx, group.ID, lamptime); err != nil {
		return fmt.Errorf("data merge lamptime error: %w", err)
	}

	lamportime, err := a.data.TickLamptime(ctx, group.ID)
	if err != nil {
		return fmt.Errorf("data tick lamptime error: %w", err)
	}

	memberLog := &pb.Log{
		Id:      fmt.Sprintf("%d_%s_%s", lamportime, "member", account.ID.String()),
		GroupId: group.ID,
		PeerId:  []byte(account.ID),
		LogType: pb.Log_MEMBER,
		Member: &pb.Log_Member{
			Id:     []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MemberOperate: pb.Log_APPLY,
		Payload:       []byte(""),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamportime,
		Signature:     []byte(""),
	}

	if err = a.data.SaveLog(ctx, memberLog); err != nil {
		return fmt.Errorf("save log error: %w", err)
	}

	if err = a.data.SetName(ctx, group.ID, group.Name); err != nil {
		return fmt.Errorf("ds set group name error: %w", err)
	}

	if err = a.data.SetAvatar(ctx, group.ID, group.Avatar); err != nil {
		return fmt.Errorf("ds set group avatar error: %w", err)
	}

	if err = a.data.SetState(ctx, group.ID, types.GroupStateNormal); err != nil {
		return fmt.Errorf("ds set group state error: %w", err)
	}

	if err = a.data.SetSession(ctx, group.ID); err != nil {
		return fmt.Errorf("ds set group session error: %w", err)
	}

	if err = a.emitters.evtGroupsChange.Emit(gevent.EvtGroupsChange{
		AddGroups: []gevent.Groups{
			{
				GroupID:     group.ID,
				PeerIDs:     []peer.ID{},
				AcptPeerIDs: []peer.ID{},
			},
		},
	}); err != nil {
		return fmt.Errorf("emite add group error: %w", err)
	}

	return nil
}

func (a *AdminProto) CreateGroup(ctx context.Context, account *types.Account, name string, avatar string, members []types.Contact) (*types.Group, error) {

	hostID := a.host.ID()
	groupID := uuid.NewString()

	// 创建组
	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return nil, err
	}
	createLog := pb.Log{
		Id:      fmt.Sprintf("%d_%s_%s", lamptime, "create", hostID.String()),
		GroupId: groupID,
		PeerId:  []byte(hostID),
		LogType: pb.Log_CREATE,
		Member: &pb.Log_Member{
			Id:     []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MemberOperate: pb.Log_NONE,
		Payload:       []byte(hostID),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if err = a.data.SaveLog(ctx, &createLog); err != nil {
		return nil, err
	}

	// 设置名称
	lamptime, err = a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return nil, err
	}
	nameLog := pb.Log{
		Id:            fmt.Sprintf("%d_%s_%s", lamptime, "name", hostID.String()),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.Log_NAME,
		Member:        nil,
		MemberOperate: pb.Log_NONE,
		Payload:       []byte(name),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}
	if err = a.data.SaveLog(ctx, &nameLog); err != nil {
		return nil, err
	}

	// 设置头像
	lamptime, err = a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return nil, err
	}
	avatarLog := pb.Log{
		Id:            fmt.Sprintf("%d_%s_%s", lamptime, "avatar", hostID.String()),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.Log_AVATAR,
		Member:        nil,
		MemberOperate: pb.Log_NONE,
		Payload:       []byte(avatar),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}
	if err = a.data.SaveLog(ctx, &avatarLog); err != nil {
		return nil, err
	}

	var memberIDs []peer.ID
	memberIDs = append(memberIDs, account.ID)

	// 设置创建者
	lamptime, err = a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return nil, err
	}

	memberLog := pb.Log{
		Id:      fmt.Sprintf("%d_%s_%s", lamptime, "member", hostID.String()),
		GroupId: groupID,
		PeerId:  []byte(hostID),
		LogType: pb.Log_MEMBER,
		Member: &pb.Log_Member{
			Id:     []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MemberOperate: pb.Log_CREATOR,
		Payload:       []byte(hostID),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if err = a.data.SaveLog(ctx, &memberLog); err != nil {
		return nil, err
	}

	// 设置其他成员
	for _, member := range members {
		memberIDs = append(memberIDs, member.ID)

		lamptime, err = a.data.TickLamptime(ctx, groupID)
		if err != nil {
			return nil, err
		}

		memberLog := pb.Log{
			Id:      fmt.Sprintf("%d_%s_%s", lamptime, "member", hostID.String()),
			GroupId: groupID,
			PeerId:  []byte(hostID),
			LogType: pb.Log_MEMBER,
			Member: &pb.Log_Member{
				Id:     []byte(member.ID),
				Name:   member.Name,
				Avatar: member.Avatar,
			},
			MemberOperate: pb.Log_AGREE,
			Payload:       []byte(member.ID),
			CreateTime:    time.Now().Unix(),
			Lamportime:    lamptime,
			Signature:     []byte(""),
		}

		if err = a.data.SaveLog(ctx, &memberLog); err != nil {
			return nil, err
		}
	}

	if err = a.data.SetState(ctx, groupID, types.GroupStateNormal); err != nil {
		return nil, fmt.Errorf("ds set group state error: %w", err)
	}

	if err = a.data.SetSession(ctx, groupID); err != nil {
		return nil, fmt.Errorf("ds set group session error: %w", err)
	}

	// 触发新增群组
	if err = a.emitters.evtGroupsChange.Emit(gevent.EvtGroupsChange{
		AddGroups: []gevent.Groups{
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
	lamptime, err = a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err = a.emitters.evtInviteJoinGroup.Emit(gevent.EvtInviteJoinGroup{
		PeerIDs:       memberIDs,
		GroupID:       groupID,
		GroupName:     name,
		GroupAvatar:   avatar,
		GroupLamptime: lamptime,
	}); err != nil {
		return nil, fmt.Errorf("emit invite join group error: %w", err)
	}

	return &types.Group{
		ID:     groupID,
		Name:   name,
		Avatar: avatar,
	}, nil
}

// ExitGroup 退出群
func (a *AdminProto) ExitGroup(ctx context.Context, account *types.Account, groupID string) error {

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:      fmt.Sprintf("%d_%s_%s", lamptime, "exit", account.ID.String()),
		GroupId: groupID,
		PeerId:  []byte(account.ID),
		LogType: pb.Log_MEMBER,
		Member: &pb.Log_Member{
			Id:     []byte(account.ID),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MemberOperate: pb.Log_EXIT,
		Payload:       []byte{},
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if err = a.data.SaveLog(ctx, &pbmsg); err != nil {
		return err
	}

	if err = a.emitters.evtSendAdminLog.Emit(gevent.EvtSendAdminLog{
		MsgType: pb.Log_MEMBER,
		MsgData: &pbmsg,
	}); err != nil {
		return err
	}

	if err = a.data.SetState(ctx, groupID, types.GroupStateExit); err != nil {
		return fmt.Errorf("set state exit error: %w", err)
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
		lamptime, err := a.data.TickLamptime(ctx, groupID)
		if err != nil {
			return err
		}

		pbmsg := pb.Log{
			Id:      fmt.Sprintf("%d_%s_%s", lamptime, "exit", account.ID.String()),
			GroupId: groupID,
			PeerId:  []byte(account.ID),
			LogType: pb.Log_MEMBER,
			Member: &pb.Log_Member{
				Id:     []byte(account.ID),
				Name:   account.Name,
				Avatar: account.Avatar,
			},
			MemberOperate: pb.Log_EXIT,
			Payload:       []byte{},
			CreateTime:    time.Now().Unix(),
			Lamportime:    lamptime,
			Signature:     []byte(""),
		}

		if err = a.data.SaveLog(ctx, &pbmsg); err != nil {
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
		return fmt.Errorf("data delete group error: %w", err)
	}

	return nil
}

// DisbandGroup 解散群（所有人退出）
func (a *AdminProto) DisbandGroup(ctx context.Context, account *types.Account, groupID string) error {

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:            fmt.Sprintf("%d_%s_%s", lamptime, types.GroupStateDisband, account.ID.String()),
		GroupId:       groupID,
		PeerId:        []byte(account.ID),
		LogType:       pb.Log_DISBAND,
		Member:        nil,
		MemberOperate: pb.Log_NONE,
		Payload:       []byte(groupID),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if err = a.data.SaveLog(ctx, &pbmsg); err != nil {
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

func (a *AdminProto) GetSessionIDs(ctx context.Context) ([]string, error) {
	return a.data.GetSessionIDs(ctx)
}

func (a *AdminProto) GetSessions(ctx context.Context) ([]types.Group, error) {
	groupIDs, err := a.data.GetSessionIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("data get session ids error: %w", err)
	}

	var groups []types.Group
	for _, groupID := range groupIDs {
		name, err := a.data.GetName(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("data get name error: %w", err)
		}

		avatar, err := a.data.GetAvatar(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("data get name error: %w", err)
		}

		groups = append(groups, types.Group{
			ID:     groupID,
			Name:   name,
			Avatar: avatar,
		})
	}

	return groups, nil
}

func (a *AdminProto) GetGroup(ctx context.Context, groupID string) (*types.Group, error) {
	name, err := a.data.GetName(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get name error: %w", err)
	}

	avatar, err := a.data.GetAvatar(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get name error: %w", err)
	}

	return &types.Group{
		ID:     groupID,
		Name:   name,
		Avatar: avatar,
	}, nil
}

func (a *AdminProto) GetGroupDetail(ctx context.Context, groupID string) (*types.GroupDetail, error) {
	name, err := a.data.GetName(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get name error: %w", err)
	}

	avatar, err := a.data.GetAvatar(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get name error: %w", err)
	}

	notice, err := a.data.GetNotice(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get name error: %w", err)
	}

	autoJoinGroup, err := a.data.GetAutoJoinGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get name error: %w", err)
	}

	createTime, err := a.data.GetCreateTime(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get name error: %w", err)
	}

	return &types.GroupDetail{
		ID:            groupID,
		Name:          name,
		Avatar:        avatar,
		Notice:        notice,
		AutoJoinGroup: autoJoinGroup,
		CreateTime:    createTime,
	}, nil
}

func (a *AdminProto) SetGroupName(ctx context.Context, groupID string, name string) error {

	hostID := a.host.ID()

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:            fmt.Sprintf("%d_%s_%s", lamptime, "name", hostID.String()),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.Log_NAME,
		Member:        nil,
		MemberOperate: pb.Log_NONE,
		Payload:       []byte(name),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
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

func (a *AdminProto) SetGroupAlias(ctx context.Context, groupID string, alias string) error {
	err := a.data.SetAlias(ctx, groupID, alias)
	if err != nil {
		return fmt.Errorf("data set group alias error: %w", err)
	}

	return nil
}

func (a *AdminProto) SetGroupAvatar(ctx context.Context, groupID string, avatar string) error {

	hostID := a.host.ID()

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:            fmt.Sprintf("%d_%s_%s", lamptime, "avatar", hostID.String()),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.Log_AVATAR,
		Member:        nil,
		MemberOperate: pb.Log_NONE,
		Payload:       []byte(avatar),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
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

	pbmsg := pb.Log{
		Id:            fmt.Sprintf("%d_%s_%s", lamptime, "options", hostID.String()),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.Log_AUTO_JOIN_GROUP,
		Member:        nil,
		MemberOperate: pb.Log_NONE,
		Payload:       []byte(autoJoin),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
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

func (a *AdminProto) SetGroupNotice(ctx context.Context, groupID string, notice string) error {
	hostID := a.host.ID()

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:            fmt.Sprintf("%d_%s_%s", lamptime, "notice", hostID.String()),
		GroupId:       groupID,
		PeerId:        []byte(hostID),
		LogType:       pb.Log_NOTICE,
		Member:        nil,
		MemberOperate: pb.Log_NONE,
		Payload:       []byte(notice),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
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

func (a *AdminProto) ReviewJoinGroup(ctx context.Context, groupID string, member *types.Peer, isAgree bool) error {
	hostID := a.host.ID()

	creator, err := a.data.GetCreator(ctx, groupID)
	if hostID != creator {
		return fmt.Errorf("no permission to review join group")
	}

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	operate := pb.Log_AGREE
	if !isAgree {
		operate = pb.Log_REJECTED
	}

	pbmsg := pb.Log{
		Id:      fmt.Sprintf("%d_%s_%s", lamptime, "member", hostID.String),
		GroupId: groupID,
		PeerId:  []byte(hostID),
		LogType: pb.Log_MEMBER,
		Member: &pb.Log_Member{
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

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
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

	creator, err := a.data.GetCreator(ctx, groupID)
	if hostID != creator {
		return fmt.Errorf("no permission to remove member")
	}

	members, err := a.data.GetMembers(ctx, groupID)
	if err != nil {
		return fmt.Errorf("ds get group members error: %w", err)
	}

	var findMember types.GroupMember
	for _, member := range members {
		if member.ID == memberID {
			findMember = member
			break
		}
	}

	if findMember.ID != memberID {
		return fmt.Errorf("not found remove member")
	}

	lamptime, err := a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return err
	}

	pbmsg := pb.Log{
		Id:      fmt.Sprintf("%d_%s_%s", lamptime, "member", hostID.String()),
		GroupId: groupID,
		PeerId:  []byte(hostID),
		LogType: pb.Log_MEMBER,
		Member: &pb.Log_Member{
			Id:     []byte(findMember.ID),
			Name:   findMember.Name,
			Avatar: findMember.Avatar,
		},
		MemberOperate: pb.Log_REMOVE,
		Payload:       []byte(""),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
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
	members, err := a.data.GetMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("ds get group members error: %w", err)
	}

	memberIDs := make([]peer.ID, len(members))
	for i, member := range members {
		memberIDs[i] = member.ID
	}

	return memberIDs, nil
}

func (a *AdminProto) GetAgreeMemberIDs(ctx context.Context, groupID string) ([]peer.ID, error) {
	members, err := a.data.GetAgreeMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("ds get group members error: %w", err)
	}

	memberIDs := make([]peer.ID, len(members))
	for i, member := range members {
		memberIDs[i] = member.ID
	}

	return memberIDs, nil
}

func (a *AdminProto) GetGroupMembers(ctx context.Context, groupID string, keywords string, offset int, limit int) ([]types.GroupMember, error) {
	if offset < 0 || limit <= 0 {
		return nil, nil
	}

	members, err := a.data.GetMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get members error: %w", err)
	}

	var filters []types.GroupMember
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

	return filters[offset:endOffset], nil
}
