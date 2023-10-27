package fileds

import (
	"context"

	pb "github.com/jianbo-zh/dchat/service/filesvc/protobuf/pb/filepb"
)

type FileIface interface {
	GetFileSessionIDs(ctx context.Context, fileID string) ([]string, error)

	SaveSessionResource(ctx context.Context, sessionID string, resourceID string) error
	GetSessionResourceIDs(ctx context.Context, sessionID string) ([]string, error)
	RemoveSessionResource(ctx context.Context, sessionID string, resourceID string) error

	SaveSessionFile(ctx context.Context, sessionID string, uploadFile *pb.FileInfo) error
	GetSessionFileIDs(ctx context.Context, sessionID string) ([]string, error)
	GetSessionFiles(ctx context.Context, sessionID string, keywords string, offset int, limit int) ([]*pb.FileInfo, error)
	RemoveSessionFile(ctx context.Context, sessionID string, fileID string) error
}
