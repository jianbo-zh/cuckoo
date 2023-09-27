package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/internal/myerror"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/depositsvc"
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

func (g *GroupSvc) getAccountSvc() (accountsvc.AccountServiceIface, error) {
	cuckoo, err := g.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	accountSvc, err := cuckoo.GetAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return accountSvc, nil
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

func (g *GroupSvc) getDepositSvc() (depositsvc.DepositServiceIface, error) {
	cuckoo, err := g.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	depositSvc, err := cuckoo.GetDepositSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return depositSvc, nil
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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	err = groupSvc.ExitGroup(ctx, request.GroupId)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.DeleteGroup error: %w", err)
	}

	reply = &proto.ExitGroupReply{
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

func (g *GroupSvc) DisbandGroup(ctx context.Context, request *proto.DisbandGroupRequest) (reply *proto.DisbandGroupReply, err error) {

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

	err = groupSvc.DisbandGroup(ctx, request.GroupId)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.DeleteGroup error: %w", err)
	}

	reply = &proto.DisbandGroupReply{
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
		Id:             grp.ID,
		Avatar:         grp.Avatar,
		Name:           grp.Name,
		DepositAddress: grp.DepositAddress.String(),
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

	grp, err := groupSvc.GetGroupDetail(ctx, request.GroupId)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.GetGroup error: %w", err)
	}

	group := &proto.GroupDetail{
		GroupId:        grp.ID,
		Avatar:         grp.Avatar,
		Name:           grp.Name,
		Notice:         grp.Notice,
		AutoJoinGroup:  grp.AutoJoinGroup,
		DepositAddress: grp.DepositAddress.String(),
		CreateTime:     grp.CreateTime,
		UpdateTime:     grp.UpdateTime,
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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	groups, err := groupSvc.GetGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get groups error: %w", err)
	}

	groupList := make([]*proto.Group, len(groups))
	for i, group := range groups {
		groupList[i] = &proto.Group{
			Id:     group.ID,
			Name:   group.Name,
			Avatar: group.Avatar,
		}
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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("get group svc error: %w", err)
	}

	err = groupSvc.SetGroupName(ctx, request.GroupId, request.Name)
	if err != nil {
		return nil, fmt.Errorf("svc set group name error: %w", err)
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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("get group svc error: %w", err)
	}

	err = groupSvc.SetGroupAvatar(ctx, request.GroupId, request.Avatar)
	if err != nil {
		return nil, fmt.Errorf("svc set group name error: %w", err)
	}

	reply = &proto.SetGroupAvatarReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Avatar: request.Avatar,
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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("get group svc error: %w", err)
	}

	err = groupSvc.SetGroupNotice(ctx, request.GroupId, request.Notice)
	if err != nil {
		return nil, fmt.Errorf("svc set group name error: %w", err)
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

func (g *GroupSvc) SetGroupAutoJoin(ctx context.Context, request *proto.SetGroupAutoJoinRequest) (reply *proto.SetGroupAutoJoinReply, err error) {

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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("get group svc error: %w", err)
	}

	err = groupSvc.SetGroupAutoJoin(ctx, request.GroupId, request.IsAuto)
	if err != nil {
		return nil, fmt.Errorf("svc set group name error: %w", err)
	}

	reply = &proto.SetGroupAutoJoinReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		IsAuto: request.IsAuto,
	}
	return reply, nil
}

func (g *GroupSvc) SetGroupDepositAddress(ctx context.Context, request *proto.SetGroupDepositAddressRequest) (reply *proto.SetGroupDepositAddressReply, err error) {

	log.Infoln("SetGroupDepositAddress request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SetGroupDepositAddress panic: ", e)
		} else if err != nil {
			log.Errorln("SetGroupDepositAddress error: ", err.Error())
		} else {
			log.Infoln("SetGroupDepositAddress reply: ", reply.String())
		}
	}()

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("get group svc error: %w", err)
	}

	var depositPeerID peer.ID
	if strings.TrimSpace(request.DepositAddress) != "" {
		depositPeerID, err = peer.Decode(strings.TrimSpace(request.DepositAddress))
		if err != nil || depositPeerID.Validate() != nil {
			return nil, fmt.Errorf("peer decode error")
		}
	}

	err = groupSvc.SetGroupDepositAddress(ctx, request.GroupId, depositPeerID)
	if err != nil {
		return nil, fmt.Errorf("svc set group deposit peer id error: %w", err)
	}

	reply = &proto.SetGroupDepositAddressReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		DepositAddress: depositPeerID.String(),
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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	members, err := groupSvc.GetGroupMembers(ctx, request.GroupId, request.Keywords, int(request.Offset), int(request.Limit))

	memberList := make([]*proto.GroupMember, len(members))

	if len(members) > 0 {
		accountSvc, err := g.getAccountSvc()
		if err != nil {
			return nil, fmt.Errorf("get account svc error: %w", err)
		}

		var peerIDs []peer.ID
		for _, member := range members {
			peerIDs = append(peerIDs, member.ID)
		}

		onlineStateMap, err := accountSvc.GetOnlineState(ctx, peerIDs)
		if err != nil {
			return nil, fmt.Errorf("get online state error: %w", err)
		}

		for i, member := range members {
			memberList[i] = &proto.GroupMember{
				Id:       member.ID.String(),
				Name:     member.Name,
				Avatar:   member.Avatar,
				IsOnline: onlineStateMap[member.ID],
			}
		}
	}

	reply = &proto.GetGroupMembersReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Members: memberList,
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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	memberID, err := peer.Decode(request.MemberId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
	}

	err = groupSvc.RemoveGroupMember(ctx, request.GroupId, memberID)
	if err != nil {
		return nil, fmt.Errorf("svc remove group member error: %w", err)
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
		return nil, fmt.Errorf("get group svc error: %w", err)
	}

	accountSvc, err := g.getAccountSvc()
	if err != nil {
		return nil, fmt.Errorf("get account service error: %w", err)
	}

	depositSvc, err := g.getDepositSvc()
	if err != nil {
		return nil, fmt.Errorf("get deposit service error: %w", err)
	}

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("svc get account error: %w", err)
	}

	var sendErr error
	groupID := request.GroupId
	msgID, err := groupSvc.SendGroupMessage(ctx, groupID, decodeMsgType(request.MsgType), request.MimeType, request.Payload)
	if err != nil {
		if msgID != "" && errors.Is(err, myerror.ErrSendGroupMessageFailed) && account.AutoDepositMessage {
			group, err2 := groupSvc.GetGroup(ctx, groupID)
			if err2 != nil {
				sendErr = fmt.Errorf("svc get group error: %w", err2)

			} else if group.DepositAddress.Validate() != nil { // 没有设置寄存节点
				sendErr = err

			} else {
				msgData, err := groupSvc.GetGroupMessageData(ctx, groupID, msgID)
				if err != nil {
					sendErr = fmt.Errorf("svc get group msg data error: %w", err)

				} else {
					err = depositSvc.PushGroupMessage(group.DepositAddress, groupID, msgID, msgData)
					if err != nil {
						sendErr = fmt.Errorf("deposit group msg error: %w", err)
					}
				}
			}

		} else {
			sendErr = err
		}
	}

	if sendErr != nil {
		if msgID != "" {
			// delete msg
			if err = groupSvc.DeleteGroupMessage(ctx, groupID, msgID); err != nil {
				return nil, fmt.Errorf("svc delete group msg error: %w", err)
			}
		}

		return nil, sendErr
	}

	message := &proto.GroupMessage{
		Id:      msgID,
		GroupId: groupID,
		Sender: &proto.Peer{
			Id:     account.ID.String(),
			Name:   account.Name,
			Avatar: account.Avatar,
		},
		MsgType:    request.MsgType,
		MimeType:   request.MimeType,
		Payload:    request.Payload,
		CreateTime: time.Now().Unix(),
	}

	reply = &proto.SendGroupMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Message: message,
	}
	return reply, nil
}

func (g *GroupSvc) GetGroupMessage(ctx context.Context, request *proto.GetGroupMessageRequest) (reply *proto.GetGroupMessageReply, err error) {

	log.Infoln("GetGroupMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetGroupMessage panic: ", e)
		} else if err != nil {
			log.Errorln("GetGroupMessage error: ", err.Error())
		} else {
			log.Infoln("GetGroupMessage reply: ", reply.String())
		}
	}()

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	msg, err := groupSvc.GetGroupMessage(ctx, request.GroupId, request.MsgId)
	if err != nil {
		return nil, fmt.Errorf("groupSvc.ListMessage error: %w", err)
	}

	reply = &proto.GetGroupMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Message: &proto.GroupMessage{
			Id:      msg.ID,
			GroupId: msg.GroupID,
			Sender: &proto.Peer{
				Id:     msg.FromPeer.ID.String(),
				Name:   msg.FromPeer.Name,
				Avatar: msg.FromPeer.Avatar,
			},
			MsgType:    encodeMsgType(msg.MsgType),
			MimeType:   msg.MimeType,
			Payload:    msg.Payload,
			CreateTime: msg.CreateTime,
		},
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

	msgs, err := groupSvc.GetGroupMessages(ctx, request.GroupId, int(request.Offset), int(request.Limit))
	if err != nil {
		return nil, fmt.Errorf("groupSvc.ListMessage error: %w", err)
	}

	msglist := make([]*proto.GroupMessage, len(msgs))
	for i, msg := range msgs {
		msglist[i] = &proto.GroupMessage{
			Id:      msg.ID,
			GroupId: msg.GroupID,
			Sender: &proto.Peer{
				Id:     msg.FromPeer.ID.String(),
				Name:   msg.FromPeer.Name,
				Avatar: msg.FromPeer.Avatar,
			},
			MsgType:    encodeMsgType(msg.MsgType),
			MimeType:   msg.MimeType,
			Payload:    msg.Payload,
			CreateTime: msg.CreateTime,
		}
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

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	err = groupSvc.ClearGroupMessage(ctx, request.GroupId)
	if err != nil {
		return nil, fmt.Errorf("svc clear group message error: %w", err)
	}

	reply = &proto.ClearGroupMessageReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}
	return reply, nil
}
