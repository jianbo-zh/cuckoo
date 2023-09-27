package myevent

import "github.com/libp2p/go-libp2p/core/peer"

type EvtDownloadRequest struct {
	FromPeerIDs []peer.ID
	FileName    string
	FileSize    int64
	HashAlgo    string
	HashValue   string
}

type EvtDownloadResult struct {
	FileName   string
	FileSize   int64
	HashAlgo   string
	HashValue  string
	FilePath   string
	IsSuccess  bool
	FailReason string
}

type EvtDownloadProcess struct {
	FileName     string
	FileHash     string
	FileSize     int64
	DownloadSize int64
}
