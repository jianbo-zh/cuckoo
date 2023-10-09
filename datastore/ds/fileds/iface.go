package fileds

import (
	"context"

	pb "github.com/jianbo-zh/dchat/protobuf/pb/filepb"
)

type FileIface interface {
	GetFileSessionIDs(ctx context.Context, fileID string) ([]string, error)

	SaveSessionResource(ctx context.Context, sessionID string, resourceID string) error
	GetSessionResourceIDs(ctx context.Context, sessionID string) ([]string, error)

	SaveSessionUploadFile(ctx context.Context, sessionID string, uploadFile *pb.FileInfo) error
	GetSessionUploadFiles(ctx context.Context, sessionID string) ([]*pb.FileInfo, error)

	SaveSessionDownloadFile(ctx context.Context, sessionID string, downloadFile *pb.FileInfo) error
	GetSessionDownloadFiles(ctx context.Context, sessionID string) ([]*pb.FileInfo, error)
}
