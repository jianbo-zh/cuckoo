package ds

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/service/depositsvc/protocol/deposit/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type DepositMessageIface interface {
	ipfsds.Batching

	SaveContactMessage(msg *pb.ContactMessage) error
	GetContactMessages(peerID peer.ID, startID string, limit int) (msgs []*pb.ContactMessage, err error)

	SaveGroupMessage(msg *pb.GroupMessage) error
	GetGroupMessages(groupID string, startID string, limit int) (msgs []*pb.GroupMessage, err error)

	SetContactLastID(peerID peer.ID, depositID string) error
	GetContactLastID(peerID peer.ID) (string, error)

	SetGroupLastID(groupID string, depositID string) error
	GetGroupLastID(groupID string) (string, error)
}
