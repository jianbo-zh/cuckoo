package service

import (
	"context"
	"strings"
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
)

var _ proto.GroupSvcServer = (*GroupSvc)(nil)

type GroupSvc struct {
	proto.UnimplementedGroupSvcServer
}

func (g *GroupSvc) ClearGroupMessage(ctx context.Context, request *proto.ClearGroupMessageRequest) (*proto.ClearGroupMessageReply, error) {

	groupMessages = make([]*proto.GroupMessage, 0)

	reply := &proto.ClearGroupMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}
func (g *GroupSvc) CreateGroup(ctx context.Context, request *proto.CreateGroupRequest) (*proto.CreateGroupReply, error) {

	members := make([]*proto.Contact, 0)
	for _, PeerID := range request.PeerIDs {
		for _, contact := range contacts {
			if contact.PeerID == PeerID {
				members = append(members, &proto.Contact{
					PeerID: contact.PeerID,
					Avatar: contact.Avatar,
					Name:   contact.Name,
					Alias:  contact.Name,
				})
			}
		}
	}

	groups = append(groups, &proto.GroupFull{
		GroupID: "groupID",
		Avatar:  request.Avatar,
		Name:    request.Name,
		Alias:   request.Name,
		Notice:  "",
		Admin: &proto.Contact{
			PeerID: account.PeerID,
			Avatar: account.Avatar,
			Name:   account.Name,
			Alias:  account.Name,
		},
		JoinGroupWithoutReview: true,
		CreateTime:             time.Now().Unix(),
		UpdateTime:             time.Now().Unix(),
	})

	for _, member := range members {
		groupMembers = append(groupMembers, &proto.GroupMember{
			GroupID: "groupID",
			PeerID:  member.PeerID,
			Avatar:  member.Avatar,
			Name:    member.Name,
			Alias:   member.Name,
		})
	}

	reply := &proto.CreateGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) DeleteGroup(ctx context.Context, request *proto.DeleteGroupRequest) (*proto.DeleteGroupReply, error) {

	groups = make([]*proto.GroupFull, 0)

	reply := &proto.DeleteGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) ExitGroup(ctx context.Context, request *proto.ExitGroupRequest) (*proto.ExitGroupReply, error) {
	reply := &proto.ExitGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupBase(ctx context.Context, request *proto.GetGroupBaseRequest) (*proto.GetGroupBaseReply, error) {

	var group *proto.GroupBase
	if len(groups) > 0 {
		group = &proto.GroupBase{
			GroupID:     groups[0].GroupID,
			Avatar:      groups[0].Avatar,
			Name:        groups[0].Name,
			Alias:       groups[0].Alias,
			LastMessage: "",
			UpdateTime:  time.Now().Unix(),
		}
	}

	reply := &proto.GetGroupBaseReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Group: group,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupFull(ctx context.Context, request *proto.GetGroupFullRequest) (*proto.GetGroupFullReply, error) {

	var group *proto.GroupFull
	if len(groups) > 0 {
		group = groups[0]
	}

	reply := &proto.GetGroupFullReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Group: group,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupList(ctx context.Context, request *proto.GetGroupListRequest) (*proto.GetGroupListReply, error) {

	groupList := make([]*proto.GroupBase, 0)
	for _, group := range groups {
		groupList = append(groupList, &proto.GroupBase{
			GroupID: group.GroupID,
			Avatar:  group.Avatar,
			Name:    group.Name,
			Alias:   group.Alias,
		})
	}

	reply := &proto.GetGroupListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		GroupList: groupList,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupMessageList(ctx context.Context, request *proto.GetGroupMessageListRequest) (*proto.GetGroupMessageListReply, error) {
	reply := &proto.GetGroupMessageListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		MessageList: groupMessages,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupMemberList(ctx context.Context, request *proto.GetGroupMemberListRequest) (*proto.GetGroupMemberListReply, error) {

	filterMemberList := make([]*proto.GroupMember, 0)
	for _, member := range groupMembers {
		if strings.Contains(member.Name, request.Keywords) {
			filterMemberList = append(filterMemberList, member)
		}
	}

	if len(filterMemberList) > 0 {

		offset := int(request.Offset)
		limit := int(request.Limit)

		if offset < 0 || offset >= len(filterMemberList) || limit <= 0 {
			filterMemberList = make([]*proto.GroupMember, 0)

		} else {
			endOffset := offset + limit
			if endOffset > len(filterMemberList) {
				endOffset = len(filterMemberList)
			}

			filterMemberList = filterMemberList[offset:endOffset]
		}
	}

	reply := &proto.GetGroupMemberListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		MemberList: filterMemberList,
	}
	return reply, nil
}

func (g *GroupSvc) InviteJoinGroup(ctx context.Context, request *proto.InviteJoinGroupRequest) (*proto.InviteJoinGroupReply, error) {

	groupMembers = append(groupMembers, &proto.GroupMember{
		GroupID: request.GroupID,
		PeerID:  request.PeerID,
		Avatar:  "avatar1",
		Name:    "name1",
		Alias:   "alias1",
	})

	reply := &proto.InviteJoinGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) RemoveGroupMember(ctx context.Context, request *proto.RemoveGroupMemberRequest) (*proto.RemoveGroupMemberReply, error) {

	if len(groupMembers) > 0 {
		groupMembers2 := make([]*proto.GroupMember, 0)
		for _, member := range groupMembers {
			if member.PeerID == request.PeerID && member.GroupID == request.GroupID {
				groupMembers2 = append(groupMembers2, member)
			}
		}
		groupMembers = groupMembers2
	}

	reply := &proto.RemoveGroupMemberReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) SendGroupMessage(ctx context.Context, request *proto.SendGroupMessageRequest) (*proto.SendGroupMessageReply, error) {

	groupMessages = append(groupMessages, &proto.GroupMessage{
		ID:      "id",
		GroupID: request.GroupID,
		Sender: &proto.Contact{
			PeerID: "PeerID",
			Avatar: "avatar",
			Name:   "name",
			Alias:  "alias",
		},
		MsgType:    request.MsgType,
		MimeType:   request.MimeType,
		Data:       request.Data,
		CreateTime: time.Now().Unix(),
	})

	reply := &proto.SendGroupMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) SetGroupAlias(ctx context.Context, request *proto.SetGroupAliasRequest) (*proto.SetGroupAliasReply, error) {

	if len(groups) > 0 {
		groups[0].Alias = request.Alias
	}

	reply := &proto.SetGroupAliasReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) SetGroupAvatar(ctx context.Context, request *proto.SetGroupAvatarRequest) (*proto.SetGroupAvatarReply, error) {

	if len(groups) > 0 {
		groups[0].Avatar = request.Avatar
	}

	reply := &proto.SetGroupAvatarReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) SetGroupNotice(ctx context.Context, request *proto.SetGroupNoticeRequest) (*proto.SetGroupNoticeReply, error) {

	if len(groups) > 0 {
		groups[0].Notice = request.Notice
	}

	reply := &proto.SetGroupNoticeReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Notice: request.Notice,
	}
	return reply, nil
}

func (g *GroupSvc) SetJoinGroupReview(ctx context.Context, request *proto.SetJoinGroupReviewRequest) (*proto.SetJoinGroupReviewReply, error) {

	if len(groups) > 0 {
		groups[0].JoinGroupWithoutReview = request.IsReview
	}

	reply := &proto.SetJoinGroupReviewReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		IsReview: request.IsReview,
	}
	return reply, nil
}
