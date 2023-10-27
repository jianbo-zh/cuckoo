package filesvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/service/filesvc/protocols/fileproto"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
)

var log = logging.Logger("filesvc")

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

func (f *FileService) CopyFileToResource(ctx context.Context, srcFile string) (string, error) {
	return f.fileProto.CopyFileToResource(ctx, srcFile)
}

func (f *FileService) CopyFileToFile(ctx context.Context, srcFile string) (string, error) {
	return f.fileProto.CopyFileToFile(ctx, srcFile)
}

// DownloadSessionFile 下载会话文件
func (f *FileService) DownloadSessionFile(ctx context.Context, sessionID string, peerIDs []peer.ID, file *mytype.FileInfo) error {
	return f.fileProto.DownloadFile(ctx, sessionID, peerIDs, file)
}

// GetSessionFiles 获取会话文件
func (f *FileService) GetSessionFiles(ctx context.Context, sessionID string, keywords string, offset int, limit int) ([]mytype.FileInfo, error) {
	files, err := f.fileProto.GetSessionFiles(ctx, sessionID, keywords, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("proto.GetSessionFiles error: %w", err)
	}

	var fileList []mytype.FileInfo
	for _, file := range files {
		fileList = append(fileList, *decodeFile(file))
	}

	return fileList, nil
}

// DeleteSessionFile 删除会话文件
func (f *FileService) DeleteSessionFile(ctx context.Context, sessionID string, fileIDs []string) error {
	return f.fileProto.DeleteSessionFiles(ctx, sessionID, fileIDs)
}

func (f *FileService) DownloadAvatar(ctx context.Context, peerID peer.ID, avatar string) error {
	return f.fileProto.DownloadResource(ctx, peerID, avatar)
}

func (f *FileService) Close() {}

// // CalcFileID 计算文件Hash并保存
// func (f *FileService) CalcFileID(ctx context.Context, filePath string) (*mytype.FileID, error) {
// 	return f.fileProto.CalcFileID(ctx, filePath)
// }
