package ds

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/deposit/protocol/peer/pb"
)

type DepositMessageIface interface {
	ipfsds.Batching

	SaveDepositMessage(msg *pb.DepositMessage) error
	GetDepositMessages(peerID string, offset int, limit int, startTime int64, lastID string) (msgs []*pb.DepositMessage, err error)

	SetLastAckID(peerID string, ackID string) error
	GetLastAckID(peerID string) (string, error)
}
