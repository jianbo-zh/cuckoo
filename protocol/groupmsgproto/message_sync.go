package groupmsgproto

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/grouppb"
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

func (m *MessageProto) goSync(groupID string, peerID peer.ID) {

	log.Infoln("sync group msg log start")

	ctx := context.Background()
	stream, err := m.host.NewStream(network.WithDialPeerTimeout(ctx, mytype.DialTimeout), peerID, SYNC_ID)
	if err != nil {
		log.Errorf("host sync new stream error: %w", err)
		return
	}
	defer stream.Close()

	wt := pbio.NewDelimitedWriter(stream)
	if err = wt.WriteMsg(&pb.GroupSyncMessage{Type: pb.GroupSyncMessage_INIT, GroupId: groupID}); err != nil {
		log.Errorf("pbio write sync init error: %w", err)
		return
	}

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)

	err = m.loopSync(groupID, stream, rd, wt)
	if err != nil {
		log.Errorf("loop sync error: %v", err)
	}
}

func (m *MessageProto) loopSync(groupID string, stream network.Stream, rd pbio.ReadCloser, wt pbio.WriteCloser) error {

	var syncmsg pb.GroupSyncMessage

	for {
		syncmsg.Reset()

		// 设置读取超时，
		stream.SetReadDeadline(time.Now().Add(5 * StreamTimeout))
		if err := rd.ReadMsg(&syncmsg); err != nil {
			return fmt.Errorf("pbio read sync msg error: %w", err)
		}
		stream.SetReadDeadline(time.Time{})

		switch syncmsg.Type {
		case pb.GroupSyncMessage_SUMMARY:
			if err := m.handleSyncSummary(groupID, &syncmsg, wt); err != nil {
				return fmt.Errorf("handle sync summary error: %w", err)
			}

		case pb.GroupSyncMessage_RANGE_HASH:
			if err := m.handleSyncRangeHash(groupID, &syncmsg, wt); err != nil {
				return fmt.Errorf("handle sync range hash error: %w", err)
			}

		case pb.GroupSyncMessage_RANGE_IDS:
			if err := m.handleSyncRangeIDs(groupID, &syncmsg, wt); err != nil {
				return fmt.Errorf("handle sync range ids error: %w", err)
			}

		case pb.GroupSyncMessage_PUSH_MSG:
			if err := m.handleSyncPushMsg(groupID, &syncmsg); err != nil {
				return fmt.Errorf("handle sync push msg error: %w", err)
			}

		case pb.GroupSyncMessage_PULL_MSG:
			if err := m.handleSyncPullMsg(groupID, &syncmsg, wt); err != nil {
				return fmt.Errorf("handle sync pull msg error: %w", err)
			}

		case pb.GroupSyncMessage_DONE:
			log.Infoln("sync message done")
			return nil

		default:
			// no defined
			log.Warn("undefined sync msg")
		}
	}
}

func (m *MessageProto) handleSyncSummary(groupID string, syncmsg *pb.GroupSyncMessage, wt pbio.WriteCloser) error {

	var remoteSummary pb.GroupMessageSummary
	if err := proto.Unmarshal(syncmsg.Payload, &remoteSummary); err != nil {
		return fmt.Errorf("proto unmarshal msg payload error: %w", err)
	}

	err := m.data.MergeLamportTime(context.Background(), groupID, remoteSummary.Lamptime)
	if err != nil {
		return fmt.Errorf("merge lamptime error: %w", err)
	}

	localSummary, err := m.getMessageSummary(groupID)
	if err != nil {
		return fmt.Errorf("get msg summary error: %w", err)
	}

	if !remoteSummary.IsEnd {
		// 结束标记，避免无限循环
		localSummary.IsEnd = true
		payload, err := proto.Marshal(localSummary)
		if err != nil {
			return fmt.Errorf("proto marshal local summary error: %w", err)
		}

		if err = wt.WriteMsg(&pb.GroupSyncMessage{
			Type:    pb.GroupSyncMessage_SUMMARY,
			Payload: payload,
		}); err != nil {
			return fmt.Errorf("pbio write summary msg error: %w", err)
		}
	}

	if localSummary.TailId > remoteSummary.TailId {
		// 如果有最新的数据则发送给对方
		msgs, err := m.getRangeMessages(groupID, remoteSummary.TailId, localSummary.TailId)
		if err != nil {
			return fmt.Errorf("get range msg error: %w", err)
		}

		for _, msg := range msgs {
			bs, err := proto.Marshal(msg)
			if err != nil {
				return fmt.Errorf("proto marshal msg error: %w", err)
			}

			fmt.Println("---tail msg: ", msg.String())

			if err = wt.WriteMsg(&pb.GroupSyncMessage{
				Type:    pb.GroupSyncMessage_PUSH_MSG,
				Payload: bs,
			}); err != nil {
				return fmt.Errorf("pbio write push msg error: %w", err)
			}
		}
	}

	// 当前数据要少些，则作为同步主动方
	startID := localSummary.HeadId
	if remoteSummary.HeadId < localSummary.HeadId {
		startID = remoteSummary.HeadId
	}

	endID := remoteSummary.TailId
	hash, err := m.rangeHash(groupID, startID, endID)
	if err != nil {
		return fmt.Errorf("range hash error: %w", err)
	}

	bs, err := proto.Marshal(&pb.GroupMessageRangeHash{
		StartId: startID,
		EndId:   endID,
		Hash:    hash,
	})
	if err != nil {
		return fmt.Errorf("proto marshal range hash error: %w", err)
	}

	if err = wt.WriteMsg(&pb.GroupSyncMessage{
		Type:    pb.GroupSyncMessage_RANGE_HASH,
		Payload: bs,
	}); err != nil {
		return fmt.Errorf("pbio write range hash msg error: %w", err)
	}

	return nil
}

func (m *MessageProto) handleSyncRangeHash(groupID string, syncmsg *pb.GroupSyncMessage, wt pbio.WriteCloser) error {
	var hashmsg pb.GroupMessageRangeHash
	if err := proto.Unmarshal(syncmsg.Payload, &hashmsg); err != nil {
		return fmt.Errorf("proto unmarshal payload error: %w", err)
	}

	// 我也计算hash
	hashBytes, err := m.rangeHash(groupID, hashmsg.StartId, hashmsg.EndId)
	if err != nil {
		// log.Errorf("range hash error: %v", err)
		return fmt.Errorf("range hash error: %w", err)
	}

	if bytes.Equal(hashmsg.Hash, hashBytes) {
		// hash相同，不需要再同步了
		return nil
	}

	// hash 不同，则消息不一致，则同步消息ID
	hashids, err := m.getRangeIDs(groupID, hashmsg.StartId, hashmsg.EndId)
	if err != nil {
		return fmt.Errorf("get range ids error: %w", err)
	}

	bs, err := proto.Marshal(hashids)
	if err != nil {
		return fmt.Errorf("proto marshal hashids error: %w", err)
	}

	if err = wt.WriteMsg(&pb.GroupSyncMessage{
		Type:    pb.GroupSyncMessage_RANGE_IDS,
		Payload: bs,
	}); err != nil {
		return fmt.Errorf("pbio write range ids msg error: %w", err)
	}

	return nil
}

func (m *MessageProto) handleSyncRangeIDs(groupID string, syncmsg *pb.GroupSyncMessage, wt pbio.WriteCloser) error {
	var idmsg pb.GroupMessageRangeIDs
	if err := proto.Unmarshal(syncmsg.Payload, &idmsg); err != nil {
		return fmt.Errorf("proto unmarshal error: %w", err)
	}

	idmsg2, err := m.getRangeIDs(groupID, idmsg.StartId, idmsg.EndId)
	if err != nil {
		return fmt.Errorf("get range ids error: %w", err)
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
		msgs, err := m.getMessagesByIDs(groupID, moreIDs)
		if err != nil {
			return fmt.Errorf("get msgs by ids error: %w", err)
		}

		for _, msg := range msgs {
			bs, err := proto.Marshal(msg)
			if err != nil {
				return fmt.Errorf("proto marshal msg error: %w", err)
			}

			fmt.Println("---handle range ids msg: ", msg.String())

			if err = wt.WriteMsg(&pb.GroupSyncMessage{
				Type:    pb.GroupSyncMessage_PUSH_MSG,
				Payload: bs,
			}); err != nil {
				return fmt.Errorf("pbio write push msg error: %w", err)
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
		bs, err := proto.Marshal(&pb.GroupMessagePullMsg{
			Ids: lessIDs,
		})
		if err != nil {
			return fmt.Errorf("proto marshal pull msg error: %w", err)
		}

		if err = wt.WriteMsg(&pb.GroupSyncMessage{
			Type:    pb.GroupSyncMessage_PULL_MSG,
			Payload: bs,
		}); err != nil {
			return fmt.Errorf("pbio write pull msg error: %w", err)
		}
	}

	return nil
}

func (m *MessageProto) handleSyncPushMsg(groupID string, syncmsg *pb.GroupSyncMessage) error {
	var msg pb.GroupMessage
	if err := proto.Unmarshal(syncmsg.Payload, &msg); err != nil {
		return fmt.Errorf("proto unmarshal error: %w, len: %d", err, len(syncmsg.Payload))
	}

	if err := m.saveMessage(context.Background(), groupID, &msg); err != nil {
		return fmt.Errorf("data save msg error: %w", err)
	}

	return nil
}

func (m *MessageProto) handleSyncPullMsg(groupID string, syncmsg *pb.GroupSyncMessage, wt pbio.WriteCloser) error {
	var pullmsg pb.GroupMessagePullMsg
	if err := proto.Unmarshal(syncmsg.Payload, &pullmsg); err != nil {
		return fmt.Errorf("proto unmarshal msg payload error: %w", err)
	}

	msgs, err := m.data.GetMessagesByIDs(groupID, pullmsg.Ids)
	if err != nil {
		return fmt.Errorf("get msg by ids error: %w", err)
	}

	for _, msg := range msgs {
		bs, err := proto.Marshal(msg)
		if err != nil {
			return fmt.Errorf("proto marshal msg error: %w", err)
		}

		fmt.Println("---handle pull msg: ", msg.String())

		if err = wt.WriteMsg(&pb.GroupSyncMessage{
			Type:    pb.GroupSyncMessage_PUSH_MSG,
			Payload: bs,
		}); err != nil {
			return fmt.Errorf("pbio write push msg error: %w", err)
		}
	}

	return nil
}

func (m *MessageProto) getMessageSummary(groupID string) (*pb.GroupMessageSummary, error) {

	ctx := context.Background()

	// headID
	headID, err := m.data.GetMessageHead(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get msg head error: %w", err)
	}

	// tailID
	tailID, err := m.data.GetMessageTail(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get msg tail error: %w", err)
	}

	// len
	length, err := m.data.GetMessageLength(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get msg length error: %w", err)
	}

	// lamport time
	lamptime, err := m.data.GetLamportTime(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("data get lamptime error: %w", err)
	}

	return &pb.GroupMessageSummary{
		HeadId:   headID,
		TailId:   tailID,
		Length:   length,
		Lamptime: lamptime,
	}, nil
}

func (m *MessageProto) getRangeMessages(groupID string, startID string, endID string) ([]*pb.GroupMessage, error) {
	return m.data.GetRangeMessages(groupID, startID, endID)
}

func (m *MessageProto) rangeHash(groupID string, startID string, endID string) ([]byte, error) {
	ids, err := m.data.GetRangeIDs(groupID, startID, endID)
	if err != nil {
		return nil, fmt.Errorf("data get range ids error: %w", err)
	}
	hash := sha1.New()
	for _, id := range ids {
		if _, err = hash.Write([]byte(id)); err != nil {
			return nil, fmt.Errorf("hash write error: %w", err)
		}
	}

	return hash.Sum(nil), nil
}

func (m *MessageProto) getRangeIDs(groupID string, startID string, endID string) (*pb.GroupMessageRangeIDs, error) {
	ids, err := m.data.GetRangeIDs(groupID, startID, endID)
	if err != nil {
		return nil, fmt.Errorf("data get range ids error: %w", err)
	}

	return &pb.GroupMessageRangeIDs{
		StartId: startID,
		EndId:   endID,
		Ids:     ids,
	}, nil
}

func (m *MessageProto) getMessagesByIDs(groupID string, msgIDs []string) ([]*pb.GroupMessage, error) {
	return m.data.GetMessagesByIDs(groupID, msgIDs)
}
