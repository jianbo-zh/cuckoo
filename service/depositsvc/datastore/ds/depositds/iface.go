package depositds

import (
	ipfsds "github.com/ipfs/go-datastore"
	pb "github.com/jianbo-zh/dchat/service/depositsvc/protobuf/pb/depositpb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type DepositMessageIface interface {
	ipfsds.Batching

	SaveContactMessage(msg *pb.DepositContactMessage) error
	GetContactMessages(peerID peer.ID, startID string, limit int) (msgs []*pb.DepositContactMessage, err error)

	SaveGroupMessage(msg *pb.DepositGroupMessage) error
	GetGroupMessages(groupID string, startID string, limit int) (msgs []*pb.DepositGroupMessage, err error)

	SaveSystemMessage(msg *pb.DepositSystemMessage) error
	GetSystemMessages(peerID peer.ID, startID string, limit int) (msgs []*pb.DepositSystemMessage, err error)

	SetContactLastID(peerID peer.ID, depositID string) error
	GetContactLastID(peerID peer.ID) (string, error)

	SetGroupLastID(groupID string, depositID string) error
	GetGroupLastID(groupID string) (string, error)

	SetSystemLastID(peerID peer.ID, depositID string) error
	GetSystemLastID(peerID peer.ID) (string, error)
}
