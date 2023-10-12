package service

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
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

func (f *FileSvc) ConvertFileToResource(ctx context.Context, request *proto.ConvertFileToResourceRequest) (reply *proto.ConvertFileToResourceReply, err error) {

	log.Infoln("ConvertFileToResource request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("ConvertFileToResource panic: ", e)
		} else if err != nil {
			log.Errorln("ConvertFileToResource error: ", err.Error())
		} else {
			log.Infoln("ConvertFileToResource reply: ", reply.String())
		}
	}()

	fileSvc, err := f.getFileSvc()
	if err != nil {
		return nil, fmt.Errorf("get file svc error: %w", err)
	}

	resourceID, err := fileSvc.ConvertFileToResource(ctx, request.SrcFile)
	if err != nil {
		return nil, fmt.Errorf("svc convertFileToResource error: %w", err)
	}

	return &proto.ConvertFileToResourceReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Resource: resourceID,
	}, nil
}

// DownloadContactFile 下载联系人文件
func (f *FileSvc) DownloadContactFile(ctx context.Context, request *proto.DownloadContactFileRequest) (reply *proto.DownloadContactFileReply, err error) {

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

	cuckoo, err := f.getter.GetCuckoo()
	if err != nil {
		return nil, fmt.Errorf("getter.GetCuckoo error: %s", err.Error())
	}

	fileSvc, err := cuckoo.GetFileSvc()
	if err != nil {
		return nil, fmt.Errorf("cuckoo.GetPeerSvc error: %s", err.Error())
	}

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get file svc error: %w", err)
	}

	contactID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer.Decode error: %w", err)
	}

	msg, err := contactSvc.GetMessage(ctx, contactID, request.MsgId)
	if err != nil {
		return nil, fmt.Errorf("svc.GetMessage error: %w", err)
	}

	file, err := parseContactMessageFile(msg)
	if err != nil {
		return nil, fmt.Errorf("parse contact message error: %w", err)
	}

	if request.FileId != file.FileID {
		return nil, fmt.Errorf("file id error: %w", err)
	}

	err = fileSvc.DownloadContactFile(ctx, contactID, file)
	if err != nil {
		return nil, fmt.Errorf("svc download file error: %w", err)
	}

	return &proto.DownloadContactFileReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		FileId: request.FileId,
	}, nil
}

// DownloadGroupFile 下载群组文件
func (f *FileSvc) DownloadGroupFile(ctx context.Context, request *proto.DownloadGroupFileRequest) (reply *proto.DownloadGroupFileReply, err error) {

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

	peerIDs, err := groupSvc.GetGroupOnlineMemberIDs(ctx, request.GroupId)
	if err != nil {
		return nil, fmt.Errorf("svc.GetGroupOnlineMemberIDs error: %w", err)
	}

	msg, err := groupSvc.GetGroupMessage(ctx, request.GroupId, request.MsgId)
	if err != nil {
		return nil, fmt.Errorf("svc.GetMessage error: %w", err)
	}

	file, err := parseGroupMessageFile(msg)
	if err != nil {
		return nil, fmt.Errorf("parse contact message error: %w", err)
	}

	err = fileSvc.DownloadGroupFile(ctx, request.GroupId, peerIDs, file)
	if err != nil {
		return nil, fmt.Errorf("svc.DownloadFile error: %w", err)
	}

	return &proto.DownloadGroupFileReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		FileId: request.FileId,
	}, nil
}

// GetContactFiles 获取联系人相关文件
func (f *FileSvc) GetContactFiles(ctx context.Context, request *proto.GetContactFilesRequest) (reply *proto.GetContactFilesReply, err error) {

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

	fileSvc, err := f.getFileSvc()
	if err != nil {
		return nil, fmt.Errorf("get file svc error: %w", err)
	}

	contactID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
	}

	files, err := fileSvc.GetContactFiles(ctx, contactID, request.Keywords, int(request.Offset), int(request.Limit))
	if err != nil {
		return nil, fmt.Errorf("svc.GetContactFiles error: %w", err)
	}

	var fileList []*proto.FileInfo
	for _, file := range files {
		fileList = append(fileList, encodeFileInfo(&file))
	}

	return &proto.GetContactFilesReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Files: fileList,
	}, nil
}

// GetGroupFiles 群组相关文件
func (f *FileSvc) GetGroupFiles(ctx context.Context, request *proto.GetGroupFilesRequest) (reply *proto.GetGroupFilesReply, err error) {

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

	fileSvc, err := f.getFileSvc()
	if err != nil {
		return nil, fmt.Errorf("get file svc error: %w", err)
	}

	files, err := fileSvc.GetGroupFiles(ctx, request.GroupId, request.Keywords, int(request.Offset), int(request.Limit))
	if err != nil {
		return nil, fmt.Errorf("svc.GetContactFiles error: %w", err)
	}

	var fileList []*proto.FileInfo
	for _, file := range files {
		fileList = append(fileList, encodeFileInfo(&file))
	}

	return &proto.GetGroupFilesReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Files: fileList,
	}, nil
}

// 删除联系人文件
func (f *FileSvc) DeleteContactFile(ctx context.Context, request *proto.DeleteContactFileRequest) (reply *proto.DeleteContactFileReply, err error) {

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

	fileSvc, err := f.getFileSvc()
	if err != nil {
		return nil, fmt.Errorf("get file svc error: %w", err)
	}

	contactID, err := peer.Decode(request.ContactId)
	if err != nil {
		return nil, fmt.Errorf("peer decode error: %w", err)
	}

	if err := fileSvc.DeleteContactFile(ctx, contactID, request.FileIds); err != nil {
		return nil, fmt.Errorf("svc.DeleteContactFile error: %w", err)
	}

	return &proto.DeleteContactFileReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}, nil
}
