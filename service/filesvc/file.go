package filesvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/internal/myhost"
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

// // CalcFileID 计算文件Hash并保存
// func (f *FileService) CalcFileID(ctx context.Context, filePath string) (*mytype.FileID, error) {
// 	return f.fileProto.CalcFileID(ctx, filePath)
// }

func (f *FileService) DownloadAvatar(ctx context.Context, peerID peer.ID, avatar string) error {
	return f.fileProto.DownloadResource(ctx, peerID, avatar)
}

func (f *FileService) SendPeerFile(ctx context.Context, peerID peer.ID, file string) error {
	return f.fileProto.SendPeerFile(ctx, peerID, file)
}

func (f *FileService) Close() {}
