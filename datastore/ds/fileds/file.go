package fileds

import (
	"context"
	"fmt"
	"strings"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
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

func (f *FileDataStore) GetFileSessionIDs(ctx context.Context, fileID string) ([]string, error) {
	prefix := fileDsKey.FileSessionPrefix(fileID)
	results, err := f.Query(ctx, query.Query{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var sessionIDs []string
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("results.Next error: %w", result.Error)
		}

		sessionIDs = append(sessionIDs, strings.TrimPrefix(result.Key, prefix))
	}

	return sessionIDs, nil
}

func (f *FileDataStore) SaveSessionResource(ctx context.Context, sessionID string, resourceID string) error {
	batch, err := f.Batch(ctx)
	if err != nil {
		return fmt.Errorf("ds batch error: %w", err)
	}

	if err = batch.Put(ctx, fileDsKey.SessionResourceKey(sessionID, resourceID), []byte("true")); err != nil {
		return fmt.Errorf("batch.Put session resource key error: %w", err)
	}

	if err = batch.Put(ctx, fileDsKey.FileSessionKey(resourceID, sessionID), []byte("true")); err != nil {
		return fmt.Errorf("batch.Put file session key error: %w", err)
	}

	if err = batch.Commit(ctx); err != nil {
		return fmt.Errorf("batch.Commit error: %w", err)
	}

	return nil
}

func (f *FileDataStore) GetSessionResourceIDs(ctx context.Context, sessionID string) ([]string, error) {
	prefix := fileDsKey.SessionResourcePrefix(sessionID)
	results, err := f.Query(ctx, query.Query{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var resourceIDs []string
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("results.Next error: %w", result.Error)
		}

		resourceIDs = append(resourceIDs, strings.TrimPrefix(result.Key, prefix))
	}

	return resourceIDs, nil
}

func (f *FileDataStore) SaveSessionUploadFile(ctx context.Context, sessionID string, uploadFile *pb.FileInfo) error {
	batch, err := f.Batch(ctx)
	if err != nil {
		return fmt.Errorf("ds batch error: %w", err)
	}

	bs, err := proto.Marshal(uploadFile)
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	if err = batch.Put(ctx, fileDsKey.SessionUploadFileKey(sessionID, uploadFile.Id), bs); err != nil {
		return fmt.Errorf("batch.Put session upload file key error: %w", err)
	}

	if err = batch.Put(ctx, fileDsKey.FileSessionKey(uploadFile.Id, sessionID), []byte("true")); err != nil {
		return fmt.Errorf("batch.Put file session key error: %w", err)
	}

	if err = batch.Commit(ctx); err != nil {
		return fmt.Errorf("batch.Commit error: %w", err)
	}

	return nil
}

func (f *FileDataStore) GetSessionUploadFiles(ctx context.Context, sessionID string) ([]*pb.FileInfo, error) {
	results, err := f.Query(ctx, query.Query{
		Prefix: fileDsKey.SessionUploadFilePrefix(sessionID),
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var files []*pb.FileInfo
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("results.Next error: %w", result.Error)
		}

		var file pb.FileInfo
		if err := proto.Unmarshal(result.Value, &file); err != nil {
			return nil, fmt.Errorf("proto.Unmarshal error: %w", err)
		}

		files = append(files, &file)
	}

	return files, nil
}

func (f *FileDataStore) SaveSessionDownloadFile(ctx context.Context, sessionID string, downloadFile *pb.FileInfo) error {
	batch, err := f.Batch(ctx)
	if err != nil {
		return fmt.Errorf("ds batch error: %w", err)
	}

	bs, err := proto.Marshal(downloadFile)
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	if err = batch.Put(ctx, fileDsKey.SessionDownloadFileKey(sessionID, downloadFile.Id), bs); err != nil {
		return fmt.Errorf("batch.Put session download file key error: %w", err)
	}

	if err = batch.Put(ctx, fileDsKey.FileSessionKey(downloadFile.Id, sessionID), []byte("true")); err != nil {
		return fmt.Errorf("batch.Put file session key error: %w", err)
	}

	if err = batch.Commit(ctx); err != nil {
		return fmt.Errorf("batch.Commit error: %w", err)
	}

	return nil
}

func (f *FileDataStore) GetSessionDownloadFiles(ctx context.Context, sessionID string) ([]*pb.FileInfo, error) {

	results, err := f.Query(ctx, query.Query{
		Prefix: fileDsKey.SessionDownloadFilePrefix(sessionID),
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var files []*pb.FileInfo
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("results.Next error: %w", result.Error)
		}

		var file pb.FileInfo
		if err := proto.Unmarshal(result.Value, &file); err != nil {
			return nil, fmt.Errorf("proto.Unmarshal error: %w", err)
		}

		files = append(files, &file)
	}

	return files, nil
}
