package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/internal/util"
	"github.com/jianbo-zh/dchat/service/accountsvc"
	"github.com/jianbo-zh/dchat/service/depositsvc"
	"github.com/jianbo-zh/dchat/service/filesvc"
	"github.com/jianbo-zh/dchat/service/groupsvc"
	"github.com/libp2p/go-libp2p/core/peer"
	goproto "google.golang.org/protobuf/proto"
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

func (g *GroupSvc) getFileSvc() (filesvc.FileServiceIface, error) {
	cuckoo, err := g.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	fileSvc, err := cuckoo.GetFileSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return fileSvc, nil
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

	cuckoo, err := g.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	groupSvc, err := cuckoo.GetGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	fileSvc, err := cuckoo.GetFileSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetFileSvc error: %s", err.Error())
	}

	avatarID, err := fileSvc.CopyFileToResource(ctx, request.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("svc.CopyFileToResource error: %w", err)
	}

	var memberIDs []peer.ID
	for _, pid := range request.MemberIds {
		peerID, err := peer.Decode(pid)
		if err != nil {
			return nil, fmt.Errorf("peer.Decode error: %w", err)
		}
		memberIDs = append(memberIDs, peerID)
	}

	_, err = groupSvc.CreateGroup(ctx, request.Name, avatarID, request.Content, memberIDs)
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

	cuckoo, err := g.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	groupSvc, err := cuckoo.GetGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("g.getGroupSvc error: %w", err)
	}

	members, err := groupSvc.GetGroupMembers(ctx, request.GroupId, request.Keywords, int(request.Offset), int(request.Limit))

	memberList := make([]*proto.GroupMember, len(members))

	if len(members) > 0 {
		onlineMemberIDs, err := groupSvc.GetGroupOnlineMemberIDs(ctx, request.GroupId)
		if err != nil {
			return nil, fmt.Errorf("svc.GetGroupOnlineMemberIDs error: %w", err)
		}

		onlineMemberIDsMap := make(map[peer.ID]struct{})
		for _, onlineID := range onlineMemberIDs {
			onlineMemberIDsMap[onlineID] = struct{}{}
		}

		for i, member := range members {
			onlineState := proto.ConnState_OfflineState
			if _, exists := onlineMemberIDsMap[member.ID]; exists {
				onlineState = proto.ConnState_OnlineState
			}

			memberList[i] = &proto.GroupMember{
				Id:          member.ID.String(),
				Name:        member.Name,
				Avatar:      member.Avatar,
				OnlineState: onlineState,
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

// SendGroupTextMessage 发送文本消息
func (g *GroupSvc) SendGroupTextMessage(request *proto.SendGroupTextMessageRequest, server proto.GroupSvc_SendGroupTextMessageServer) (err error) {
	log.Infoln("SendGroupTextMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendGroupTextMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SendGroupTextMessage error: ", err.Error())
		}
	}()

	return g.sendGroupMessage(context.Background(), server, mytype.TextMsgType, request.GroupId, "text/plain", []byte(request.Content), "", nil)
}

// SendGroupImageMessage 发送图片消息
func (g *GroupSvc) SendGroupImageMessage(request *proto.SendGroupImageMessageRequest, server proto.GroupSvc_SendGroupImageMessageServer) (err error) {
	log.Infoln("SendGroupImageMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendGroupImageMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SendGroupImageMessage error: ", err.Error())
		}
	}()

	fileSvc, err := g.getFileSvc()
	if err != nil {
		return fmt.Errorf("get file svc error: %w", err)
	}

	ctx := context.Background()
	thumbnailID, err := fileSvc.CopyFileToResource(ctx, request.ThumbnailPath)
	if err != nil {
		return fmt.Errorf("copy file to resource error: %w", err)
	}

	imageID, err := fileSvc.CopyFileToFile(ctx, request.FilePath)
	if err != nil {
		return fmt.Errorf("copy file to file error: %w", err)
	}

	payload, err := goproto.Marshal(&proto.ImageMessagePayload{
		ThumbnailId: thumbnailID,
		ImageId:     imageID,
		Name:        request.Name,
		Size:        request.Size,
		Width:       request.Width,
		Height:      request.Height,
	})
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	file := mytype.FileInfo{
		FileID:      imageID,
		FileName:    request.Name,
		FileSize:    request.Size,
		FileType:    mytype.ImageFile,
		MimeType:    request.MimeType,
		ThumbnailID: thumbnailID,
		Width:       request.Width,
		Height:      request.Height,
	}

	return g.sendGroupMessage(ctx, server, mytype.ImageMsgType, request.GroupId, request.MimeType, payload, thumbnailID, &file)
}

// SendGroupVoiceMessage 发送语音消息
func (g *GroupSvc) SendGroupVoiceMessage(request *proto.SendGroupVoiceMessageRequest, server proto.GroupSvc_SendGroupVoiceMessageServer) (err error) {
	log.Infoln("SendGroupVoiceMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendGroupVoiceMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SendGroupVoiceMessage error: ", err.Error())
		}
	}()

	fileSvc, err := g.getFileSvc()
	if err != nil {
		return fmt.Errorf("get file svc error: %w", err)
	}

	ctx := context.Background()
	voiceID, err := fileSvc.CopyFileToResource(ctx, request.FilePath)
	if err != nil {
		return fmt.Errorf("copy file to resource error: %w", err)
	}

	payload, err := goproto.Marshal(&proto.VoiceMessagePayload{
		VoiceId:  voiceID,
		Duration: request.Duration,
	})
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	return g.sendGroupMessage(ctx, server, mytype.VoiceMsgType, request.GroupId, request.MimeType, payload, voiceID, nil)
}

// SendGroupAudioMessage 发送音频消息
func (g *GroupSvc) SendGroupAudioMessage(request *proto.SendGroupAudioMessageRequest, server proto.GroupSvc_SendGroupAudioMessageServer) (err error) {
	log.Infoln("SendGroupAudioMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendGroupAudioMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SendGroupAudioMessage error: ", err.Error())
		}
	}()

	fileSvc, err := g.getFileSvc()
	if err != nil {
		return fmt.Errorf("get file svc error: %w", err)
	}

	ctx := context.Background()
	audioID, err := fileSvc.CopyFileToFile(ctx, request.FilePath)
	if err != nil {
		return fmt.Errorf("copy file to resource error: %w", err)
	}

	payload, err := goproto.Marshal(&proto.AudioMessagePayload{
		AudioId:  audioID,
		Name:     request.Name,
		Size:     request.Size,
		Duration: request.Duration,
	})
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	file := mytype.FileInfo{
		FileID:   audioID,
		FileName: request.Name,
		FileSize: request.Size,
		FileType: mytype.AudioFile,
		MimeType: request.MimeType,
		Duration: request.Duration,
	}

	return g.sendGroupMessage(ctx, server, mytype.AudioMsgType, request.GroupId, request.MimeType, payload, "", &file)
}

// SendGroupVideoMessage 发送视频消息
func (g *GroupSvc) SendGroupVideoMessage(request *proto.SendGroupVideoMessageRequest, server proto.GroupSvc_SendGroupVideoMessageServer) (err error) {
	log.Infoln("SendGroupVideoMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendGroupVideoMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SendGroupVideoMessage error: ", err.Error())
		}
	}()

	fileSvc, err := g.getFileSvc()
	if err != nil {
		return fmt.Errorf("get file svc error: %w", err)
	}

	ctx := context.Background()
	videoID, err := fileSvc.CopyFileToFile(ctx, request.FilePath)
	if err != nil {
		return fmt.Errorf("copy file to resource error: %w", err)
	}

	payload, err := goproto.Marshal(&proto.VideoMessagePayload{
		VideoId:  videoID,
		Name:     request.Name,
		Size:     request.Size,
		Duration: request.Duration,
	})
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	file := mytype.FileInfo{
		FileID:   videoID,
		FileName: request.Name,
		FileSize: request.Size,
		FileType: mytype.VideoFile,
		MimeType: request.MimeType,
		Duration: request.Duration,
	}

	return g.sendGroupMessage(ctx, server, mytype.VideoMsgType, request.GroupId, request.MimeType, payload, "", &file)
}

// SendGroupFileMessage 发送文件消息
func (g *GroupSvc) SendGroupFileMessage(request *proto.SendGroupFileMessageRequest, server proto.GroupSvc_SendGroupFileMessageServer) (err error) {
	log.Infoln("SendGroupFileMessage request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("SendGroupFileMessage panic: ", e)
		} else if err != nil {
			log.Errorln("SendGroupFileMessage error: ", err.Error())
		}
	}()

	fileSvc, err := g.getFileSvc()
	if err != nil {
		return fmt.Errorf("get group svc error: %w", err)
	}

	ctx := context.Background()
	fileID, err := fileSvc.CopyFileToFile(ctx, request.FilePath)
	if err != nil {
		return fmt.Errorf("svc.CopyFileToFile error: %w", err)
	}

	payload, err := goproto.Marshal(&proto.FileMessagePayload{
		FileId: fileID,
		Name:   request.Name,
		Size:   request.Size,
	})
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	file := mytype.FileInfo{
		FileID:   fileID,
		FileName: request.Name,
		FileSize: request.Size,
		FileType: mytype.OtherFile,
		MimeType: request.MimeType,
	}

	return g.sendGroupMessage(ctx, server, mytype.FileMsgType, request.GroupId, request.MimeType, payload, "", &file)
}

func (g *GroupSvc) sendGroupMessage(ctx context.Context, server proto.GroupSvc_SendGroupTextMessageServer,
	msgType string, groupID string, mimeType string, payload []byte, resourceID string, file *mytype.FileInfo) (err error) {

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return fmt.Errorf("get group svc error: %w", err)
	}

	accountSvc, err := g.getAccountSvc()
	if err != nil {
		return fmt.Errorf("get account service error: %w", err)
	}

	account, err := accountSvc.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("svc get account error: %w", err)
	}

	resultCh, err := groupSvc.SendGroupMessage(ctx, groupID, msgType, mimeType, payload, resourceID, file)
	if err != nil {
		return fmt.Errorf("svc.SendGroupMessage error: %w", err)
	}

	i := 0
	for msg := range resultCh {
		i++
		reply := proto.SendGroupMessageReply{
			Result: &proto.Result{
				Code:    0,
				Message: "ok",
			},
			IsUpdated: i > 1,
			Message: &proto.GroupMessage{
				Id:      msg.ID,
				GroupId: groupID,
				Sender: &proto.Peer{
					Id:     account.ID.String(),
					Name:   account.Name,
					Avatar: account.Avatar,
				},
				MsgType:    encodeMsgType(msg.MsgType),
				MimeType:   msg.MimeType,
				Payload:    msg.Payload,
				IsDeposit:  msg.IsDeposit,
				State:      encodeMessageState(msg.State),
				CreateTime: msg.Timestamp,
			},
		}
		if err := server.Send(&reply); err != nil {
			return fmt.Errorf("server.Send error: %w", err)
		}

		log.Infoln("sendGroupMessage reply: ", reply.String())
	}

	return nil
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

	cuckoo, err := g.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	groupSvc, err := cuckoo.GetGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	sessionSvc, err := cuckoo.GetSessionSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetSessionSvc error: %s", err.Error())
	}

	sessionID := mytype.GroupSessionID(request.GroupId)
	if err = sessionSvc.ResetUnreads(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("svc reset unreads error: %w", err)
	}
	if err = sessionSvc.UpdateSessionTime(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("svc update session time error: %w", err)
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
			IsDeposit:  msg.IsDeposit,
			State:      encodeMessageState(msg.State),
			CreateTime: msg.Timestamp,
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

	cuckoo, err := g.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	groupSvc, err := cuckoo.GetGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	sessionSvc, err := cuckoo.GetSessionSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetSessionSvc error: %s", err.Error())
	}

	sessionID := mytype.GroupSessionID(request.GroupId)
	if err = sessionSvc.ResetUnreads(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("svc reset unreads error: %w", err)
	}
	if err = sessionSvc.UpdateSessionTime(ctx, sessionID.String()); err != nil {
		return nil, fmt.Errorf("svc update session time error: %w", err)
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
			IsDeposit:  msg.IsDeposit,
			State:      encodeMessageState(msg.State),
			CreateTime: msg.Timestamp,
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

func (g *GroupSvc) GetGroupQRCodeToken(ctx context.Context, request *proto.GetGroupQRCodeTokenRequest) (reply *proto.GetGroupQRCodeTokenReply, err error) {

	log.Infoln("GetGroupQRCodeToken request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetGroupQRCodeToken panic: ", e)
		} else if err != nil {
			log.Errorln("GetGroupQRCodeToken error: ", err.Error())
		} else {
			log.Infoln("GetGroupQRCodeToken reply: ", reply.String())
		}
	}()

	groupSvc, err := g.getGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("get account svc error: %w", err)
	}

	group, err := groupSvc.GetGroup(ctx, request.GroupId)
	if err != nil {
		return nil, fmt.Errorf("svc.GetContact error: %w", err)
	}

	QRCodeToken := util.EncodeGroupToken(group.ID, group.DepositAddress, group.Name, group.Avatar)

	reply = &proto.GetGroupQRCodeTokenReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Token: QRCodeToken,
	}
	return reply, nil
}
