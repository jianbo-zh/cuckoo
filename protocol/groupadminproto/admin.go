package groupadminproto

// 群管理相关协议

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	ipfsds "github.com/ipfs/go-datastore"
	ds "github.com/jianbo-zh/dchat/datastore/ds/groupadminds"
	"github.com/jianbo-zh/dchat/datastore/ds/sessionds"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/internal/protocol"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/grouppb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
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
	Members []mytype.GroupMember
}

type AdminProto struct {
	host myhost.Host

	data ds.AdminIface

	sessionData sessionds.SessionIface

	emitters struct {
		evtInviteJoinGroup event.Emitter
		evtGroupsChange    event.Emitter
		evtCheckAvatar     event.Emitter
	}

	groupConns map[string]map[peer.ID]struct{}
	groupStats map[string]*GroupLogStats
}

func NewAdminProto(lhost myhost.Host, ids ipfsds.Batching, eventBus event.Bus) (*AdminProto, error) {
	var err error

	admsvc := &AdminProto{
		host:        lhost,
		data:        ds.AdminWrap(ids),
		sessionData: sessionds.SessionWrap(ids),
		groupConns:  make(map[string]map[peer.ID]struct{}),
		groupStats:  make(map[string]*GroupLogStats),
	}

	lhost.SetStreamHandler(ID, admsvc.handler)
	lhost.SetStreamHandler(SYNC_ID, admsvc.syncHandler)

	if admsvc.emitters.evtInviteJoinGroup, err = eventBus.Emitter(&myevent.EvtInviteJoinGroup{}); err != nil {
		return nil, fmt.Errorf("set emitter error: %v", err)
	}

	if admsvc.emitters.evtGroupsChange, err = eventBus.Emitter(&myevent.EvtGroupsChange{}); err != nil {
		return nil, fmt.Errorf("set group change emitter error: %v", err)
	}

	if admsvc.emitters.evtCheckAvatar, err = eventBus.Emitter(&myevent.EvtCheckAvatar{}); err != nil {
		return nil, fmt.Errorf("set group change emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(myevent.EvtGroupConnectChange), new(myevent.EvtGroupNetworkSuccess)}, eventbus.Name("adminlog"))
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
	var syncmsg pb.GroupSyncLog
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
	if err = wt.WriteMsg(&pb.GroupSyncLog{
		Type:    pb.GroupSyncLog_SUMMARY,
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

	remotePeerID := stream.Conn().RemotePeer()

	if err := stream.Scope().SetService(ServiceName); err != nil {
		log.Errorf("failed to attaching stream to identify service: %v", err)
		stream.Reset()
		return
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rd.Close()

	stream.SetDeadline(time.Now().Add(StreamTimeout))

	var msg pb.GroupLog
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %v", err)
		stream.Reset()
		return
	}

	stream.SetReadDeadline(time.Time{})

	ctx := context.Background()
	if _, err := a.data.GetLog(ctx, msg.GroupId, msg.Id); err != nil {
		if errors.Is(err, ipfsds.ErrNotFound) {
			// 保存日志
			if err = a.data.SaveLog(ctx, &msg); err != nil {
				log.Errorf("save log error: %v", err)
				stream.Reset()
				return
			}

			// 转发消息
			err = a.broadcastMessage(msg.GroupId, &msg, remotePeerID)
			if err != nil {
				log.Errorf("a.broadcast message error: %v", err)
				stream.Reset()
				return
			}
		} else {
			log.Errorf("data get admin log error: %v", err)
			stream.Reset()
			return
		}
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

			switch evt := e.(type) {
			case myevent.EvtGroupConnectChange:
				if evt.IsConnected {
					if _, exists := a.groupConns[evt.GroupID]; !exists {
						a.groupConns[evt.GroupID] = make(map[peer.ID]struct{})
					}

					if _, exists := a.groupConns[evt.GroupID][evt.PeerID]; !exists {
						a.groupConns[evt.GroupID][evt.PeerID] = struct{}{}

						if a.host.ID() < evt.PeerID {
							// 小方主动发起同步
							go a.goSync(evt.GroupID, evt.PeerID)
						}
					}
					log.Debugln("event connect peer: ", evt.PeerID.String())

				} else {
					log.Debugln("event disconnect peer: ", evt.PeerID.String())
					delete(a.groupConns[evt.GroupID], evt.PeerID)
				}
			case myevent.EvtGroupNetworkSuccess:
				if avatar, err := a.data.GetAvatar(ctx, evt.GroupID); err != nil {
					log.Errorf("data get group avatar error: %v", err)

				} else {
					if err = a.emitters.evtCheckAvatar.Emit(myevent.EvtCheckAvatar{
						Avatar:  avatar,
						PeerIDs: evt.PeerIDs,
					}); err != nil {
						log.Errorf("emite check avatar event error: %w", err)
					}
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

// 发送消息
func (a *AdminProto) broadcastMessage(groupID string, msg *pb.GroupLog, excludePeerIDs ...peer.ID) error {

	connectPeerIDs := a.getConnectPeers(groupID)
	fmt.Printf("get connect peers: %v\n", connectPeerIDs)

	for _, peerID := range connectPeerIDs {
		if len(excludePeerIDs) > 0 {
			isExcluded := false
			for _, excludePeerID := range excludePeerIDs {
				if peerID == excludePeerID {
					isExcluded = true
				}
			}
			if isExcluded {
				continue
			}
		}

		go a.goSendAdminMessage(peerID, msg)
	}

	return nil
}

func (a *AdminProto) getConnectPeers(groupID string) []peer.ID {
	var peerIDs []peer.ID
	for peerID := range a.groupConns[groupID] {
		peerIDs = append(peerIDs, peerID)
	}
	return peerIDs
}

// 发送消息（指定peerID）
func (a *AdminProto) goSendAdminMessage(peerID peer.ID, msg *pb.GroupLog) {
	ctx := context.Background()
	stream, err := a.host.NewStream(network.WithDialPeerTimeout(ctx, mytype.DialTimeout), peerID, ID)
	if err != nil {
		log.Errorf("host new stream error: %v", err)
		return
	}

	stream.SetWriteDeadline(time.Now().Add(StreamTimeout))

	wt := pbio.NewDelimitedWriter(stream)
	defer wt.Close()

	if err := wt.WriteMsg(msg); err != nil {
		log.Errorf("pbio write admin msg error: %v", err)
		return
	}
}

// AgreeJoinGroup 同意加入群
func (a *AdminProto) AgreeJoinGroup(ctx context.Context, account *mytype.Account, group *mytype.Group, lamptime uint64) error {

	if err := a.data.MergeLamptime(ctx, group.ID, lamptime); err != nil {
		return fmt.Errorf("data merge lamptime error: %w", err)
	}

	lamptime, err := a.data.TickLamptime(ctx, group.ID)
	if err != nil {
		return fmt.Errorf("data tick lamptime error: %w", err)
	}

	memberLog := &pb.GroupLog{
		Id:      logIdMember(lamptime, account.ID),
		GroupId: group.ID,
		PeerId:  []byte(account.ID),
		LogType: pb.GroupLog_MEMBER,
		Member: &pb.GroupLog_Member{
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

	if err = a.data.SaveLog(ctx, memberLog); err != nil {
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
				PeerIDs:     []peer.ID{},
				AcptPeerIDs: []peer.ID{},
			},
		},
	}); err != nil {
		return fmt.Errorf("emite add group error: %w", err)
	}

	return nil
}

func (a *AdminProto) CreateGroup(ctx context.Context, account *mytype.Account, name string, avatar string, members []mytype.Contact) (*mytype.Group, error) {

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
		Member: &pb.GroupLog_Member{
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

	if err = a.data.SaveLog(ctx, &createLog); err != nil {
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
	if err = a.data.SaveLog(ctx, &nameLog); err != nil {
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

	memberLog := pb.GroupLog{
		Id:      logIdMember(lamptime, hostID),
		GroupId: groupID,
		PeerId:  []byte(hostID),
		LogType: pb.GroupLog_MEMBER,
		Member: &pb.GroupLog_Member{
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

		memberLog := pb.GroupLog{
			Id:      logIdMember(lamptime, hostID),
			GroupId: groupID,
			PeerId:  []byte(hostID),
			LogType: pb.GroupLog_MEMBER,
			Member: &pb.GroupLog_Member{
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

		if err = a.data.SaveLog(ctx, &memberLog); err != nil {
			return nil, err
		}
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
	lamptime, err = a.data.TickLamptime(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err = a.emitters.evtInviteJoinGroup.Emit(myevent.EvtInviteJoinGroup{
		PeerIDs:       memberIDs,
		GroupID:       groupID,
		GroupName:     name,
		GroupAvatar:   avatar,
		GroupLamptime: lamptime,
	}); err != nil {
		return nil, fmt.Errorf("emit invite join group error: %w", err)
	}

	return &mytype.Group{
		ID:     groupID,
		Name:   name,
		Avatar: avatar,
	}, nil
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
		Member: &pb.GroupLog_Member{
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

	if err = a.data.SaveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	if err = a.data.SetState(ctx, groupID, mytype.GroupStateExit); err != nil {
		return fmt.Errorf("set state exit error: %w", err)
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
			Member: &pb.GroupLog_Member{
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

		if err = a.data.SaveLog(ctx, &pbmsg); err != nil {
			return fmt.Errorf("data save log error: %w", err)
		}

		if err = a.broadcastMessage(groupID, &pbmsg); err != nil {
			return fmt.Errorf("broadcast msg error: %w", err)
		}
	}

	if err = a.data.DeleteGroup(ctx, groupID); err != nil {
		return fmt.Errorf("data delete group error: %w", err)
	}

	return nil
}

// DisbandGroup 解散群（所有人退出）
func (a *AdminProto) DisbandGroup(ctx context.Context, account *mytype.Account, groupID string) error {

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

	if err = a.data.SaveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	return nil
}

func (a *AdminProto) GetGroupIDs(ctx context.Context) ([]string, error) {
	return a.data.GetGroupIDs(ctx)
}

func (a *AdminProto) GetSessions(ctx context.Context) ([]mytype.Group, error) {
	groupIDs, err := a.data.GetGroupIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("data get session ids error: %w", err)
	}

	var groups []mytype.Group
	for _, groupID := range groupIDs {
		name, err := a.data.GetName(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("data get name error: %w", err)
		}

		avatar, err := a.data.GetAvatar(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("data get name error: %w", err)
		}

		groups = append(groups, mytype.Group{
			ID:     groupID,
			Name:   name,
			Avatar: avatar,
		})
	}

	return groups, nil
}

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

func (a *AdminProto) GetGroupDetail(ctx context.Context, groupID string) (*mytype.GroupDetail, error) {
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

	createTime, err := a.data.GetCreateTime(ctx, groupID)
	if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("data get create time error: %w", err)
	}

	return &mytype.GroupDetail{
		ID:             groupID,
		Name:           name,
		Avatar:         avatar,
		Notice:         notice,
		AutoJoinGroup:  autoJoinGroup,
		DepositAddress: depositPeerID,
		CreateTime:     createTime,
	}, nil
}

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

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
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

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
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

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	return nil
}

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

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	return nil
}

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

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	return nil
}

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
		Member: &pb.GroupLog_Member{
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
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
	}

	return nil
}

func (a *AdminProto) RemoveMember(ctx context.Context, groupID string, memberID peer.ID) error {
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

	var findMember mytype.GroupMember
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

	pbmsg := pb.GroupLog{
		Id:      logIdMember(lamptime, hostID),
		GroupId: groupID,
		PeerId:  []byte(hostID),
		LogType: pb.GroupLog_MEMBER,
		Member: &pb.GroupLog_Member{
			Id:     []byte(findMember.ID),
			Name:   findMember.Name,
			Avatar: findMember.Avatar,
		},
		MemberOperate: pb.GroupLog_REMOVE,
		Payload:       []byte(""),
		CreateTime:    time.Now().Unix(),
		Lamportime:    lamptime,
		Signature:     []byte(""),
	}

	if err := a.data.SaveLog(ctx, &pbmsg); err != nil {
		return fmt.Errorf("data save log error: %w", err)
	}

	if err = a.broadcastMessage(groupID, &pbmsg); err != nil {
		return fmt.Errorf("broadcast msg error: %w", err)
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

func (a *AdminProto) GetGroupMembers(ctx context.Context, groupID string, keywords string, offset int, limit int) ([]mytype.GroupMember, error) {
	if offset < 0 || limit <= 0 {
		return nil, nil
	}

	members, err := a.data.GetMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get members error: %w", err)
	}

	var filters []mytype.GroupMember
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
