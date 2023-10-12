package filesvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/protocol/fileproto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
)

var _ FileServiceIface = (*FileService)(nil)

type FileService struct {
	fileProto *fileproto.FileProto
}

func NewFileService(ctx context.Context, conf config.FileServiceConfig, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus) (*FileService, error) {

	fileProto, err := fileproto.NewFileProto(conf, lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("new download client proto error: %w", err)
	}

	dlsvc := FileService{
		fileProto: fileProto,
	}

	return &dlsvc, nil
}

func (f *FileService) ConvertFileToResource(ctx context.Context, srcFile string) (string, error) {
	return f.fileProto.CopyFileToResource(ctx, srcFile)
}

func (f *FileService) DownloadContactFile(ctx context.Context, contactID peer.ID, file *mytype.FileInfo) error {
	sessionID := mytype.ContactSessionID(contactID)
	return f.fileProto.DownloadFile(ctx, sessionID.String(), []peer.ID{contactID}, file)
}

func (f *FileService) DownloadGroupFile(ctx context.Context, groupID string, peerIDs []peer.ID, file *mytype.FileInfo) error {
	sessionID := mytype.GroupSessionID(groupID)
	return f.fileProto.DownloadFile(ctx, sessionID.String(), peerIDs, file)
}

func (f *FileService) DownloadAvatar(ctx context.Context, peerID peer.ID, avatar string) error {
	return f.fileProto.DownloadResource(ctx, peerID, avatar)
}

func (f *FileService) GetContactFiles(ctx context.Context, contactID peer.ID, keywords string, offset int, limit int) ([]mytype.FileInfo, error) {
	sessionID := mytype.ContactSessionID(contactID)
	files, err := f.fileProto.GetSessionFiles(ctx, sessionID.String(), keywords, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("proto.GetSessionFiles error: %w", err)
	}

	var fileList []mytype.FileInfo
	for _, file := range files {
		fileList = append(fileList, *decodeFile(file))
	}

	return fileList, nil
}

func (f *FileService) GetGroupFiles(ctx context.Context, groupID string, keywords string, offset int, limit int) ([]mytype.FileInfo, error) {
	sessionID := mytype.GroupSessionID(groupID)
	files, err := f.fileProto.GetSessionFiles(ctx, sessionID.String(), keywords, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("proto.GetSessionFiles error: %w", err)
	}

	var fileList []mytype.FileInfo
	for _, file := range files {
		fileList = append(fileList, *decodeFile(file))
	}

	return fileList, nil
}

// DeleteContactFile 删除联系人文件
func (f *FileService) DeleteContactFile(ctx context.Context, contactID peer.ID, fileIDs []string) error {
	sessionID := mytype.ContactSessionID(contactID)
	return f.fileProto.DeleteSessionFiles(ctx, sessionID.String(), fileIDs)
}

func (f *FileService) Close() {}

// // CalcFileID 计算文件Hash并保存
// func (f *FileService) CalcFileID(ctx context.Context, filePath string) (*mytype.FileID, error) {
// 	return f.fileProto.CalcFileID(ctx, filePath)
// }
