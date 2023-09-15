package adminproto

import (
	"bytes"
	"context"
	"crypto/sha1"
	"time"

	"github.com/jianbo-zh/dchat/service/groupsvc/protocol/adminproto/pb"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
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

func (a *AdminProto) sync(groupID string, peerID peer.ID) {

	stream, err := a.host.NewStream(context.Background(), peerID, SYNC_ID)
	if err != nil {
		return
	}
	defer stream.Close()

	wt := pbio.NewDelimitedWriter(stream)
	if err = wt.WriteMsg(&pb.SyncLog{
		Type:    pb.SyncLog_INIT,
		GroupId: groupID,
	}); err != nil {
		return
	}

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)

	err = a.loopSync(groupID, stream, rd, wt)
	if err != nil {
		log.Errorf("loop sync error: %v", err)
	}
}

func (a *AdminProto) loopSync(groupID string, stream network.Stream, rd pbio.ReadCloser, wt pbio.WriteCloser) error {

	var syncmsg pb.SyncLog

	for {
		syncmsg.Reset()

		// 设置读取超时，
		stream.SetReadDeadline(time.Now().Add(5 * StreamTimeout))
		if err := rd.ReadMsg(&syncmsg); err != nil {
			return err
		}
		stream.SetReadDeadline(time.Time{})

		switch syncmsg.Type {
		case pb.SyncLog_SUMMARY:
			if err := a.handleSyncSummary(groupID, &syncmsg, wt); err != nil {
				return err
			}

		case pb.SyncLog_RANGE_HASH:
			if err := a.handleSyncRangeHash(groupID, &syncmsg, wt); err != nil {
				return err
			}

		case pb.SyncLog_RANGE_IDS:
			if err := a.handleSyncRangeIDs(groupID, &syncmsg, wt); err != nil {
				return err
			}

		case pb.SyncLog_PUSH_MSG:
			if err := a.handleSyncPushMsg(groupID, &syncmsg); err != nil {
				return err
			}

		case pb.SyncLog_PULL_MSG:
			if err := a.handleSyncPullMsg(groupID, &syncmsg, wt); err != nil {
				return err
			}

		case pb.SyncLog_DONE:
			return nil

		default:
			// no defined
		}
	}
}

func (a *AdminProto) handleSyncSummary(groupID string, syncmsg *pb.SyncLog, wt pbio.WriteCloser) error {

	var remoteSummary pb.DataSummary
	if err := proto.Unmarshal(syncmsg.Payload, &remoteSummary); err != nil {
		return err
	}

	err := a.data.MergeLamportTime(context.Background(), groupID, remoteSummary.Lamptime)
	if err != nil {
		return err
	}

	localSummary, err := a.getMessageSummary(groupID)
	if err != nil {
		return err
	}

	if !remoteSummary.IsEnd {
		// 结束标记，避免无限循环
		localSummary.IsEnd = true
		payload, err := proto.Marshal(localSummary)
		if err != nil {
			return err
		}

		if err = wt.WriteMsg(&pb.SyncLog{
			Type:    pb.SyncLog_SUMMARY,
			Payload: payload,
		}); err != nil {
			return err
		}
	}

	if localSummary.TailId > remoteSummary.TailId {
		// 如果有最新的数据则发送给对方
		msgs, err := a.getRangeMessages(groupID, remoteSummary.TailId, localSummary.TailId)
		if err != nil {
			return err
		}

		for _, msg := range msgs {
			bs, err := proto.Marshal(msg)
			if err != nil {
				return err
			}

			if err = wt.WriteMsg(&pb.SyncLog{
				Type:    pb.SyncLog_PUSH_MSG,
				Payload: bs,
			}); err != nil {
				return err
			}
		}
	}

	if localSummary.HeadId > remoteSummary.HeadId {
		// 当前数据要少些，则作为同步主动方
		startID := localSummary.HeadId
		endID := remoteSummary.TailId

		if localSummary.TailId < remoteSummary.TailId {
			endID = localSummary.TailId
		}

		hash, err := a.rangeHash(groupID, startID, endID)
		if err != nil {
			return err
		}

		bs, err := proto.Marshal(&pb.DataRangeHash{
			StartId: startID,
			EndId:   endID,
			Hash:    hash,
		})
		if err != nil {
			return err
		}

		if err = wt.WriteMsg(&pb.SyncLog{
			Type:    pb.SyncLog_RANGE_HASH,
			Payload: bs,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (a *AdminProto) handleSyncRangeHash(groupID string, syncmsg *pb.SyncLog, wt pbio.WriteCloser) error {
	var hashmsg pb.DataRangeHash
	if err := proto.Unmarshal(syncmsg.Payload, &hashmsg); err != nil {
		return err
	}

	// 我也计算hash
	hashBytes, err := a.rangeHash(groupID, hashmsg.StartId, hashmsg.EndId)
	if err != nil {
		log.Errorf("range hash error: %v", err)
	}

	if bytes.Equal(hashmsg.Hash, hashBytes) {
		// hash相同，不需要再同步了
		return wt.WriteMsg(&pb.SyncLog{Type: pb.SyncLog_DONE})
	}

	// hash 不同，则消息不一致，则同步消息ID
	hashids, err := a.getRangeIDs(groupID, hashmsg.StartId, hashmsg.EndId)
	if err != nil {
		return err
	}

	bs, err := proto.Marshal(hashids)
	if err != nil {
		return err
	}

	if err = wt.WriteMsg(&pb.SyncLog{
		Type:    pb.SyncLog_RANGE_IDS,
		Payload: bs,
	}); err != nil {
		return err
	}

	return nil
}

func (a *AdminProto) handleSyncRangeIDs(groupID string, syncmsg *pb.SyncLog, wt pbio.WriteCloser) error {
	var idmsg pb.DataRangeIDs
	if err := proto.Unmarshal(syncmsg.Payload, &idmsg); err != nil {
		return err
	}

	idmsg2, err := a.getRangeIDs(groupID, idmsg.StartId, idmsg.EndId)
	if err != nil {
		return err
	}

	// 比较不同点
	mapIds := make(map[string]struct{}, len(idmsg.Ids))
	mapIds2 := make(map[string]struct{}, len(idmsg.Ids))

	for _, id := range idmsg.Ids {
		mapIds[id] = struct{}{}
	}

	var moreIDs []string
	for _, id := range idmsg2.Ids {
		mapIds2[id] = struct{}{}
		if _, exists := mapIds[id]; !exists {
			moreIDs = append(moreIDs, id)
		}
	}

	if len(moreIDs) > 0 {
		msgs, err := a.getMessagesByIDs(groupID, moreIDs)
		if err != nil {
			return err
		}

		for _, msg := range msgs {
			bs, err := proto.Marshal(msg)
			if err != nil {
				return err
			}

			if err = wt.WriteMsg(&pb.SyncLog{
				Type:    pb.SyncLog_PUSH_MSG,
				Payload: bs,
			}); err != nil {
				return err
			}
		}
	}

	var lessIDs []string
	for _, id := range idmsg.Ids {
		if _, exists := mapIds2[id]; !exists {
			lessIDs = append(lessIDs, id)
		}
	}

	if len(lessIDs) > 0 {
		bs, err := proto.Marshal(&pb.DataPullMsg{
			Ids: lessIDs,
		})
		if err != nil {
			return err
		}

		if err = wt.WriteMsg(&pb.SyncLog{
			Type:    pb.SyncLog_PULL_MSG,
			Payload: bs,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (a *AdminProto) handleSyncPushMsg(groupID string, syncmsg *pb.SyncLog) error {
	var msg pb.Log
	if err := proto.Unmarshal(syncmsg.Payload, &msg); err != nil {
		return err
	}

	if err := a.data.SaveLog(context.Background(), a.host.ID(), groupID, &msg); err != nil {
		return err
	}
	return nil
}

func (a *AdminProto) handleSyncPullMsg(groupID string, syncmsg *pb.SyncLog, wt pbio.WriteCloser) error {
	var pullmsg pb.DataPullMsg
	if err := proto.Unmarshal(syncmsg.Payload, &pullmsg); err != nil {
		return err
	}

	msgs, err := a.data.GetMessagesByIDs(groupID, pullmsg.Ids)
	if err != nil {
		return err
	}

	for _, msg := range msgs {
		bs, err := proto.Marshal(msg)
		if err != nil {
			return err
		}

		if err = wt.WriteMsg(&pb.SyncLog{
			Type:    pb.SyncLog_PUSH_MSG,
			Payload: bs,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (a *AdminProto) getMessageSummary(groupID string) (*pb.DataSummary, error) {

	ctx := context.Background()

	// headID
	headID, err := a.data.GetMessageHead(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// tailID
	tailID, err := a.data.GetMessageTail(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// len
	length, err := a.data.GetMessageLength(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// lamport time
	lamptime, err := a.data.GetLamportTime(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return &pb.DataSummary{
		HeadId:   headID,
		TailId:   tailID,
		Length:   length,
		Lamptime: lamptime,
	}, nil
}

func (a *AdminProto) getRangeMessages(groupID string, startID string, endID string) ([]*pb.Log, error) {
	return a.data.GetRangeMessages(groupID, startID, endID)
}

func (a *AdminProto) rangeHash(groupID string, startID string, endID string) ([]byte, error) {
	ids, err := a.data.GetRangeIDs(groupID, startID, endID)
	if err != nil {
		return nil, err
	}
	hash := sha1.New()
	for _, id := range ids {
		if _, err = hash.Write([]byte(id)); err != nil {
			return nil, err
		}
	}

	return hash.Sum(nil), nil
}

func (a *AdminProto) getRangeIDs(groupID string, startID string, endID string) (*pb.DataRangeIDs, error) {
	ids, err := a.data.GetRangeIDs(groupID, startID, endID)
	if err != nil {
		return nil, err
	}

	return &pb.DataRangeIDs{
		StartId: startID,
		EndId:   endID,
		Ids:     ids,
	}, nil
}

func (a *AdminProto) getMessagesByIDs(groupID string, msgIDs []string) ([]*pb.Log, error) {
	return a.data.GetMessagesByIDs(groupID, msgIDs)
}
