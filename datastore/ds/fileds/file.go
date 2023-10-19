package fileds

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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

func (f *FileDataStore) RemoveSessionResource(ctx context.Context, sessionID string, resourceID string) error {
	batch, err := f.Batch(ctx)
	if err != nil {
		return fmt.Errorf("ds batch error: %w", err)
	}

	if err := batch.Delete(ctx, fileDsKey.SessionResourceKey(sessionID, resourceID)); err != nil {
		return fmt.Errorf("batch.Delete error: %w", err)
	}

	if err := batch.Delete(ctx, fileDsKey.FileSessionKey(resourceID, sessionID)); err != nil {
		return fmt.Errorf("batch.Delete error: %w", err)
	}

	return batch.Commit(ctx)
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

func (f *FileDataStore) SaveSessionFile(ctx context.Context, sessionID string, file *pb.FileInfo) error {
	batch, err := f.Batch(ctx)
	if err != nil {
		return fmt.Errorf("ds batch error: %w", err)
	}

	bs, err := proto.Marshal(file)
	if err != nil {
		return fmt.Errorf("proto.Marshal error: %w", err)
	}

	if err = batch.Put(ctx, fileDsKey.SessionFileKey(sessionID, file.FileId), bs); err != nil {
		return fmt.Errorf("batch.Put session upload file key error: %w", err)
	}

	if err = batch.Put(ctx, fileDsKey.SessionFileTimeKey(sessionID, file.FileId), []byte(strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
		return fmt.Errorf("batch.Put session upload file key error: %w", err)
	}

	if err = batch.Put(ctx, fileDsKey.FileSessionKey(file.FileId, sessionID), []byte("true")); err != nil {
		return fmt.Errorf("batch.Put file session key error: %w", err)
	}

	if err = batch.Commit(ctx); err != nil {
		return fmt.Errorf("batch.Commit error: %w", err)
	}

	return nil
}

func (f *FileDataStore) RemoveSessionFile(ctx context.Context, sessionID string, fileID string) error {
	batch, err := f.Batch(ctx)
	if err != nil {
		return fmt.Errorf("ds batch error: %w", err)
	}

	if err := batch.Delete(ctx, fileDsKey.SessionFileKey(sessionID, fileID)); err != nil {
		return fmt.Errorf("batch.Delete error: %w", err)
	}

	if err := batch.Delete(ctx, fileDsKey.FileSessionKey(fileID, sessionID)); err != nil {
		return fmt.Errorf("batch.Delete error: %w", err)
	}

	if err := batch.Delete(ctx, fileDsKey.SessionFileTimeKey(sessionID, fileID)); err != nil {
		return fmt.Errorf("batch.Delete error: %w", err)
	}

	return batch.Commit(ctx)
}

func (f *FileDataStore) GetSessionFileIDs(ctx context.Context, sessionID string) ([]string, error) {
	prefix := fileDsKey.SessionFilePrefix(sessionID)
	results, err := f.Query(ctx, query.Query{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var fileIDs []string
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("results.Next error: %w", result.Error)
		}

		fileIDs = append(fileIDs, strings.TrimPrefix(result.Key, prefix))
	}

	return fileIDs, nil
}

func (f *FileDataStore) GetSessionFiles(ctx context.Context, sessionID string, keywords string, offset int, limit int) ([]*pb.FileInfo, error) {
	prefix := fileDsKey.SessionFileTimePrefix(sessionID)
	results, err := f.Query(ctx, query.Query{
		Prefix: fileDsKey.SessionFileTimePrefix(sessionID),
		Orders: []query.Order{query.OrderByValueDescending{}},
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var files []*pb.FileInfo
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("results.Next error: %w", result.Error)
		}

		fileID := strings.TrimPrefix(result.Key, prefix)
		bs, err := f.Get(ctx, fileDsKey.SessionFileKey(sessionID, fileID))
		if err != nil {
			return nil, fmt.Errorf("ds get session file error: %w", err)
		}

		var file pb.FileInfo
		if err := proto.Unmarshal(bs, &file); err != nil {
			return nil, fmt.Errorf("proto unmarshal error: %w", err)
		}

		files = append(files, &file)
	}

	return files, nil
}
