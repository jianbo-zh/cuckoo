package filesvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

type FileServiceIface interface {
	CopyFileToResource(ctx context.Context, srcFile string) (resourceID string, err error)
	CopyFileToFile(ctx context.Context, srcFile string) (fileID string, err error)

	DownloadSessionFile(ctx context.Context, sessionID string, peerIDs []peer.ID, file *mytype.FileInfo) error
	GetSessionFiles(ctx context.Context, sessionID string, keywords string, offset int, limit int) ([]mytype.FileInfo, error)
	DeleteSessionFile(ctx context.Context, sessionID string, fileIDs []string) error

	DownloadAvatar(ctx context.Context, peerID peer.ID, avatar string) error
	Close()
}
