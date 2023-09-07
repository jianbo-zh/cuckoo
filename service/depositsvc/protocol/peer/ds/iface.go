package ds

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/depositsvc/protocol/peer/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type DepositMessageIface interface {
	ipfsds.Batching

	SaveDepositMessage(msg *pb.OfflineMessage) error
	GetDepositMessages(peerID peer.ID, offset int, limit int, startTime int64, lastID string) (msgs []*pb.OfflineMessage, err error)

	SetLastAckID(peerID peer.ID, ackID string) error
	GetLastAckID(peerID peer.ID) (string, error)
}
