package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/service/groupsvc"
	"github.com/libp2p/go-libp2p/core/peer"
)

var _ proto.GroupSvcServer = (*GroupSvc)(nil)

type GroupSvc struct {
	getter cuckoo.CuckooGetter
	proto.UnimplementedGroupSvcServer
}

func NewGroupSvc(getter cuckoo.CuckooGetter) *GroupSvc {
	return &GroupSvc{
		getter: getter,
	}
}

func (g *GroupSvc) getGroupSvc() (groupsvc.GroupServiceIface, error) {
	cuckoo, err := g.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	groupSvc, err := cuckoo.GetGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return groupSvc, nil
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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	var memberIDs []peer.ID
	for _, pid := range request.MemberIds {
		peerID, err := peer.Decode(pid)
		if err != nil {
			return nil, fmt.Errorf("peer.Decode error: %w", err)
		}
		memberIDs = append(memberIDs, peerID)
	}

	_, err = groupSvc.CreateGroup(ctx, request.Name, request.Avatar, memberIDs)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.CreateGroup error: %w", err)
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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	err = groupSvc.DeleteGroup(ctx, request.GroupId)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.DeleteGroup error: %w", err)
	}

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
	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	grp, err := groupSvc.GetGroup(ctx, request.GroupId)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.GetGroup error: %w", err)
	}

	group := &proto.Group{
		Id:     grp.ID,
		Avatar: grp.Avatar,
		Name:   grp.Name,
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

func (g *GroupSvc) GetGroupFull(ctx context.Context, request *proto.GetGroupFullRequest) (*proto.GetGroupFullReply, error) {

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	grp, err := groupSvc.GetGroup(ctx, request.GroupId)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.GetGroup error: %w", err)
	}

	group := &proto.GroupFull{
		GroupId:    grp.ID,
		Avatar:     grp.Avatar,
		Name:       grp.Name,
		UpdateTime: time.Now().Unix(),
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

func (g *GroupSvc) GetGroups(ctx context.Context, request *proto.GetGroupsRequest) (*proto.GetGroupsReply, error) {

	groupList := make([]*proto.Group, 0)
	for _, group := range groups {
		groupList = append(groupList, &proto.Group{
			Id:     group.GroupId,
			Avatar: group.Avatar,
			Name:   group.Name,
		})
	}

	reply := &proto.GetGroupsReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Groups: groupList,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupMessages(ctx context.Context, request *proto.GetGroupMessagesRequest) (*proto.GetGroupMessagesReply, error) {

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	msgs, err := groupSvc.ListMessages(ctx, request.GroupId, int(request.Offset), int(request.Limit))
	if err != nil {
		return nil, fmt.Errorf("groupSvc.ListMessage error: %w", err)
	}

	fmt.Printf("list message size: %d\n", len(msgs))

	var msglist []*proto.GroupMessage
	for _, msg := range msgs {
		msglist = append(msglist, &proto.GroupMessage{
			Id:      msg.ID,
			GroupId: msg.GroupID,
			Sender: &proto.Peer{
				Id:     msg.FromPeer.PeerID.String(),
				Name:   msg.FromPeer.Name,
				Avatar: msg.FromPeer.Avatar,
			},
			MsgType:    "text",
			MimeType:   msg.MimeType,
			Payload:    msg.Payload,
			CreateTime: msg.Timestamp,
		})
	}

	reply := &proto.GetGroupMessagesReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Messages: msglist,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupMembers(ctx context.Context, request *proto.GetGroupMembersRequest) (*proto.GetGroupMembersReply, error) {

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

	reply := &proto.GetGroupMembersReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Members: filterMemberList,
	}
	return reply, nil
}

func (g *GroupSvc) InviteJoinGroup(ctx context.Context, request *proto.InviteJoinGroupRequest) (*proto.InviteJoinGroupReply, error) {

	groupMembers = append(groupMembers, &proto.GroupMember{
		GroupId: request.GroupId,
		PeerId:  request.ContactId,
		Avatar:  "avatar1",
		Name:    "name1",
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
			if member.PeerId == request.MemberId && member.GroupId == request.GroupId {
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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	err = groupSvc.SendMessage(ctx, request.GroupId, request.MsgType, request.MimeType, request.Payload)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.SendMessge error: %w", err)
	}

	sendMsg := proto.GroupMessage{
		Id:      "id",
		GroupId: request.GroupId,
		Sender: &proto.Peer{
			Id:     account.PeerId,
			Avatar: account.Avatar,
			Name:   account.Name,
		},
		MsgType:    request.MsgType,
		MimeType:   request.MimeType,
		Payload:    request.Payload,
		CreateTime: time.Now().Unix(),
	}

	// groupMessages = append(groupMessages, &sendMsg)

	reply := &proto.SendGroupMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Message: &sendMsg,
	}
	return reply, nil
}

func (g *GroupSvc) SetGroupName(ctx context.Context, request *proto.SetGroupNameRequest) (*proto.SetGroupNameReply, error) {

	if len(groups) > 0 {
		groups[0].Name = request.Name
	}

	reply := &proto.SetGroupNameReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Name: request.Name,
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
		groups[0].AutoJoinGroup = request.IsAuto
	}

	reply := &proto.SetJoinGroupReviewReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		IsAuto: request.IsAuto,
	}
	return reply, nil
}
