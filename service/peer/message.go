package peer

import (
	"context"

	"github.com/jianbo-zh/dchat/service/peer/protocol/message"
	"github.com/jianbo-zh/dchat/service/peer/protocol/message/pb"
	peerpeer "github.com/jianbo-zh/dchat/service/peer/protocol/peer"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerSvc struct {
	msgSvc  *message.PeerMessageSvc
	peerSvc *peerpeer.PeerPeerSvc
}

func Get() PeerServiceIface {
	if peersvc == nil {
		panic("peer service must init before use")
	}

	return peersvc
}

func (p *PeerSvc) GetMessages(ctx context.Context, peerID peer.ID, offset int, limit int) ([]PeerMessage, error) {
	var peerMsgs []PeerMessage

	msgs, err := p.msgSvc.GetMessages(ctx, peerID, offset, limit)
	if err != nil {
		return peerMsgs, err
	}

	for _, msg := range msgs {
		peerMsgs = append(peerMsgs, PeerMessage{
			ID:         msg.Id,
			Type:       convMsgType(msg.Type),
			SenderID:   peer.ID(msg.SenderId),
			ReceiverID: peer.ID(msg.ReceiverId),
			Payload:    msg.Payload,
			Timestamp:  msg.Timestamp,
			Lamportime: msg.Lamportime,
		})
	}

	return peerMsgs, nil
}

func (p *PeerSvc) SendTextMessage(ctx context.Context, peerID peer.ID, msg string) error {
	return p.msgSvc.SendTextMessage(ctx, peerID, msg)
}

func (p *PeerSvc) SendGroupInviteMessage(ctx context.Context, peerID peer.ID, groupID string) error {
	return p.msgSvc.SendGroupInviteMessage(ctx, peerID, groupID)
}

func convMsgType(mtype pb.Message_Type) MsgType {
	switch mtype {
	case pb.Message_TEXT:
		return MsgTypeText
	case pb.Message_AUDIO:
		return MsgTypeAudio
	case pb.Message_VIDEO:
		return MsgTypeVideo
	case pb.Message_INVITE:
		return MsgTypeInvite
	default:
		return MsgTypeText
	}
}
