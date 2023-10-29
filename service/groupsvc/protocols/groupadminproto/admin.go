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

var log = logging.Logger("cuckoo/grougadminproto")

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

	sub, err := eventBus.Subscribe([]any{
		new(myevent.EvtHostBootComplete), new(myevent.EvtGroupNetworkSuccess),
		new(myevent.EvtGroupConnectChange), new(myevent.EvtPullGroupLog),
	})

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
	log.Debug("pullHandler: ")

	defer stream.Close()

	remotePeerID := stream.Conn().RemotePeer()
	var ctx = context.Background()

	// 接收对方消息摘要
	var recvSumMsg pb.GroupLogHash
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	if err := rd.ReadMsg(&recvSumMsg); err != nil {
		log.Errorf("pbio read msg error: %v", err)
		return
	}

	log.Debugln("recv sum msg: ", recvSumMsg.String())

	groupID := recvSumMsg.GroupId

	// 判断是否是群成员
	agreePeerIDs, err := a.GetAgreePeerIDs(ctx, groupID)
	if err != nil {
		log.Errorf("get agree peer ids error: %w", err)
		stream.Reset()
		return
	}

	isAgreePeer := false
	for _, peerID := range agreePeerIDs {
		if remotePeerID == peerID {
			isAgreePeer = true
			break
		}
	}

	if !isAgreePeer {
		log.Warnln("not accept peer, so refuse pull request")
		stream.Reset()
		return
	}

	// 首先计算消息摘要（头尾md5）
	logIDs, err := a.data.GetLogIDs(ctx, groupID)
	if err != nil {
		log.Errorf("data.GetLogIDs error: %v", err)
		stream.Reset()
		return
	}

	log.Debugln("data.GetLogIDs len ", len(logIDs))

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

	log.Debugln("pbio write sendSumMsg: ", sendSumMsg.String())

	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&sendSumMsg); err != nil {
		log.Errorf("pbio write msg error: %w", err)
		return
	}

	// 如果本地和远端都一样，则正常退出
	if string(recvSumMsg.Hash) == string(sendSumMsg.Hash) {
		log.Debugln("summary hash equal, return")
		return
	}

	// 不一致，准备接收对方的消息ID列表，然后进行比对
	var logIDPack pb.GroupLogIDPack
	var leftNum int32
	recvLogIDs := make(map[string]struct{})
	for {
		logIDPack.Reset()
		if err := rd.ReadMsg(&logIDPack); err != nil {
			log.Errorf("pbio read log pack error: %w", err)
			return
		}

		log.Debugf("pbio read group log id: %s, left: %d", logIDPack.LogId, logIDPack.LeftNum)

		if logIDPack.LogId != "" {
			recvLogIDs[logIDPack.LogId] = struct{}{}
		}

		if logIDPack.LeftNum <= 0 {
			// 无剩余则退出
			break

		} else if leftNum != 0 && logIDPack.LeftNum >= leftNum {
			// 越剩越多则报错
			log.Errorln("left num more get: %d, cur: %d", logIDPack.LeftNum, leftNum)
			stream.Reset()
			return
		}

		leftNum = logIDPack.LeftNum
	}

	if len(recvLogIDs) == 0 {
		log.Debugln("group log ids empty, return")
		return
	}

	var sendLogIDs []string
	for _, logID := range logIDs {
		if _, exists := recvLogIDs[logID]; !exists {
			sendLogIDs = append(sendLogIDs, logID)
		}
	}

	// 发送日志给对方
	sendLogs, err := a.data.GetLogsByIDs(groupID, sendLogIDs)
	if err != nil {
		log.Errorf("data.GetLogsByIDs error: %v", err)
		stream.Reset()
		return
	}

	logNum := len(sendLogs)
	for i, groupLog := range sendLogs {
		log.Debugln("pbio write group log pack ", i)
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

			case myevent.EvtHostBootComplete:
				if evt.IsSucc {
					go a.handleHostBootCompleteEvent()
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

func (a *AdminProto) handleHostBootCompleteEvent() {
	log.Debugln("handleHostBootCompleteEvent: ")

	ctx := context.Background()

	groupIDs, err := a.GetGroupIDs(ctx)
	if err != nil {
		log.Errorf("data.GetGroupIDs error: %w", err)
		return
	}

	var normalGroupIDs []string
	for _, groupID := range groupIDs {
		state, err := a.data.GetState(ctx, groupID)
		if err != nil {
			log.Errorf("data.GetState error: %w", err)
			return
		}

		switch state {
		case mytype.GroupStateDisband:
			creator, err := a.data.GetCreator(ctx, groupID)
			if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
				log.Errorf("data.GetCreate error: %w", err)
			}
			if a.host.ID() == creator {
				normalGroupIDs = append(normalGroupIDs, groupID)
			}
		case mytype.GroupStateNormal:
			normalGroupIDs = append(normalGroupIDs, groupID)
		}
	}

	for _, groupID := range normalGroupIDs {
		depositAddress, err := a.data.GetDepositAddress(ctx, groupID)
		if err != nil && !errors.Is(err, ipfsds.ErrNotFound) {
			log.Errorf("data get group deposit address error: %v", err)

		} else if depositAddress != peer.ID("") {
			if err = a.emitters.evtSyncGroupMessage.Emit(myevent.EvtSyncGroupMessage{
				GroupID:        groupID,
				DepositAddress: depositAddress,
			}); err != nil {
				log.Errorf("emite sync group message event error: %w", err)
			}
		}
	}
}

func (a *AdminProto) handlePullGroupLogEvent(groupID string, peerID peer.ID, resultCh chan<- error) {
	log.Debugln("handlePullGroupLogEvent: ")

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

	log.Debugln("pbio send sum msg ", sendSumMsg.String())
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

	log.Debugln("receive sum msg ", recvSumMsg.String())

	if string(recvSumMsg.Hash) == string(sendSumMsg.Hash) {
		log.Debugln("summary hash equal, return")
		return
	}

	// 消息摘要不一致，则发送本地所有消息ID给对方
	for i := 0; i < logNum; i++ {
		log.Debugln("pbio write group log id ", i)
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

		} else if leftNum != 0 && groupLogPack.LeftNum >= leftNum {
			resultErr = fmt.Errorf("left num more left: %d, cur: %d", groupLogPack.LeftNum, leftNum)
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

	log.Debugln("emitGroupMemberChange: ", groupID)

	if err := a.emitters.evtGroupMemberChange.Emit(myevent.EvtGroupMemberChange{
		GroupID: groupID,
	}); err != nil {
		return fmt.Errorf("emit group member change error: %w", err)
	}

	return nil
}
