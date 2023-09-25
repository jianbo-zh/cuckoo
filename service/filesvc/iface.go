package filesvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/types"
)

type FileServiceIface interface {
	CalcFileHash(ctx context.Context, filePath string) (*types.FileHash, error)
	Close()
}
