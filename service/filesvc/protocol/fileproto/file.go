package fileproto

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/service/filesvc/protocol/fileproto/ds"
)

type FileProto struct {
	data ds.FileIface
}

func NewFileProto(ids ipfsds.Batching) (*FileProto, error) {
	fileProto := FileProto{
		data: ds.FileWrap(ids),
	}

	return &fileProto, nil
}

func (f *FileProto) CalcFileHash(ctx context.Context, filePath string) (*mytype.FileHash, error) {
	// 计算hash
	ofile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("os open file error: %w", err)
	}
	defer ofile.Close()

	fi, err := ofile.Stat()
	if err != nil {
		return nil, fmt.Errorf("file state error: %w", err)

	} else if fi.IsDir() {
		return nil, fmt.Errorf("file is dir, need file")
	}

	hashSum := md5.New()

	size, err := io.Copy(hashSum, ofile)
	if err != nil {
		return nil, fmt.Errorf("io copy calc file hash error: %w", err)
	}

	if fi.Size() != size {
		return nil, fmt.Errorf("io copy size error, filesize: %d, copysize: %d", fi.Size(), size)
	}

	fileHash := mytype.FileHash{
		HashAlgo:  "md5",
		HashValue: fmt.Sprintf("%x", hashSum.Sum(nil)),
	}

	// 存储结果
	err = f.data.SaveFile(ctx, filePath, fi.Size(), fileHash.HashAlgo, fileHash.HashValue)
	if err != nil {
		return nil, fmt.Errorf("data save file error: %w", err)
	}

	return &fileHash, nil
}
