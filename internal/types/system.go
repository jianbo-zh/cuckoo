package types

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
	Sender      Peer
	Receiver    Peer
	Content     string
	SystemState string
	CreateTime  int64
	UpdateTime  int64
}
