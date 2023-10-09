package filesvc

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

type FileServiceIface interface {
	ConvertFileToResource(ctx context.Context, srcFile string) (string, error)
	// CalcFileID(ctx context.Context, filePath string) (*mytype.FileID, error)
	DownloadAvatar(ctx context.Context, peerID peer.ID, avatar string) error
	SendPeerFile(ctx context.Context, peerID peer.ID, file string) error
	Close()
}
