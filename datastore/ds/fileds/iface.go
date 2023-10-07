package fileds

import (
	"context"

	pb "github.com/jianbo-zh/dchat/protobuf/pb/filepb"
)

type FileIface interface {
	SaveFile(ctx context.Context, filePath string, fileSize int64, hashAlgo string, hashValue string) error
	GetFile(ctx context.Context, hashAlgo string, hashValue string) (*pb.File, error)
}
