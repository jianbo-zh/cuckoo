package service

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/jianbo-zh/dchat/service/filesvc"
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
