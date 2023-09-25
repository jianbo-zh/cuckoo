package ds

import (
	"context"

	"github.com/jianbo-zh/dchat/service/filesvc/protocol/fileproto/pb"
)

type FileIface interface {
	SaveFile(ctx context.Context, filePath string, fileSize int64, hashAlgo string, hashValue string) error
	GetFile(ctx context.Context, hashAlgo string, hashValue string) (*pb.File, error)
}

type DownloadIface interface {
}
