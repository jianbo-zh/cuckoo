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
	fileProto     *fileproto.FileProto
	downloadProto *fileproto.DownloadProto
}

func NewFileService(ctx context.Context, conf config.FileServiceConfig, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus) (*FileService, error) {

	fileProto, err := fileproto.NewFileProto(ids)
	if err != nil {
		return nil, fmt.Errorf("new download client proto error: %w", err)
	}

	downloadProto, err := fileproto.NewDownloadProto(conf, lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("new download client proto error: %w", err)
	}

	dlsvc := FileService{
		fileProto:     fileProto,
		downloadProto: downloadProto,
	}

	return &dlsvc, nil
}

// CalcFileHash 计算文件Hash并保存
func (f *FileService) CalcFileHash(ctx context.Context, filePath string) (*mytype.FileHash, error) {
	return f.fileProto.CalcFileHash(ctx, filePath)
}

func (f *FileService) AvatarDownload(ctx context.Context, peerID peer.ID, avatar string) error {
	return f.downloadProto.AvatarDownload(ctx, peerID, avatar)
}

func (f *FileService) Close() {}
