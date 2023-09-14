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

func (g *GroupSvc) ClearGroupMessage(ctx context.Context, request *proto.ClearGroupMessageRequest) (reply *proto.ClearGroupMessageReply, err error) {

	log.Infoln("ClearGroupMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("ClearGroupMessage panic: ", e)
		} else if err != nil {
			log.Errorln("ClearGroupMessage error: ", err.Error())
		} else {
			log.Infoln("ClearGroupMessage reply: ", reply.String())
		}
	}()

	groupMessages = make([]*proto.GroupMessage, 0)

	reply = &proto.ClearGroupMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}
func (g *GroupSvc) CreateGroup(ctx context.Context, request *proto.CreateGroupRequest) (reply *proto.CreateGroupReply, err error) {

	log.Infoln("CreateGroup request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("CreateGroup panic: ", e)
		} else if err != nil {
			log.Errorln("CreateGroup error: ", err.Error())
		} else {
			log.Infoln("CreateGroup reply: ", reply.String())
		}
	}()

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

	reply = &proto.CreateGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) DeleteGroup(ctx context.Context, request *proto.DeleteGroupRequest) (reply *proto.DeleteGroupReply, err error) {

	log.Infoln("DeleteGroup request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("DeleteGroup panic: ", e)
		} else if err != nil {
			log.Errorln("DeleteGroup error: ", err.Error())
		} else {
			log.Infoln("DeleteGroup reply: ", reply.String())
		}
	}()

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	err = groupSvc.DeleteGroup(ctx, request.GroupId)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.DeleteGroup error: %w", err)
	}

	reply = &proto.DeleteGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) ExitGroup(ctx context.Context, request *proto.ExitGroupRequest) (reply *proto.ExitGroupReply, err error) {

	log.Infoln("ExitGroup request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("ExitGroup panic: ", e)
		} else if err != nil {
			log.Errorln("ExitGroup error: ", err.Error())
		} else {
			log.Infoln("ExitGroup reply: ", reply.String())
		}
	}()

	reply = &proto.ExitGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) GetGroup(ctx context.Context, request *proto.GetGroupRequest) (reply *proto.GetGroupReply, err error) {

	log.Infoln("GetGroup request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetGroup panic: ", e)
		} else if err != nil {
			log.Errorln("GetGroup error: ", err.Error())
		} else {
			log.Infoln("GetGroup reply: ", reply.String())
		}
	}()

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

	reply = &proto.GetGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Group: group,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupDetail(ctx context.Context, request *proto.GetGroupDetailRequest) (reply *proto.GetGroupDetailReply, err error) {

	log.Infoln("GetGroupDetail request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetGroupDetail panic: ", e)
		} else if err != nil {
			log.Errorln("GetGroupDetail error: ", err.Error())
		} else {
			log.Infoln("GetGroupDetail reply: ", reply.String())
		}
	}()

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	grp, err := groupSvc.GetGroup(ctx, request.GroupId)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.GetGroup error: %w", err)
	}

	group := &proto.GroupDetail{
		GroupId:    grp.ID,
		Avatar:     grp.Avatar,
		Name:       grp.Name,
		UpdateTime: time.Now().Unix(),
	}

	reply = &proto.GetGroupDetailReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Group: group,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroups(ctx context.Context, request *proto.GetGroupsRequest) (reply *proto.GetGroupsReply, err error) {

	log.Infoln("GetGroups request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetGroups panic: ", e)
		} else if err != nil {
			log.Errorln("GetGroups error: ", err.Error())
		} else {
			log.Infoln("GetGroups reply: ", reply.String())
		}
	}()

	groupList := make([]*proto.Group, 0)
	for _, group := range groups {
		groupList = append(groupList, &proto.Group{
			Id:     group.GroupId,
			Avatar: group.Avatar,
			Name:   group.Name,
		})
	}

	reply = &proto.GetGroupsReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Groups: groupList,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupMessages(ctx context.Context, request *proto.GetGroupMessagesRequest) (reply *proto.GetGroupMessagesReply, err error) {

	log.Infoln("GetGroupMessages request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetGroupMessages panic: ", e)
		} else if err != nil {
			log.Errorln("GetGroupMessages error: ", err.Error())
		} else {
			log.Infoln("GetGroupMessages reply: ", reply.String())
		}
	}()

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
			MsgType:    msgTypeToProto("text"),
			MimeType:   msg.MimeType,
			Payload:    msg.Payload,
			CreateTime: msg.Timestamp,
		})
	}

	reply = &proto.GetGroupMessagesReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Messages: msglist,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupMembers(ctx context.Context, request *proto.GetGroupMembersRequest) (reply *proto.GetGroupMembersReply, err error) {

	log.Infoln("GetGroupMembers request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetGroupMembers panic: ", e)
		} else if err != nil {
			log.Errorln("GetGroupMembers error: ", err.Error())
		} else {
			log.Infoln("GetGroupMembers reply: ", reply.String())
		}
	}()

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

	reply = &proto.GetGroupMembersReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Members: filterMemberList,
	}
	return reply, nil
}

func (g *GroupSvc) InviteJoinGroup(ctx context.Context, request *proto.InviteJoinGroupRequest) (reply *proto.InviteJoinGroupReply, err error) {

	log.Infoln("InviteJoinGroup request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("InviteJoinGroup panic: ", e)
		} else if err != nil {
			log.Errorln("InviteJoinGroup error: ", err.Error())
		} else {
			log.Infoln("InviteJoinGroup reply: ", reply.String())
		}
	}()

	groupMembers = append(groupMembers, &proto.GroupMember{
		GroupId: request.GroupId,
		PeerId:  request.ContactId,
		Avatar:  "avatar1",
		Name:    "name1",
	})

	reply = &proto.InviteJoinGroupReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) RemoveGroupMember(ctx context.Context, request *proto.RemoveGroupMemberRequest) (reply *proto.RemoveGroupMemberReply, err error) {

	log.Infoln("RemoveGroupMember request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("RemoveGroupMember panic: ", e)
		} else if err != nil {
			log.Errorln("RemoveGroupMember error: ", err.Error())
		} else {
			log.Infoln("RemoveGroupMember reply: ", reply.String())
		}
	}()

	if len(groupMembers) > 0 {
		groupMembers2 := make([]*proto.GroupMember, 0)
		for _, member := range groupMembers {
			if member.PeerId == request.MemberId && member.GroupId == request.GroupId {
				groupMembers2 = append(groupMembers2, member)
			}
		}
		groupMembers = groupMembers2
	}

	reply = &proto.RemoveGroupMemberReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) SendGroupMessage(ctx context.Context, request *proto.SendGroupMessageRequest) (reply *proto.SendGroupMessageReply, err error) {

	log.Infoln("SendGroupMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendGroupMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SendGroupMessage error: ", err.Error())
		} else {
			log.Infoln("SendGroupMessage reply: ", reply.String())
		}
	}()

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	err = groupSvc.SendMessage(ctx, request.GroupId, protoToMsgType(request.MsgType), request.MimeType, request.Payload)
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

	reply = &proto.SendGroupMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Message: &sendMsg,
	}
	return reply, nil
}

func (g *GroupSvc) SetGroupName(ctx context.Context, request *proto.SetGroupNameRequest) (reply *proto.SetGroupNameReply, err error) {

	log.Infoln("SetGroupName request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetGroupName panic: ", e)
		} else if err != nil {
			log.Errorln("SetGroupName error: ", err.Error())
		} else {
			log.Infoln("SetGroupName reply: ", reply.String())
		}
	}()

	if len(groups) > 0 {
		groups[0].Name = request.Name
	}

	reply = &proto.SetGroupNameReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Name: request.Name,
	}
	return reply, nil
}

func (g *GroupSvc) SetGroupAvatar(ctx context.Context, request *proto.SetGroupAvatarRequest) (reply *proto.SetGroupAvatarReply, err error) {

	log.Infoln("SetGroupAvatar request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetGroupAvatar panic: ", e)
		} else if err != nil {
			log.Errorln("SetGroupAvatar error: ", err.Error())
		} else {
			log.Infoln("SetGroupAvatar reply: ", reply.String())
		}
	}()

	if len(groups) > 0 {
		groups[0].Avatar = request.Avatar
	}

	reply = &proto.SetGroupAvatarReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}

func (g *GroupSvc) SetGroupNotice(ctx context.Context, request *proto.SetGroupNoticeRequest) (reply *proto.SetGroupNoticeReply, err error) {

	log.Infoln("SetGroupNotice request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetGroupNotice panic: ", e)
		} else if err != nil {
			log.Errorln("SetGroupNotice error: ", err.Error())
		} else {
			log.Infoln("SetGroupNotice reply: ", reply.String())
		}
	}()

	if len(groups) > 0 {
		groups[0].Notice = request.Notice
	}

	reply = &proto.SetGroupNoticeReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Notice: request.Notice,
	}
	return reply, nil
}

func (g *GroupSvc) SetJoinGroupReview(ctx context.Context, request *proto.SetJoinGroupReviewRequest) (reply *proto.SetJoinGroupReviewReply, err error) {

	log.Infoln("SetJoinGroupReview request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetJoinGroupReview panic: ", e)
		} else if err != nil {
			log.Errorln("SetJoinGroupReview error: ", err.Error())
		} else {
			log.Infoln("SetJoinGroupReview reply: ", reply.String())
		}
	}()

	if len(groups) > 0 {
		groups[0].AutoJoinGroup = request.IsAuto
	}

	reply = &proto.SetJoinGroupReviewReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		IsAuto: request.IsAuto,
	}
	return reply, nil
}
