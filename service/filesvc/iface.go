package filesvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

type FileServiceIface interface {
	ConvertFileToResource(ctx context.Context, srcFile string) (string, error)
	DownloadContactFile(ctx context.Context, peerID peer.ID, file *mytype.FileInfo) error
	DownloadGroupFile(ctx context.Context, groupID string, peerIDs []peer.ID, file *mytype.FileInfo) error

	GetContactFiles(ctx context.Context, contactID peer.ID, keywords string, offset int, limit int) ([]mytype.FileInfo, error)
	GetGroupFiles(ctx context.Context, groupID string, keywords string, offset int, limit int) ([]mytype.FileInfo, error)

	DeleteContactFile(ctx context.Context, contactID peer.ID, fileIDs []string) error

	// CalcFileID(ctx context.Context, filePath string) (*mytype.FileID, error)
	DownloadAvatar(ctx context.Context, peerID peer.ID, avatar string) error
	Close()
}
