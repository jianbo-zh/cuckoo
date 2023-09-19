package types

import "github.com/libp2p/go-libp2p/core/peer"

const (
	SystemTypeApplyAddContact string = "apply_add_contact"
	SystemTypeInviteJoinGroup string = "invite_join_group"
)

const (
	SystemStateSended   string = "is_send"
	SystemStateAgreed   string = "is_agree"
	SystemStateRejected string = "is_reject"
)

type SystemMessage struct {
	ID          string
	SystemType  string
	GroupID     string
	FromPeer    Peer
	ToPeerID    peer.ID
	Content     string
	SystemState string
	CreateTime  int64
	UpdateTime  int64
}
