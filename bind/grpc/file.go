package service

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/service/filesvc"
	"github.com/libp2p/go-libp2p/core/peer"
)

var _ proto.FileSvcServer = (*FileSvc)(nil)

type FileSvc struct {
	getter cuckoo.CuckooGetter
	proto.UnimplementedFileSvcServer
}

func NewFileSvc(getter cuckoo.CuckooGetter) *FileSvc {
	return &FileSvc{
		getter: getter,
	}
}

func (f *FileSvc) getFileSvc() (filesvc.FileServiceIface, error) {
	cuckoo, err := f.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	fileSvc, err := cuckoo.GetFileSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	return fileSvc, nil
}

// DownloadContactFile 下载联系人文件
func (f *FileSvc) DownloadContactFile(ctx context.Context, request *proto.DownloadContactFileRequest) (reply *proto.DownloadFileReply, err error) {

	log.Infoln("DownloadContactFile request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("DownloadContactFile panic: ", e)
		} else if err != nil {
			log.Errorln("DownloadContactFile error: ", err.Error())
		} else {
			log.Infoln("DownloadContactFile reply: ", reply.String())
		}
	}()

	contactID, err := peer.Decode(request.Id)
	if err != nil {
		return nil, fmt.Errorf("peer.Decode error: %w", err)
	}

	sessionID := mytype.ContactSessionID(contactID)

	if reply, err = f.downloadSessionFile(ctx, sessionID, request.MsgId, request.FileId); err != nil {
		return nil, fmt.Errorf("download session file error: %w")
	}

	return reply, nil
}

// DownloadGroupFile 下载群组文件
func (f *FileSvc) DownloadGroupFile(ctx context.Context, request *proto.DownloadGroupFileRequest) (reply *proto.DownloadFileReply, err error) {

	log.Infoln("DownloadGroupFile request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("DownloadGroupFile panic: ", e)
		} else if err != nil {
			log.Errorln("DownloadGroupFile error: ", err.Error())
		} else {
			log.Infoln("DownloadGroupFile reply: ", reply.String())
		}
	}()

	sessionID := mytype.GroupSessionID(request.Id)

	if reply, err = f.downloadSessionFile(ctx, sessionID, request.MsgId, request.FileId); err != nil {
		return nil, fmt.Errorf("download session file error: %w")
	}

	return reply, nil
}

func (f *FileSvc) downloadSessionFile(ctx context.Context, sessionID mytype.SessionID, msgID string, fileID string) (reply *proto.DownloadFileReply, err error) {

	cuckoo, err := f.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	fileSvc, err := cuckoo.GetFileSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetFileSvc error: %s", err.Error())
	}

	groupSvc, err := cuckoo.GetGroupSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetGroupSvc error: %s", err.Error())
	}

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get file svc error: %w", err)
	}

	var peerIDs []peer.ID
	var file *mytype.FileInfo

	switch sessionID.Type {
	case mytype.ContactSession:
		contactID := peer.ID(sessionID.Value)
		peerIDs = []peer.ID{contactID}
		msg, err := contactSvc.GetMessage(ctx, contactID, msgID)
		if err != nil {
			return nil, fmt.Errorf("svc.GetMessage error: %w", err)
		}

		file, err = parseContactMessageFile(msg)
		if err != nil {
			return nil, fmt.Errorf("parse contact message error: %w", err)
		}

	case mytype.GroupSession:
		groupID := string(sessionID.Value)
		peerIDs, err = groupSvc.GetGroupOnlineMemberIDs(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("svc.GetGroupOnlineMemberIDs error: %w", err)
		}

		msg, err := groupSvc.GetGroupMessage(ctx, groupID, msgID)
		if err != nil {
			return nil, fmt.Errorf("svc.GetMessage error: %w", err)
		}

		file, err = parseGroupMessageFile(msg)
		if err != nil {
			return nil, fmt.Errorf("parse contact message error: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupport session type")
	}

	err = fileSvc.DownloadSessionFile(ctx, sessionID.String(), peerIDs, file)
	if err != nil {
		return nil, fmt.Errorf("svc.DownloadFile error: %w", err)
	}

	return &proto.DownloadFileReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		FileId: fileID,
	}, nil
}

// GetContactFiles 获取联系人文件
func (f *FileSvc) GetContactFiles(ctx context.Context, request *proto.GetContactFilesRequest) (reply *proto.GetFilesReply, err error) {

	log.Infoln("GetContactFiles request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetContactFiles panic: ", e)
		} else if err != nil {
			log.Errorln("GetContactFiles error: ", err.Error())
		} else {
			log.Infoln("GetContactFiles reply: ", reply.String())
		}
	}()

	contactID, err := peer.Decode(request.Id)
	if err != nil {
		return nil, fmt.Errorf("peer.Decode error: %w", err)
	}

	sessionID := mytype.ContactSessionID(contactID)

	if reply, err = f.getSessionFiles(ctx, sessionID, request.Keywords, request.Offset, request.Limit); err != nil {
		return nil, fmt.Errorf("get session files error: %w", err)
	}

	return reply, nil
}

// GetGroupFiles 获取群组文件
func (f *FileSvc) GetGroupFiles(ctx context.Context, request *proto.GetGroupFilesRequest) (reply *proto.GetFilesReply, err error) {

	log.Infoln("GetGroupFiles request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetGroupFiles panic: ", e)
		} else if err != nil {
			log.Errorln("GetGroupFiles error: ", err.Error())
		} else {
			log.Infoln("GetGroupFiles reply: ", reply.String())
		}
	}()

	sessionID := mytype.GroupSessionID(request.Id)

	if reply, err = f.getSessionFiles(ctx, sessionID, request.Keywords, request.Offset, request.Limit); err != nil {
		return nil, fmt.Errorf("get session files error: %w", err)
	}

	return reply, nil
}

func (f *FileSvc) getSessionFiles(ctx context.Context, sessionID mytype.SessionID, keywords string, offset int32, limit int32) (reply *proto.GetFilesReply, err error) {

	fileSvc, err := f.getFileSvc()
	if err != nil {
		return nil, fmt.Errorf("get file svc error: %w", err)
	}

	files, err := fileSvc.GetSessionFiles(ctx, sessionID.String(), keywords, int(offset), int(limit))
	if err != nil {
		return nil, fmt.Errorf("svc.GetContactFiles error: %w", err)
	}

	var fileList []*proto.FileInfo
	for _, file := range files {
		fileList = append(fileList, encodeFileInfo(&file))
	}

	return &proto.GetFilesReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Files: fileList,
	}, nil
}

func (f *FileSvc) DeleteContactFile(ctx context.Context, request *proto.DeleteContactFileRequest) (reply *proto.DeleteFileReply, err error) {

	log.Infoln("DeleteContactFile request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("DeleteContactFile panic: ", e)
		} else if err != nil {
			log.Errorln("DeleteContactFile error: ", err.Error())
		} else {
			log.Infoln("DeleteContactFile reply: ", reply.String())
		}
	}()

	contactID, err := peer.Decode(request.Id)
	if err != nil {
		return nil, fmt.Errorf("peer.Decode error: %w", err)
	}

	sessionID := mytype.ContactSessionID(contactID)

	if reply, err = f.deleteSessionFile(ctx, sessionID, request.FileIds); err != nil {
		return nil, fmt.Errorf("delete session files error: %w", err)
	}

	return reply, nil
}

func (f *FileSvc) DeleteGroupFile(ctx context.Context, request *proto.DeleteGroupFileRequest) (reply *proto.DeleteFileReply, err error) {

	log.Infoln("DeleteGroupFile request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("DeleteGroupFile panic: ", e)
		} else if err != nil {
			log.Errorln("DeleteGroupFile error: ", err.Error())
		} else {
			log.Infoln("DeleteGroupFile reply: ", reply.String())
		}
	}()

	sessionID := mytype.GroupSessionID(request.Id)

	if reply, err = f.deleteSessionFile(ctx, sessionID, request.FileIds); err != nil {
		return nil, fmt.Errorf("delete session files error: %w", err)
	}

	return reply, nil
}

func (f *FileSvc) deleteSessionFile(ctx context.Context, sessionID mytype.SessionID, fileIDs []string) (reply *proto.DeleteFileReply, err error) {

	fileSvc, err := f.getFileSvc()
	if err != nil {
		return nil, fmt.Errorf("get file svc error: %w", err)
	}

	if err := fileSvc.DeleteSessionFile(ctx, sessionID.String(), fileIDs); err != nil {
		return nil, fmt.Errorf("svc.DeleteContactFile error: %w", err)
	}

	return &proto.DeleteFileReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}, nil
}
