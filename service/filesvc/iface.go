package filesvc

import (
	"context"

	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

type FileServiceIface interface {
	CalcFileHash(ctx context.Context, filePath string) (*mytype.FileHash, error)
	AvatarDownload(ctx context.Context, peerID peer.ID, avatar string) error
	Close()
}
