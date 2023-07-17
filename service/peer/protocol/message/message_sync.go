package message

import (
	"bytes"
	"context"
	"crypto/sha1"
	"time"

	"github.com/jianbo-zh/dchat/service/peer/protocol/message/pb"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
	"google.golang.org/protobuf/proto"
)

func (p *PeerMessageSvc) RunSync(peerID peer.ID) {
	ctx := context.Background()
	stream, err := p.host.NewStream(ctx, peerID, SYNC_ID)
	if err != nil {
		return
	}
	defer stream.Close()

	summary, err := p.getMessageSummary(peerID)
	if err != nil {
		return
	}

	bs, err := proto.Marshal(summary)
	if err != nil {
		return
	}

	wt := pbio.NewDelimitedWriter(stream)
	if err = wt.WriteMsg(&pb.PeerSyncMessage{
		Type:    pb.PeerSyncMessage_SUMMARY,
		Payload: bs,
	}); err != nil {
		return
	}

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)

	err = p.loopSync(peerID, stream, rd, wt)
	if err != nil {
		log.Errorf("loop sync error: %v", err)
	}
}

func (p *PeerMessageSvc) SyncHandler(stream network.Stream) {

	defer stream.Close()

	peerID := stream.Conn().RemotePeer()

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	wt := pbio.NewDelimitedWriter(stream)

	// 后台同步处理
	err := p.loopSync(peerID, stream, rd, wt)
	if err != nil {
		log.Errorf("loop sync error: %v", err)
	}
}

func (p *PeerMessageSvc) loopSync(peerID peer.ID, stream network.Stream, rd pbio.ReadCloser, wt pbio.WriteCloser) error {

	var syncmsg pb.PeerSyncMessage

	for {
		syncmsg.Reset()

		// 设置读取超时，
		stream.SetReadDeadline(time.Now().Add(5 * StreamTimeout))
		if err := rd.ReadMsg(&syncmsg); err != nil {
			return err
		}
		stream.SetReadDeadline(time.Time{})

		switch syncmsg.Type {
		case pb.PeerSyncMessage_SUMMARY:
			if err := p.handleSyncSummary(peerID, &syncmsg, wt); err != nil {
				return err
			}

		case pb.PeerSyncMessage_RANGE_HASH:
			if err := p.handleSyncRangeHash(peerID, &syncmsg, wt); err != nil {
				return err
			}

		case pb.PeerSyncMessage_RANGE_IDS:
			if err := p.handleSyncRangeIDs(peerID, &syncmsg, wt); err != nil {
				return err
			}

		case pb.PeerSyncMessage_PUSH_MSG:
			if err := p.handleSyncPushMsg(peerID, &syncmsg); err != nil {
				return err
			}

		case pb.PeerSyncMessage_PULL_MSG:
			if err := p.handleSyncPullMsg(peerID, &syncmsg, wt); err != nil {
				return err
			}

		case pb.PeerSyncMessage_DONE:
			return nil

		default:
			// no defined
		}
	}
}

func (p *PeerMessageSvc) handleSyncSummary(peerID peer.ID, syncmsg *pb.PeerSyncMessage, wt pbio.WriteCloser) error {

	var summary pb.DataSummary
	if err := proto.Unmarshal(syncmsg.Payload, &summary); err != nil {
		return err
	}

	summary2, err := p.getMessageSummary(peerID)
	if err != nil {
		return err
	}

	if !summary.IsEnd {
		// 结束标记，避免无限循环
		summary2.IsEnd = true
		payload, err := proto.Marshal(summary2)
		if err != nil {
			return err
		}

		if err = wt.WriteMsg(&pb.PeerSyncMessage{
			Type:    pb.PeerSyncMessage_SUMMARY,
			Payload: payload,
		}); err != nil {
			return err
		}
	}

	if summary2.TailId > summary.TailId {
		// 如果有最新的数据则发送给对方
		msgs, err := p.getRangeMessages(peerID, summary.TailId, summary2.TailId)
		if err != nil {
			return err
		}

		for _, msg := range msgs {
			bs, err := proto.Marshal(msg)
			if err != nil {
				return err
			}

			if err = wt.WriteMsg(&pb.PeerSyncMessage{
				Type:    pb.PeerSyncMessage_PUSH_MSG,
				Payload: bs,
			}); err != nil {
				return err
			}
		}
	}

	if summary2.HeadId > summary.HeadId {
		// 当前数据要少些，则作为同步主动方
		startID := summary2.HeadId
		endID := summary.TailId

		if summary2.TailId < summary.TailId {
			endID = summary2.TailId
		}

		hash, err := p.rangeHash(peerID, startID, endID)
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

		if err = wt.WriteMsg(&pb.PeerSyncMessage{
			Type:    pb.PeerSyncMessage_RANGE_HASH,
			Payload: bs,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *PeerMessageSvc) handleSyncRangeHash(peerID peer.ID, syncmsg *pb.PeerSyncMessage, wt pbio.WriteCloser) error {
	var hashmsg pb.DataRangeHash
	if err := proto.Unmarshal(syncmsg.Payload, &hashmsg); err != nil {
		return err
	}

	// 我也计算hash
	hashBytes, err := p.rangeHash(peerID, hashmsg.StartId, hashmsg.EndId)
	if err != nil {
		log.Errorf("range hash error: %v", err)
	}

	if bytes.Equal(hashmsg.Hash, hashBytes) {
		// hash相同，不需要再同步了
		return wt.WriteMsg(&pb.PeerSyncMessage{Type: pb.PeerSyncMessage_DONE})
	}

	// hash 不同，则消息不一致，则同步消息ID
	hashids, err := p.getRangeIDs(peerID, hashmsg.StartId, hashmsg.EndId)
	if err != nil {
		return err
	}

	bs, err := proto.Marshal(hashids)
	if err != nil {
		return err
	}

	if err = wt.WriteMsg(&pb.PeerSyncMessage{
		Type:    pb.PeerSyncMessage_RANGE_IDS,
		Payload: bs,
	}); err != nil {
		return err
	}

	return nil
}

func (p *PeerMessageSvc) handleSyncRangeIDs(peerID peer.ID, syncmsg *pb.PeerSyncMessage, wt pbio.WriteCloser) error {
	var idmsg pb.DataRangeIDs
	if err := proto.Unmarshal(syncmsg.Payload, &idmsg); err != nil {
		return err
	}

	idmsg2, err := p.getRangeIDs(peerID, idmsg.StartId, idmsg.EndId)
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
		msgs, err := p.getMessagesByIDs(peerID, moreIDs)
		if err != nil {
			return err
		}

		for _, msg := range msgs {
			bs, err := proto.Marshal(msg)
			if err != nil {
				return err
			}

			if err = wt.WriteMsg(&pb.PeerSyncMessage{
				Type:    pb.PeerSyncMessage_PUSH_MSG,
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

		if err = wt.WriteMsg(&pb.PeerSyncMessage{
			Type:    pb.PeerSyncMessage_PULL_MSG,
			Payload: bs,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *PeerMessageSvc) handleSyncPushMsg(peerID peer.ID, syncmsg *pb.PeerSyncMessage) error {
	var msg pb.Message
	if err := proto.Unmarshal(syncmsg.Payload, &msg); err != nil {
		return err
	}

	if err := p.data.SaveMessage(context.Background(), peerID, &msg); err != nil {
		return err
	}
	return nil
}

func (p *PeerMessageSvc) handleSyncPullMsg(peerID peer.ID, syncmsg *pb.PeerSyncMessage, wt pbio.WriteCloser) error {
	var pullmsg pb.DataPullMsg
	if err := proto.Unmarshal(syncmsg.Payload, &pullmsg); err != nil {
		return err
	}

	msgs, err := p.data.GetMessagesByIDs(peerID, pullmsg.Ids)
	if err != nil {
		return err
	}

	for _, msg := range msgs {
		bs, err := proto.Marshal(msg)
		if err != nil {
			return err
		}

		if err = wt.WriteMsg(&pb.PeerSyncMessage{
			Type:    pb.PeerSyncMessage_PUSH_MSG,
			Payload: bs,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *PeerMessageSvc) getMessageSummary(peerID peer.ID) (*pb.DataSummary, error) {

	ctx := context.Background()

	// headID
	headID, err := p.data.GetMessageHead(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// tailID
	tailID, err := p.data.GetMessageTail(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// len
	length, err := p.data.GetMessageLength(ctx, peerID)
	if err != nil {
		return nil, err
	}

	return &pb.DataSummary{
		HeadId: headID,
		TailId: tailID,
		Length: length,
	}, nil
}

func (p *PeerMessageSvc) getRangeIDs(peerID peer.ID, startID string, endID string) (*pb.DataRangeIDs, error) {
	ids, err := p.data.GetRangeIDs(peerID, startID, endID)
	if err != nil {
		return nil, err
	}

	return &pb.DataRangeIDs{
		StartId: startID,
		EndId:   endID,
		Ids:     ids,
	}, nil
}

func (p *PeerMessageSvc) getRangeMessages(peerID peer.ID, startID string, endID string) ([]*pb.Message, error) {
	return p.data.GetRangeMessages(peerID, startID, endID)
}

func (p *PeerMessageSvc) getMessagesByIDs(peerID peer.ID, msgIDs []string) ([]*pb.Message, error) {
	return p.data.GetMessagesByIDs(peerID, msgIDs)
}

func (p *PeerMessageSvc) rangeHash(peerID peer.ID, startID string, endID string) ([]byte, error) {
	ids, err := p.data.GetRangeIDs(peerID, startID, endID)
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
