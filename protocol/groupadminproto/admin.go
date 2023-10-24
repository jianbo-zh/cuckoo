package groupadminproto

// 群管理相关协议

import (
	"context"
	"errors"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	ds "github.com/jianbo-zh/dchat/datastore/ds/groupadminds"
	"github.com/jianbo-zh/dchat/datastore/ds/sessionds"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/myprotocol"
	"github.com/jianbo-zh/dchat/internal/mytype"
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
	LOG_ID  = myprotocol.GroupAdminID_v100
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

	sub, err := eventBus.Subscribe([]any{new(myevent.EvtGroupNetworkSuccess), new(myevent.EvtGroupConnectChange)}, eventbus.Name("adminlog"))
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

			}

		case <-ctx.Done():
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

	stream, err := a.host.NewStream(network.WithDialPeerTimeout(ctx, mytype.DialTimeout), peerID, LOG_ID)
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

	memberIDs, err := a.data.GetMemberIDs(ctx, groupID)
	if err != nil {
		return fmt.Errorf("data.GetMemberIDs error: %w", err)
	}
	agreePeerIDs, err := a.data.GetAgreePeerIDs(ctx, groupID)
	if err != nil {
		return fmt.Errorf("data.GetAgreePeerIDs error: %w", err)
	}
	refusePeerLogs, err := a.data.GetRefusePeerLogs(ctx, groupID)
	if err != nil {
		return fmt.Errorf("data.GetRefusePeerLogs error: %w", err)
	}

	if err := a.emitters.evtGroupMemberChange.Emit(myevent.EvtGroupMemberChange{
		GroupID:       groupID,
		PeerIDs:       memberIDs,
		AcptPeerIDs:   agreePeerIDs,
		RefusePeerIDs: refusePeerLogs,
	}); err != nil {
		return fmt.Errorf("emit group member change error: %w", err)
	}

	return nil
}
