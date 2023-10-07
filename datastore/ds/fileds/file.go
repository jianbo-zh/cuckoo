package fileds

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/datastore"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/filepb"
	"google.golang.org/protobuf/proto"
)

var _ FileIface = (*FileDataStore)(nil)

var fileDsKey = &datastore.FileDsKey{}

func FileWrap(d ipfsds.Batching) *FileDataStore {
	return &FileDataStore{Batching: d}
}

type FileDataStore struct {
	ipfsds.Batching
}

func (f *FileDataStore) SaveFile(ctx context.Context, filePath string, fileSize int64, hashAlgo string, hashValue string) error {

	bs, err := proto.Marshal(&pb.File{
		HashAlgo:  hashAlgo,
		HashValue: hashValue,
		FilePath:  filePath,
		FileSize:  fileSize,
	})
	if err != nil {
		return fmt.Errorf("proto marshal error: %w", err)
	}

	if err := f.Put(ctx, fileDsKey.FileKey(hashAlgo, hashValue), bs); err != nil {
		return fmt.Errorf("ds put error: %w", err)
	}

	return nil
}

func (f *FileDataStore) GetFile(ctx context.Context, hashAlgo string, hashValue string) (*pb.File, error) {

	bs, err := f.Get(ctx, fileDsKey.FileKey(hashAlgo, hashValue))
	if err != nil {
		return nil, fmt.Errorf("ds get file: %w", err)
	}

	var file pb.File
	if err = proto.Unmarshal(bs, &file); err != nil {
		return nil, fmt.Errorf("proto unmarshal error: %w", err)
	}

	return &file, nil
}
