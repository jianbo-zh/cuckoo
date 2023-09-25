package event

import "github.com/libp2p/go-libp2p/core/peer"

type DownloadRequestEvt struct {
	FromPeerIDs []peer.ID
	FileName    string
	FileSize    int64
	HashAlgo    string
	HashValue   string
}

type DownloadResultEvt struct {
	FileName   string
	FileSize   int64
	HashAlgo   string
	HashValue  string
	FilePath   string
	IsSuccess  bool
	FailReason string
}

type DownloadProcessEvt struct {
	FileName     string
	FileHash     string
	FileSize     int64
	DownloadSize int64
}
