package peer

import (
	"context"
	"time"

	peerpeer "github.com/jianbo-zh/dchat/service/peer/protocol/peer"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (p *PeerSvc) AddPeer(ctx context.Context, peerID peer.ID, nickname string) error {
	return p.peerSvc.AddPeer(peerID, peerpeer.PeerInfo{
		PeerID:   peerID,
		Nickname: nickname,
		AddTs:    time.Now().Unix(),
		AccessTs: time.Now().Unix(),
	})
}

func (p *PeerSvc) GetPeers(context.Context) ([]PeerInfo, error) {

	var peers []PeerInfo

	pinfos, err := p.peerSvc.GetPeers()
	if err != nil {
		return nil, err
	}

	for _, pi := range pinfos {
		peers = append(peers, PeerInfo{
			PeerID:   pi.PeerID,
			Nickname: pi.Nickname,
			AddTs:    pi.AddTs,
			AccessTs: pi.AccessTs,
		})
	}

	return peers, nil
}
