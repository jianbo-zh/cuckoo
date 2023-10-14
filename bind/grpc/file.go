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

// DownloadGroupFile 下载群组文件
func (f *FileSvc) DownloadSessionFile(ctx context.Context, request *proto.DownloadSessionFileRequest) (reply *proto.DownloadSessionFileReply, err error) {

	log.Infoln("DownloadSessionFile request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("DownloadSessionFile panic: ", e)
		} else if err != nil {
			log.Errorln("DownloadSessionFile error: ", err.Error())
		} else {
			log.Infoln("DownloadSessionFile reply: ", reply.String())
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

	contactSvc, err := cuckoo.GetContactSvc()
	if err != nil {
		return nil, fmt.Errorf("get file svc error: %w", err)
	}

	sessionID, err := mytype.DecodeSessionID(request.SessionId)
	if err != nil {
		return nil, fmt.Errorf("DecodeSessionID error: %w", err)
	}

	var peerIDs []peer.ID
	var file *mytype.FileInfo

	switch sessionID.Type {
	case mytype.ContactSession:
		contactID := peer.ID(sessionID.Value)
		peerIDs = []peer.ID{contactID}
		msg, err := contactSvc.GetMessage(ctx, contactID, request.MsgId)
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

		msg, err := groupSvc.GetGroupMessage(ctx, groupID, request.MsgId)
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

	err = fileSvc.DownloadSessionFile(ctx, request.SessionId, peerIDs, file)
	if err != nil {
		return nil, fmt.Errorf("svc.DownloadFile error: %w", err)
	}

	return &proto.DownloadSessionFileReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		FileId: request.FileId,
	}, nil
}

// GetGroupFiles 群组相关文件
func (f *FileSvc) GetSessionFiles(ctx context.Context, request *proto.GetSessionFilesRequest) (reply *proto.GetSessionFilesReply, err error) {

	log.Infoln("GetSessionFiles request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("GetSessionFiles panic: ", e)
		} else if err != nil {
			log.Errorln("GetSessionFiles error: ", err.Error())
		} else {
			log.Infoln("GetSessionFiles reply: ", reply.String())
		}
	}()

	fileSvc, err := f.getFileSvc()
	if err != nil {
		return nil, fmt.Errorf("get file svc error: %w", err)
	}

	files, err := fileSvc.GetSessionFiles(ctx, request.SessionId, request.Keywords, int(request.Offset), int(request.Limit))
	if err != nil {
		return nil, fmt.Errorf("svc.GetContactFiles error: %w", err)
	}

	var fileList []*proto.FileInfo
	for _, file := range files {
		fileList = append(fileList, encodeFileInfo(&file))
	}

	return &proto.GetSessionFilesReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
		Files: fileList,
	}, nil
}

// 删除联系人文件
func (f *FileSvc) DeleteSessionFile(ctx context.Context, request *proto.DeleteSessionFileRequest) (reply *proto.DeleteSessionFileReply, err error) {

	log.Infoln("DeleteSessionFile request: ", request.String())
	defer func() {
		if e := recover(); e != nil {
			log.Panicln("DeleteSessionFile panic: ", e)
		} else if err != nil {
			log.Errorln("DeleteSessionFile error: ", err.Error())
		} else {
			log.Infoln("DeleteSessionFile reply: ", reply.String())
		}
	}()

	fileSvc, err := f.getFileSvc()
	if err != nil {
		return nil, fmt.Errorf("get file svc error: %w", err)
	}

	if err := fileSvc.DeleteSessionFile(ctx, request.SessionId, request.FileIds); err != nil {
		return nil, fmt.Errorf("svc.DeleteContactFile error: %w", err)
	}

	return &proto.DeleteSessionFileReply{
		Result: &proto.Result{
			Code:    0,
			Message: "ok",
		},
	}, nil
}
