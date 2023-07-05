package admin

import "github.com/libp2p/go-libp2p/core/peer"

type Group struct{}
type CreateGroupParam struct {
	Name    string
	Members []peer.ID
}
type DisbandGroupParam struct{}
type ListGroupsParam struct{}
type GroupNameParam struct{}
type SetGroupNameParam struct{}
type SetGroupRemarkParam struct{}
type GroupNoticeParam struct{}
type SetGroupNoticeParam struct{}

type Member struct {
	PeerID peer.ID
}
type InviteMemberParam struct{}
type ApplyMemberParam struct{}
type ReviewMemberParam struct{}
type RemoveMemberParam struct{}
type ListMembersParam struct{}

type Message struct{}
type SendMessageParam struct{}
type ListMessagesParam struct{}
