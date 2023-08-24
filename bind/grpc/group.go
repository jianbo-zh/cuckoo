package service

import (
	"context"
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
		members = append(members, &proto.Contact{
			PeerID: PeerID,
			Avatar: "avatar1",
			Name:   "name1",
			Alias:  "alias1",
		})
	}

	groups = append(groups, &proto.Group{
		GroupID: "groupID",
		Name:    request.Name,
		Alias:   request.Name,
		Notice:  "",
		Admin: &proto.Contact{
			PeerID: account.PeerID,
			Avatar: account.Avatar,
			Name:   account.Name,
			Alias:  account.Name,
		},
		Members:                members,
		JoinGroupWithoutReview: true,
		CreateTime:             time.Now().Unix(),
		UpdateTime:             time.Now().Unix(),
	})

	reply := &proto.CreateGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) DeleteGroup(ctx context.Context, request *proto.DeleteGroupRequest) (*proto.DeleteGroupReply, error) {

	groups = make([]*proto.Group, 0)

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

func (g *GroupSvc) GetGroup(ctx context.Context, request *proto.GetGroupRequest) (*proto.GetGroupReply, error) {

	var group *proto.Group
	if len(groups) > 0 {
		group = groups[0]
	}

	reply := &proto.GetGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Group: group,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupList(ctx context.Context, request *proto.GetGroupListRequest) (*proto.GetGroupListReply, error) {
	reply := &proto.GetGroupListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		GroupList: groups,
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

func (g *GroupSvc) GetMemberList(ctx context.Context, request *proto.GetMemberListRequest) (*proto.GetMemberListReply, error) {
	reply := &proto.GetMemberListReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		MemberList: groupMembers,
	}
	return reply, nil
}

func (g *GroupSvc) InviteJoinGroup(ctx context.Context, request *proto.InviteJoinGroupRequest) (*proto.InviteJoinGroupReply, error) {

	groupMembers = append(groupMembers, &proto.Contact{
		PeerID: request.PeerID,
		Avatar: "avatar1",
		Name:   "name1",
		Alias:  "alias1",
	})

	reply := &proto.InviteJoinGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) RemoveMember(ctx context.Context, request *proto.RemoveMemberRequest) (*proto.RemoveMemberReply, error) {

	if len(groupMembers) > 0 {
		groupMembers = make([]*proto.Contact, 0)
	}

	reply := &proto.RemoveMemberReply{
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
	}
	return reply, nil
}
