package contactmsgproto

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/contactpb"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
	"google.golang.org/protobuf/proto"
)

func (p *PeerMessageProto) goSyncMessage(contactID peer.ID) {
	log.Debugln("start sync contact: ", contactID.String())

	ctx := context.Background()
	stream, err := p.host.NewStream(network.WithDialPeerTimeout(ctx, mytype.DialTimeout), contactID, MSG_SYNC_ID)
	if err != nil {
		log.Errorf("host sync new stream error: %v", err)
		return
	}
	defer stream.Close()

	summary, err := p.getMessageSummary(contactID)
	if err != nil {
		log.Errorf("get message summary error: %v", err)
		return
	}

	bs, err := proto.Marshal(summary)
	if err != nil {
		log.Errorf("proto marshal summary error: %v", err)
		return
	}

	wt := pbio.NewDelimitedWriter(stream)
	if err = wt.WriteMsg(&pb.ContactSyncMessage{
		Type:    pb.ContactSyncMessage_SUMMARY,
		Payload: bs,
	}); err != nil {
		log.Errorf("pbio write summary msg error: %v", err)
		return
	}

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	err = p.loopSync(contactID, stream, rd, wt)
	if err != nil {
		log.Errorf("loop sync error: %v", err)
	}

	log.Debugln("end sync contact: ", contactID.String())
}

func (p *PeerMessageProto) syncIDHandler(stream network.Stream) {
	log.Infoln("handle sync contact")

	peerID := stream.Conn().RemotePeer()
	defer stream.Close()

	// 后台同步处理
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	wt := pbio.NewDelimitedWriter(stream)

	err := p.loopSync(peerID, stream, rd, wt)
	if err != nil {
		log.Errorf("loop sync error: %v", err)
	}
}

func (p *PeerMessageProto) loopSync(peerID peer.ID, stream network.Stream, rd pbio.ReadCloser, wt pbio.WriteCloser) error {

	var syncmsg pb.ContactSyncMessage

	for {
		syncmsg.Reset()

		// 设置读取超时，
		stream.SetReadDeadline(time.Now().Add(5 * StreamTimeout))
		if err := rd.ReadMsg(&syncmsg); err != nil {
			return fmt.Errorf("pbio read sync msg error: %w", err)
		}
		stream.SetReadDeadline(time.Time{})

		switch syncmsg.Type {
		case pb.ContactSyncMessage_SUMMARY:
			log.Debugln("read sync msg summary")
			if err := p.handleSyncSummary(peerID, &syncmsg, wt); err != nil {
				return fmt.Errorf("handle sync summary error: %w", err)
			}

		case pb.ContactSyncMessage_RANGE_HASH:
			log.Debugln("read sync range hash")
			if err := p.handleSyncRangeHash(peerID, &syncmsg, wt); err != nil {
				return fmt.Errorf("handle sync range hash error: %w", err)
			}

		case pb.ContactSyncMessage_RANGE_IDS:
			log.Debugln("read sync range ids")
			if err := p.handleSyncRangeIDs(peerID, &syncmsg, wt); err != nil {
				return fmt.Errorf("handle sync range ids error: %w", err)
			}

		case pb.ContactSyncMessage_PUSH_MSG:
			log.Debugln("read sync push msg")
			if err := p.handleSyncPushMsg(peerID, &syncmsg); err != nil {
				return fmt.Errorf("handle sync push msg error: %w", err)
			}

		case pb.ContactSyncMessage_PULL_MSG:
			log.Debugln("read sync pull msg")
			if err := p.handleSyncPullMsg(peerID, &syncmsg, wt); err != nil {
				return fmt.Errorf("handle sync pull msg error: %w", err)
			}

		case pb.ContactSyncMessage_DONE:
			return nil

		default:
			// no defined
		}
	}
}

func (p *PeerMessageProto) handleSyncSummary(peerID peer.ID, syncmsg *pb.ContactSyncMessage, wt pbio.WriteCloser) error {

	var remoteSummary pb.ContactDataSummary
	if err := proto.Unmarshal(syncmsg.Payload, &remoteSummary); err != nil {
		return fmt.Errorf("proto unmarshal error: %w", err)
	}

	err := p.data.MergeLamportTime(context.Background(), peerID, remoteSummary.Lamptime)
	if err != nil {
		return fmt.Errorf("merge lamptime error: %w", err)
	}

	localSummary, err := p.getMessageSummary(peerID)
	if err != nil {
		return fmt.Errorf("get message summary error: %w", err)
	}

	if !remoteSummary.IsEnd {
		// 结束标记，避免无限循环
		localSummary.IsEnd = true
		payload, err := proto.Marshal(localSummary)
		if err != nil {
			return fmt.Errorf("proto marshal summary error: %w", err)
		}

		if err = wt.WriteMsg(&pb.ContactSyncMessage{
			Type:    pb.ContactSyncMessage_SUMMARY,
			Payload: payload,
		}); err != nil {
			return fmt.Errorf("pbio write sync summary error: %w", err)
		}
	}

	if localSummary.TailId > remoteSummary.TailId {
		// 如果有最新的数据则发送给对方
		msgs, err := p.getRangeMessages(peerID, remoteSummary.TailId, localSummary.TailId)
		if err != nil {
			return fmt.Errorf("get range messages error: %w", err)
		}

		for _, msg := range msgs {
			bs, err := proto.Marshal(msg)
			if err != nil {
				return fmt.Errorf("proto marshal msg error: %w", err)
			}

			if err = wt.WriteMsg(&pb.ContactSyncMessage{
				Type:    pb.ContactSyncMessage_PUSH_MSG,
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

	endID := remoteSummary.TailId // 本地多的已经发送了，这里同步剩余的
	hash, err := p.rangeHash(peerID, startID, endID)
	if err != nil {
		return fmt.Errorf("range hash error: %w", err)
	}

	bs, err := proto.Marshal(&pb.ContactDataRangeHash{
		StartId: startID,
		EndId:   endID,
		Hash:    hash,
	})
	if err != nil {
		return fmt.Errorf("proto marshal range hash error: %w", err)
	}

	if err = wt.WriteMsg(&pb.ContactSyncMessage{
		Type:    pb.ContactSyncMessage_RANGE_HASH,
		Payload: bs,
	}); err != nil {
		return fmt.Errorf("pbio write range hash error: %w", err)
	}

	return nil
}

func (p *PeerMessageProto) handleSyncRangeHash(peerID peer.ID, syncmsg *pb.ContactSyncMessage, wt pbio.WriteCloser) error {
	var hashmsg pb.ContactDataRangeHash
	if err := proto.Unmarshal(syncmsg.Payload, &hashmsg); err != nil {
		return fmt.Errorf("proto unmarshal error: %w", err)
	}

	// 我也计算hash
	hashBytes, err := p.rangeHash(peerID, hashmsg.StartId, hashmsg.EndId)
	if err != nil {
		return fmt.Errorf("range hash error: %w", err)
	}

	if !bytes.Equal(hashmsg.Hash, hashBytes) {
		// hash 不同，则消息不一致，则同步消息ID
		hashids, err := p.getRangeIDs(peerID, hashmsg.StartId, hashmsg.EndId)
		if err != nil {
			return fmt.Errorf("get range ids error: %w", err)
		}

		bs, err := proto.Marshal(hashids)
		if err != nil {
			return fmt.Errorf("proto marshal hash ids error: %w", err)
		}

		if err = wt.WriteMsg(&pb.ContactSyncMessage{
			Type:    pb.ContactSyncMessage_RANGE_IDS,
			Payload: bs,
		}); err != nil {
			return fmt.Errorf("pbio write range ids error: %w", err)
		}
	}

	return nil
}

func (p *PeerMessageProto) handleSyncRangeIDs(peerID peer.ID, syncmsg *pb.ContactSyncMessage, wt pbio.WriteCloser) error {
	var remoteRangeIDs pb.ContactDataRangeIDs
	if err := proto.Unmarshal(syncmsg.Payload, &remoteRangeIDs); err != nil {
		return fmt.Errorf("proto unmarshal sync msg payload error: %w", err)
	}

	localRangeIDs, err := p.getRangeIDs(peerID, remoteRangeIDs.StartId, remoteRangeIDs.EndId)
	if err != nil {
		return fmt.Errorf("get range ids error: %w", err)
	}

	// 比较不同点
	remoteIDsMap := make(map[string]struct{}, len(remoteRangeIDs.Ids))
	localIDsMap := make(map[string]struct{}, len(localRangeIDs.Ids))

	for _, id := range remoteRangeIDs.Ids {
		remoteIDsMap[id] = struct{}{}
	}

	var moreIDs []string
	for _, id := range localRangeIDs.Ids {
		localIDsMap[id] = struct{}{}
		if _, exists := remoteIDsMap[id]; !exists {
			moreIDs = append(moreIDs, id)
		}
	}

	if len(moreIDs) > 0 {
		msgs, err := p.getMessagesByIDs(peerID, moreIDs)
		if err != nil {
			return fmt.Errorf("get message by ids error: %w", err)
		}

		for _, msg := range msgs {
			bs, err := proto.Marshal(msg)
			if err != nil {
				return fmt.Errorf("proto marshal msg error: %w", err)
			}

			if err = wt.WriteMsg(&pb.ContactSyncMessage{
				Type:    pb.ContactSyncMessage_PUSH_MSG,
				Payload: bs,
			}); err != nil {
				return fmt.Errorf("pbio write push msg error: %w", err)
			}
		}
	}

	var lessIDs []string
	for _, id := range remoteRangeIDs.Ids {
		if _, exists := localIDsMap[id]; !exists {
			lessIDs = append(lessIDs, id)
		}
	}

	if len(lessIDs) > 0 {
		bs, err := proto.Marshal(&pb.ContactDataPullMsg{
			Ids: lessIDs,
		})
		if err != nil {
			return fmt.Errorf("proto marshal pull msg error: %w", err)
		}

		if err = wt.WriteMsg(&pb.ContactSyncMessage{
			Type:    pb.ContactSyncMessage_PULL_MSG,
			Payload: bs,
		}); err != nil {
			return fmt.Errorf("pbio write pull msg error: %w", err)
		}
	}

	return nil
}

func (p *PeerMessageProto) handleSyncPushMsg(contactID peer.ID, syncmsg *pb.ContactSyncMessage) error {
	var coreMsg pb.ContactMessage_CoreMessage
	if err := proto.Unmarshal(syncmsg.Payload, &coreMsg); err != nil {
		return fmt.Errorf("proto unmarshal payload error: %w", err)
	}

	if err := p.saveCoreMessage(context.Background(), contactID, &coreMsg); err != nil {
		return fmt.Errorf("data save msg error: %w", err)
	}
	return nil
}

func (p *PeerMessageProto) handleSyncPullMsg(peerID peer.ID, syncmsg *pb.ContactSyncMessage, wt pbio.WriteCloser) error {
	var pullmsg pb.ContactDataPullMsg
	if err := proto.Unmarshal(syncmsg.Payload, &pullmsg); err != nil {
		return fmt.Errorf("proto unmarshal payload error: %w", err)
	}

	msgs, err := p.data.GetMessagesByIDs(peerID, pullmsg.Ids)
	if err != nil {
		return fmt.Errorf("data get message by ids error: %w", err)
	}

	for _, msg := range msgs {
		bs, err := proto.Marshal(msg)
		if err != nil {
			return fmt.Errorf("proto marshal msg error: %w", err)
		}

		if err = wt.WriteMsg(&pb.ContactSyncMessage{
			Type:    pb.ContactSyncMessage_PUSH_MSG,
			Payload: bs,
		}); err != nil {
			return fmt.Errorf("pbio write push msg error: %w", err)
		}
	}

	return nil
}

func (p *PeerMessageProto) getMessageSummary(peerID peer.ID) (*pb.ContactDataSummary, error) {

	ctx := context.Background()

	// headID
	headID, err := p.data.GetMessageHead(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("data get msg head error: %w", err)
	}

	// tailID
	tailID, err := p.data.GetMessageTail(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("data get msg tail error: %w", err)
	}

	// len
	length, err := p.data.GetMessageLength(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("data get msg length error: %w", err)
	}

	// lamport time
	lamptime, err := p.data.GetLamportTime(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("data lamptime error: %w", err)
	}

	return &pb.ContactDataSummary{
		HeadId:   headID,
		TailId:   tailID,
		Length:   length,
		Lamptime: lamptime,
	}, nil
}

func (p *PeerMessageProto) getRangeIDs(peerID peer.ID, startID string, endID string) (*pb.ContactDataRangeIDs, error) {
	ids, err := p.data.GetRangeIDs(peerID, startID, endID)
	if err != nil {
		return nil, fmt.Errorf("data get range ids error: %w", err)
	}

	return &pb.ContactDataRangeIDs{
		StartId: startID,
		EndId:   endID,
		Ids:     ids,
	}, nil
}

func (p *PeerMessageProto) getRangeMessages(peerID peer.ID, startID string, endID string) ([]*pb.ContactMessage, error) {
	return p.data.GetRangeMessages(peerID, startID, endID)
}

func (p *PeerMessageProto) getMessagesByIDs(peerID peer.ID, msgIDs []string) ([]*pb.ContactMessage, error) {
	return p.data.GetMessagesByIDs(peerID, msgIDs)
}

func (p *PeerMessageProto) rangeHash(peerID peer.ID, startID string, endID string) ([]byte, error) {
	ids, err := p.data.GetRangeIDs(peerID, startID, endID)
	if err != nil {
		return nil, err
	}
	hash := sha1.New()
	for _, id := range ids {
		if _, err = hash.Write([]byte(id)); err != nil {
			return nil, fmt.Errorf("sha1 hash write error: %w", err)
		}
	}

	return hash.Sum(nil), nil
}
