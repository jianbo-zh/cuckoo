package sync

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ipfs/go-datastore/query"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/service/group/datastore"
	msgpb "github.com/jianbo-zh/dchat/service/group/protocol/message/pb"
	"github.com/jianbo-zh/dchat/service/group/protocol/sync/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/libp2p/go-msgio/pbio"
	"google.golang.org/protobuf/proto"
)

// 同步群消息日志

// 同步开始时，双方发送最小最大Lamport时钟，A(a1,a2) <-> B(b1,b2)
// 双方中最小Lamport时钟最大的一方，作为「同步」主动方，另一方作为「同步」被动方 （假设： a1 < b1 < b2 < a2）
// 双方中最大Lamport时钟最大的一方，作为最新消息方
// 最新消息方，主动发送最新差异的消息给对方，如：A(b2,a2] -> B
// 「同步」主动方B计算，B[b1,b2] 消息ID的hash值，发送给被动方A，A接收到后，也计算A[b1,b2] hash 值，并返回给B
// B比较2个hash是否相同，不相同，则同步 [b1,b2]区间的消息（发送消息ID列表给A），A收到消息ID列表后，比较是否有不同点（对方新的就保存，自己新的则主动转发给对方）
// B以此循环最终同步完成

// 1，发送整体范围
// 2，发送区间HASH
// 3，发送区间消息KEY
// 4，发送对方没有的消息ID

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/sync.proto=./pb pb/sync.proto

var log = logging.Logger("message")

var StreamTimeout = 1 * time.Minute

const (
	ID = "/dchat/group/syncmsg/1.0.0"

	ServiceName = "group.syncmsg"
	maxMsgSize  = 4 * 1024 // 4K
)

type GroupSyncMsgService struct {
	host host.Host

	datastore datastore.GroupIface
}

func NewGroupSyncMsgService(h host.Host, ds datastore.GroupIface, eventBus event.Bus) *GroupSyncMsgService {
	msgsyncsvc := GroupSyncMsgService{
		host:      h,
		datastore: ds,
	}

	h.SetStreamHandler(ID, msgsyncsvc.Handler)

	sub, err := eventBus.Subscribe([]any{new(gevent.EvtGroupPeerConnectChange)}, eventbus.Name("syncmsg"))
	if err != nil {
		log.Warnf("subscription failed. group admin server error: %v", err)

	} else {
		go msgsyncsvc.handleSubscribe(context.Background(), sub)
	}

	return &msgsyncsvc
}

func (syncsvc *GroupSyncMsgService) handleSubscribe(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}
			ev := e.(gevent.EvtGroupPeerConnectChange)
			// 接收消息日志
			// err := grp.datastore.LogAdminOperation(ctx, grp.host.ID(), datastore.GroupID(ev.MsgData.GroupId), ev.MsgData)
			var peerID peer.ID
			if syncsvc.host.ID().String() == ev.MsgData.PeerIdA {
				peerID, _ = peer.Decode(ev.MsgData.PeerIdB)
			} else {
				peerID, _ = peer.Decode(ev.MsgData.PeerIdA)
			}
			err := syncsvc.sync(ctx, ev.MsgData.GroupId, peerID)
			if err != nil {
				log.Errorf("log admin operation error: %v", err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (syncsvc *GroupSyncMsgService) sync(ctx context.Context, groupID string, peerID peer.ID) error {
	stream, err := syncsvc.host.NewStream(ctx, peerID, ID)
	if err != nil {
		return nil
	}

	totalmsg, err := syncsvc.totalMsg(groupID)
	if err != nil {
		return err
	}

	pbtotalmsg := pb.DataTotalLen{
		GroupId: groupID,
		HeadId:  totalmsg.HeadId,
		TailId:  totalmsg.TailId,
		Length:  totalmsg.Length,
	}

	bs, err := proto.Marshal(&pbtotalmsg)
	if err != nil {
		return err
	}

	syncmsg := pb.SyncMessage{
		GroupId: groupID,
		Step:    pb.SyncMessage_TOTAL_LEN,
		Payload: bs,
	}

	wb := pbio.NewDelimitedWriter(stream)
	err = wb.WriteMsg(&syncmsg)
	if err != nil {
		return err
	}

	for {
		// todo

	}

	return nil
}

func (syncsvc *GroupSyncMsgService) Handler(s network.Stream) {

	err := s.SetReadDeadline(time.Now().Add(StreamTimeout))
	if err != nil {
		log.Errorf("set read deadline error: %v", err)
		return
	}

	rb := pbio.NewDelimitedReader(s, maxMsgSize)
	var msg1 pb.SyncMessage
	if err := rb.ReadMsg(&msg1); err != nil {
		log.Errorf("read pb msg error: %v", err)
		return
	}
	s.SetReadDeadline(time.Time{})

	switch msg1.Step {
	case pb.SyncMessage_TOTAL_LEN:

		var dataTotal pb.DataTotalLen
		if err := proto.Unmarshal(msg1.Payload, &dataTotal); err != nil {
			log.Errorf("unmarshal pb data total len error: %v", err)
			return
		}

		// 同样返回total len
		dataTotal2, err := syncsvc.totalMsg(msg1.GroupId)
		if err != nil {
			log.Errorf("get group total msg error: %v", err)
			return
		}
		dtbs, err := proto.Marshal(dataTotal2)
		if err != nil {
			log.Errorf("marshal data total error: %v", err)
			return
		}

		wb := pbio.NewDelimitedWriter(s)
		err = wb.WriteMsg(&pb.SyncMessage{
			GroupId: msg1.GroupId,
			Payload: dtbs,
		})
		if err != nil {
			log.Errorf("send group total msg error: %v", err)
			return
		}

		if dataTotal2.TailId > dataTotal.TailId {
			// 如果有最新的数据则发送给对方
			pushmsg, err := syncsvc.getRangeMessages(dataTotal.GroupId, dataTotal.TailId, dataTotal2.TailId)
			if err != nil {
				log.Errorf("get range messages error: %v", err)
				return
			}

			bs, err := proto.Marshal(pushmsg)
			if err != nil {
				log.Errorf("marshal push msg error: %v", err)
				return
			}

			syncmsg := pb.SyncMessage{
				GroupId: msg1.GroupId,
				Step:    pb.SyncMessage_PUSH_MSG,
				Payload: bs,
			}

			err = wb.WriteMsg(&syncmsg)
			if err != nil {
				return
			}
		}

		if dataTotal2.HeadId > dataTotal.HeadId {
			// 当前数据要少些，则作为同步主动方
			startID := dataTotal.HeadId
			endID := dataTotal.TailId
			if dataTotal2.TailId < dataTotal.TailId {
				endID = dataTotal2.TailId
			}

			hash, err := syncsvc.rangeHash(msg1.GroupId, startID, endID)
			if err != nil {
				log.Errorf("range hash error: %v", err)
				return
			}

			bs, err := proto.Marshal(&pb.DataRangeHash{
				GroupId: msg1.GroupId,
				StartId: startID,
				EndId:   endID,
				Hash:    hash,
			})
			if err != nil {
				log.Errorf("pb marshal error: %v", err)
				return
			}

			err = wb.WriteMsg(&pb.SyncMessage{
				GroupId: msg1.GroupId,
				Step:    pb.SyncMessage_RANGE_HASH,
				Payload: bs,
			})
			if err != nil {
				log.Errorf("write msg error: %v", err)
				return
			}
		}

	case pb.SyncMessage_RANGE_HASH:
		// 收到计算hash

		var hashmsg pb.DataRangeHash
		if err := proto.Unmarshal(msg1.Payload, &hashmsg); err != nil {
			log.Errorf("unmarshal pb data total len error: %v", err)
			return
		}

		// 我也计算hash
		hashBs, err := syncsvc.rangeHash(msg1.GroupId, hashmsg.StartId, hashmsg.EndId)
		if err != nil {
			log.Errorf("range hash error: %v", err)
		}

		if !bytes.Equal(hashmsg.Hash, hashBs) {
			// hash 不同，则消息不一致，则同步消息ID
			hashidsmsg, err := syncsvc.getRangeMsgIDs(msg1.GroupId, hashmsg.StartId, hashmsg.EndId)
			if err != nil {
				log.Errorf("get range msg ids error: %v", err)
				return
			}

			bs, err := proto.Marshal(hashidsmsg)
			if err != nil {
				log.Errorf("marshal hash id message error: %v", err)
				return
			}

			syncmsg := pb.SyncMessage{
				GroupId: msg1.GroupId,
				Step:    pb.SyncMessage_RANGE_IDS,
				Payload: bs,
			}

			wb := pbio.NewDelimitedWriter(s)
			if err = wb.WriteMsg(&syncmsg); err != nil {
				log.Errorf("send group total msg error: %v", err)
				return
			}

		}

	case pb.SyncMessage_RANGE_IDS:
		var idmsg pb.DataRangeIDs
		if err := proto.Unmarshal(msg1.Payload, &idmsg); err != nil {
			log.Errorf("unmarshal pb data total len error: %v", err)
			return
		}

		hashidsmsg, err := syncsvc.getRangeMsgIDs(idmsg.GroupId, idmsg.StartId, idmsg.EndId)
		if err != nil {
			log.Errorf("get range message ids error: %v", err)
			return
		}

		// 比较不同点
		var moreIDs []string
		mapIds1 := make(map[string]struct{}, len(idmsg.Ids))
		for _, id := range idmsg.Ids {
			mapIds1[id] = struct{}{}
		}
		for _, id := range hashidsmsg.Ids {
			if _, exists := mapIds1[id]; !exists {
				moreIDs = append(moreIDs, id)
			}
		}

		if len(moreIDs) > 0 {
			pushmsg, err := syncsvc.getIDMessages(idmsg.GroupId, moreIDs)
			if err != nil {
				log.Errorf("get messages by ids error: %v", err)
				return
			}

			bs, err := proto.Marshal(pushmsg)
			if err != nil {
				log.Errorf("marshal push message error: %v", err)
				return
			}

			syncmsg := pb.SyncMessage{
				GroupId: idmsg.GroupId,
				Step:    pb.SyncMessage_PUSH_MSG,
				Payload: bs,
			}
			wb := pbio.NewDelimitedWriter(s)
			if err = wb.WriteMsg(&syncmsg); err != nil {
				log.Errorf("send group total msg error: %v", err)
				return
			}
		}

	case pb.SyncMessage_PUSH_MSG:
		var msgs pb.DataPushMsgs
		if err := proto.Unmarshal(msg1.Payload, &msgs); err != nil {
			log.Errorf("unmarshal pb data total len error: %v", err)
			return
		}

		for _, bs := range msgs.Msgs {
			var msg msgpb.GroupMsg
			if err = proto.Unmarshal(bs, &msg); err != nil {
				log.Errorf("unmarshal error: %v", err)
				return
			}

			err = syncsvc.datastore.PutMessage(context.Background(), datastore.GroupID(msg1.GroupId), &msg)
			if err != nil {
				log.Errorf("put message error: %v", err)
				return
			}
		}

	default:
		// no defined
	}
}

func (syncsvc *GroupSyncMsgService) totalMsg(groupID string) (*pb.DataTotalLen, error) {
	ctx := context.Background()

	// headID
	headID, err := syncsvc.datastore.GetMessageHeadID(ctx, datastore.GroupID(groupID))
	if err != nil {
		return nil, err
	}
	// tailID
	tailID, err := syncsvc.datastore.GetMessageTailID(ctx, datastore.GroupID(groupID))
	if err != nil {
		return nil, err
	}
	// len
	length, err := syncsvc.datastore.GetMessageLength(ctx, datastore.GroupID(groupID))
	if err != nil {
		return nil, err
	}

	return &pb.DataTotalLen{
		GroupId: groupID,
		HeadId:  headID,
		TailId:  tailID,
		Length:  length,
	}, nil
}

type RangeFilter struct {
	StartID      int64
	EndID        int64
	WithoutStart bool
}

func (rf *RangeFilter) Filter(e query.Entry) bool {
	arr := strings.Split(e.Key, "_")
	if len(arr) <= 1 {
		return false
	}

	id, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil {
		return false
	}

	if id >= rf.StartID && id <= rf.EndID {

		if rf.WithoutStart && id == rf.StartID {
			return false
		}

		return true
	}

	return false
}

func (syncsvc *GroupSyncMsgService) getRangeMsgIDs(groupID string, startIDStr string, endIDStr string) (*pb.DataRangeIDs, error) {

	startArr := strings.Split(startIDStr, "_")
	if len(startArr) <= 1 {
		return nil, fmt.Errorf("split start id error")
	}
	startID, _ := strconv.ParseInt(startArr[0], 10, 64)

	endArr := strings.Split(endIDStr, "_")
	if len(endArr) <= 1 {
		return nil, fmt.Errorf("split end id error")
	}
	endID, _ := strconv.ParseInt(endArr[0], 10, 64)

	results, err := syncsvc.datastore.Query(context.Background(), query.Query{
		Prefix:   "/dchat/group/" + groupID + "/message/logs/",
		Filters:  []query.Filter{&RangeFilter{StartID: startID, EndID: endID}},
		Orders:   []query.Order{query.OrderByKey{}},
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	var ids []string
	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		ids = append(ids, result.Entry.Key)
	}

	dataPushMsg := pb.DataRangeIDs{
		GroupId: groupID,
		StartId: startIDStr, // lamportID
		EndId:   endIDStr,
	}

	return &dataPushMsg, nil
}

func (syncsvc *GroupSyncMsgService) getRangeMessages(groupID string, startIDStr string, endIDStr string) (*pb.DataPushMsgs, error) {

	startArr := strings.Split(startIDStr, "_")
	if len(startArr) <= 1 {
		return nil, fmt.Errorf("split start id error")
	}
	startID, _ := strconv.ParseInt(startArr[0], 10, 64)

	endArr := strings.Split(endIDStr, "_")
	if len(endArr) <= 1 {
		return nil, fmt.Errorf("split end id error")
	}
	endID, _ := strconv.ParseInt(endArr[0], 10, 64)

	results, err := syncsvc.datastore.Query(context.Background(), query.Query{
		Prefix:  "/dchat/group/" + groupID + "/message/logs/",
		Filters: []query.Filter{&RangeFilter{StartID: startID, EndID: endID}},
		Orders:  []query.Order{query.OrderByKey{}},
	})
	if err != nil {
		return nil, err
	}

	var msgs [][]byte
	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		msgs = append(msgs, result.Entry.Value)
	}

	dataPushMsg := pb.DataPushMsgs{
		Msgs: msgs,
	}

	return &dataPushMsg, nil
}

func (syncsvc *GroupSyncMsgService) getIDMessages(groupID string, msgIDs []string) (*pb.DataPushMsgs, error) {

	var bss [][]byte
	for _, msgID := range msgIDs {
		msg, err := syncsvc.datastore.GetMessage(context.Background(), datastore.GroupID(groupID), msgID)
		if err != nil {
			return nil, err
		}

		bs, err := proto.Marshal(msg)
		if err != nil {
			return nil, err
		}

		bss = append(bss, bs)
	}

	pushmsg := pb.DataPushMsgs{
		Msgs: bss,
	}

	return &pushmsg, nil
}

func (syncsvc *GroupSyncMsgService) rangeHash(groupID string, startIDStr string, endIDStr string) ([]byte, error) {

	startArr := strings.Split(startIDStr, "_")
	if len(startArr) <= 1 {
		return nil, fmt.Errorf("split start id error")
	}
	startID, _ := strconv.ParseInt(startArr[0], 10, 64)

	endArr := strings.Split(endIDStr, "_")
	if len(endArr) <= 1 {
		return nil, fmt.Errorf("split end id error")
	}
	endID, _ := strconv.ParseInt(endArr[0], 10, 64)

	results, err := syncsvc.datastore.Query(context.Background(), query.Query{
		Prefix:   "/dchat/group/" + groupID + "/message/logs/",
		Filters:  []query.Filter{&RangeFilter{StartID: startID, EndID: endID}},
		Orders:   []query.Order{query.OrderByKey{}},
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	hash := sha1.New()

	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		if _, err = hash.Write([]byte(result.Entry.Key)); err != nil {
			return nil, err
		}
	}

	return hash.Sum(nil), nil
}
