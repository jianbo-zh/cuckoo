package groupadminproto

// 群管理相关协议

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/myprotocol"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/service/common/datastore/ds/sessionds"
	ds "github.com/jianbo-zh/dchat/service/groupsvc/datastore/ds/groupadminds"
	pb "github.com/jianbo-zh/dchat/service/groupsvc/protobuf/pb/grouppb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
	"google.golang.org/protobuf/proto"
)

var log = logging.Logger("grougadminproto")

var StreamTimeout = 1 * time.Minute

const (
	LOG_ID  = myprotocol.GroupAdminID_v100
	PULL_ID = myprotocol.GroupAdminPullID_v100
	SYNC_ID = myprotocol.GroupAdminSyncID_v100

	ServiceName = "group.admin"
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
		evtInviteJoinGroup   event.Emitter
		evtGroupsChange      event.Emitter
		evtGroupMemberChange event.Emitter
		evtSyncResource      event.Emitter
		evtSyncGroupMessage  event.Emitter
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

	lhost.SetStreamHandler(LOG_ID, admsvc.logHandler)
	lhost.SetStreamHandler(PULL_ID, admsvc.pullHandler)
	lhost.SetStreamHandler(SYNC_ID, admsvc.syncHandler)

	if admsvc.emitters.evtInviteJoinGroup, err = eventBus.Emitter(&myevent.EvtInviteJoinGroup{}); err != nil {
		return nil, fmt.Errorf("set emitter error: %v", err)
	}

	if admsvc.emitters.evtGroupsChange, err = eventBus.Emitter(&myevent.EvtGroupsChange{}); err != nil {
		return nil, fmt.Errorf("set group change emitter error: %v", err)
	}

	if admsvc.emitters.evtGroupMemberChange, err = eventBus.Emitter(&myevent.EvtGroupMemberChange{}); err != nil {
		return nil, fmt.Errorf("set group change emitter error: %v", err)
	}

	if admsvc.emitters.evtSyncResource, err = eventBus.Emitter(&myevent.EvtSyncResource{}); err != nil {
		return nil, fmt.Errorf("set sync resource emitter error: %v", err)
	}

	if admsvc.emitters.evtSyncGroupMessage, err = eventBus.Emitter(&myevent.EvtSyncGroupMessage{}); err != nil {
		return nil, fmt.Errorf("set sync group message emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(myevent.EvtGroupNetworkSuccess), new(myevent.EvtGroupConnectChange), new(myevent.EvtPullGroupLog)})

	if err != nil {
		return nil, fmt.Errorf("subscription failed. group admin server error: %v", err)
	}

	go admsvc.handleSubscribe(context.Background(), sub)

	return admsvc, nil
}

func (a *AdminProto) logHandler(stream network.Stream) {

	remotePeerID := stream.Conn().RemotePeer()

	if err := stream.Scope().SetService(ServiceName); err != nil {
		log.Errorf("failed to attaching stream to identify service: %v", err)
		stream.Reset()
		return
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	defer rd.Close()

	stream.SetDeadline(time.Now().Add(StreamTimeout))

	var msg pb.GroupLog
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("failed to read CONNECT message from remote peer: %v", err)
		stream.Reset()
		return
	}

	stream.SetReadDeadline(time.Time{})

	// 处理日志
	if err := a.receiveLog(context.Background(), &msg, remotePeerID); err != nil {
		log.Errorf("save log error: %v", err)
		stream.Reset()
		return
	}
}

// pullHandler 获取消息处理器
func (a *AdminProto) pullHandler(stream network.Stream) {

	defer stream.Close()

	var ctx = context.Background()

	// 接收对方消息摘要
	var recvSumMsg pb.GroupLogHash
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	if err := rd.ReadMsg(&recvSumMsg); err != nil {
		log.Errorf("pbio read msg error: %v", err)
		return
	}

	groupID := recvSumMsg.GroupId

	// 首先计算消息摘要（头尾md5）
	logIDs, err := a.data.GetLogIDs(ctx, groupID)
	if err != nil {
		log.Errorf("data.GetLogIDs error: %v", err)
		stream.Reset()
		return
	}

	var logHash []byte
	var headID, tailID string

	logIDNum := len(logIDs)
	if logIDNum > 0 {
		md5sum := md5.New()
		for _, logID := range logIDs {
			if _, err := md5sum.Write([]byte(logID)); err != nil {
				log.Errorf("md5sum write error: %w", err)
				stream.Reset()
				return
			}
		}
		headID = logIDs[0]
		tailID = logIDs[logIDNum-1]
		logHash = md5sum.Sum(nil)
	}

	// 计算并发送本地消息摘要
	sendSumMsg := pb.GroupLogHash{
		HeadId: headID,
		TailId: tailID,
		Length: int32(logIDNum),
		Hash:   logHash,
	}
	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&sendSumMsg); err != nil {
		log.Errorf("pbio write msg error: %w", err)
		return
	}

	// 如果本地和远端都一样，则正常退出
	if string(recvSumMsg.Hash) == string(sendSumMsg.Hash) {
		log.Debugln("summary hash equal")
		return
	}

	// 不一致，准备接收对方的消息ID列表，然后进行比对
	var groupLogIDs []string
	var groupLogIDPack pb.GroupLogIDPack
	var leftNum int32
	for {
		groupLogIDPack.Reset()
		if err := rd.ReadMsg(&groupLogIDPack); err != nil {
			log.Errorf("pbio read log pack error: %w", err)
			return
		}

		if groupLogIDPack.LogId != "" {
			groupLogIDs = append(groupLogIDs, groupLogIDPack.LogId)
		}

		if groupLogIDPack.LeftNum <= 0 {
			// 无剩余则退出
			break

		} else if groupLogIDPack.LeftNum >= leftNum {
			// 越剩越多则报错
			log.Errorln("left num more")
			stream.Reset()
			return
		}

		leftNum = groupLogIDPack.LeftNum
	}

	if len(groupLogIDs) == 0 {
		log.Debugln("group log ids empty")
		return
	}

	// 发送日志给对方
	groupLogs, err := a.data.GetLogsByIDs(groupID, groupLogIDs)
	if err != nil {
		log.Errorf("data.GetLogsByIDs error: %w", err)
		stream.Reset()
		return
	}

	logNum := len(groupLogs)
	for i, groupLog := range groupLogs {
		if err := wt.WriteMsg(&pb.GroupLogPack{
			LeftNum: int32(logNum - i - 1),
			Log:     groupLog,
		}); err != nil {
			log.Errorf("pbio write log pack error: %w", err)
			return
		}
	}
}

func (m *AdminProto) syncHandler(stream network.Stream) {
	remotePeerID := stream.Conn().RemotePeer()
	defer stream.Close()

	// 获取同步GroupID
	var syncmsg pb.GroupSyncLog
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
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
	err = m.loopSync(groupID, remotePeerID, stream, rd, wt)
	if err != nil {
		log.Errorf("loop sync error: %v", err)
	}
}

func (a *AdminProto) handleSubscribe(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}

			switch evt := e.(type) {

			case myevent.EvtGroupNetworkSuccess:
				// 组网成功
				avatar, err := a.data.GetAvatar(ctx, evt.GroupID)
				if err != nil {
					log.Errorf("data get group avatar error: %v", err)

				} else if len(avatar) > 0 {
					if err = a.emitters.evtSyncResource.Emit(myevent.EvtSyncResource{
						ResourceID: avatar,
						PeerIDs:    evt.PeerIDs,
					}); err != nil {
						log.Errorf("emite sync resource event error: %w", err)
					}
				}

				depositAddress, err := a.data.GetDepositAddress(ctx, evt.GroupID)
				if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
					log.Errorf("data get group deposit address error: %v", err)

				} else if depositAddress != peer.ID("") {
					if err = a.emitters.evtSyncGroupMessage.Emit(myevent.EvtSyncGroupMessage{
						GroupID:        evt.GroupID,
						DepositAddress: depositAddress,
					}); err != nil {
						log.Errorf("emite sync group message event error: %w", err)
					}
				}

			case myevent.EvtGroupConnectChange:
				// 建立连接
				if evt.IsConnected {
					if _, exists := a.groupConns[evt.GroupID]; !exists {
						a.groupConns[evt.GroupID] = make(map[peer.ID]struct{})
					}

					if _, exists := a.groupConns[evt.GroupID][evt.PeerID]; !exists {
						a.groupConns[evt.GroupID][evt.PeerID] = struct{}{}

						if a.host.ID() < evt.PeerID {
							// 小方主动发起同步
							go a.goSyncAdmin(evt.GroupID, evt.PeerID)
						}
					}
					log.Debugln("event connect peer: ", evt.PeerID.String())

				} else {
					log.Debugln("event disconnect peer: ", evt.PeerID.String())
					delete(a.groupConns[evt.GroupID], evt.PeerID)
				}

			case myevent.EvtPullGroupLog:
				go a.handlePullGroupLogEvent(evt.GroupID, evt.PeerID, evt.Result)

			}

		case <-ctx.Done():
			return
		}
	}
}

func (a *AdminProto) handlePullGroupLogEvent(groupID string, peerID peer.ID, resultCh chan<- error) {
	var resultErr error
	var ctx = context.Background()

	defer func() {
		resultCh <- resultErr
		close(resultCh)
	}()

	// 首先计算消息摘要（头尾md5）
	logIDs, err := a.data.GetLogIDs(ctx, groupID)
	if err != nil {
		resultErr = fmt.Errorf("data.GetLogIDs error: %w", err)
		return
	}

	var logHash []byte
	var headID, tailID string

	logNum := len(logIDs)
	if logNum > 0 {
		md5sum := md5.New()
		for _, logID := range logIDs {
			if _, err := md5sum.Write([]byte(logID)); err != nil {
				resultErr = fmt.Errorf("md5sum write error: %w", err)
				return
			}
		}
		headID = logIDs[0]
		tailID = logIDs[logNum-1]
		logHash = md5sum.Sum(nil)
	}

	stream, err := a.host.NewStream(network.WithUseTransient(ctx, ""), peerID, PULL_ID)
	if err != nil {
		resultErr = fmt.Errorf("host.NewStream error: %w", err)
		return
	}
	defer stream.Close()

	sendSumMsg := pb.GroupLogHash{
		GroupId: groupID,
		HeadId:  headID,
		TailId:  tailID,
		Length:  int32(logNum),
		Hash:    logHash,
	}

	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&sendSumMsg); err != nil {
		resultErr = fmt.Errorf("pbio write msg error: %w", err)
		return
	}

	// 接收对方消息摘要
	var recvSumMsg pb.GroupLogHash
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	if err := rd.ReadMsg(&recvSumMsg); err != nil {
		resultErr = fmt.Errorf("pbio read msg error: %w", err)
		return
	}

	if string(recvSumMsg.Hash) == string(sendSumMsg.Hash) {
		// 正常退出
		log.Debugln("summary hash equal")
		return
	}

	// 消息摘要不一致，则发送本地所有消息ID给对方
	for i := 0; i < logNum; i++ {
		if err := wt.WriteMsg(&pb.GroupLogIDPack{
			LogId:   logIDs[i],
			LeftNum: int32(logNum - i - 1),
		}); err != nil {
			resultErr = fmt.Errorf("pbio write logid pack error: %w", err)
			return
		}
	}

	// 接收对方消息列表，并更新本地
	var groupLogs []*pb.GroupLog
	var groupLogPack pb.GroupLogPack
	var leftNum int32
	for {
		groupLogPack.Reset()
		if err := rd.ReadMsg(&groupLogPack); err != nil {
			resultErr = fmt.Errorf("pbio read log pack error: %w", err)
			return
		}

		if groupLogPack.Log != nil {
			groupLogs = append(groupLogs, groupLogPack.Log)
		}

		if groupLogPack.LeftNum <= 0 {
			break

		} else if groupLogPack.LeftNum >= leftNum {
			resultErr = fmt.Errorf("left num more")
			stream.Reset()
			return
		}

		leftNum = groupLogPack.LeftNum
	}

	for _, pblog := range groupLogs {
		if _, err := a.saveLog(ctx, pblog); err != nil {
			resultErr = fmt.Errorf("save log error: %w", err)
			stream.Reset()
			return
		}
	}
}

// 发送消息
func (a *AdminProto) broadcastMessage(groupID string, msg *pb.GroupLog, excludePeerIDs ...peer.ID) error {

	if len(a.groupConns[groupID]) == 0 {
		// 这里不要报错，报错也不能完全解决问题（比如：不退群直接把软件删除了呢），可以靠关联日志同步来弥补
		return nil
	}

	ctx := context.Background()

	for peerID := range a.groupConns[groupID] {
		if len(excludePeerIDs) > 0 {
			isExcluded := false
			for _, excludePeerID := range excludePeerIDs {
				if peerID == excludePeerID {
					isExcluded = true
					break
				}
			}
			if isExcluded {
				continue
			}
		}

		go a.sendPeerMessage(ctx, peerID, msg)
	}

	return nil
}

func (a *AdminProto) sendPeerMessage(ctx context.Context, peerID peer.ID, msg *pb.GroupLog) {

	stream, err := a.host.NewStream(network.WithUseTransient(ctx, ""), peerID, LOG_ID)
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

func (a *AdminProto) emitGroupMemberChange(ctx context.Context, groupID string) error {

	if err := a.emitters.evtGroupMemberChange.Emit(myevent.EvtGroupMemberChange{
		GroupID: groupID,
	}); err != nil {
		return fmt.Errorf("emit group member change error: %w", err)
	}

	return nil
}
