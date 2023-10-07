package myevent

import "github.com/libp2p/go-libp2p/core/peer"

type EvtDownloadFile struct {
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
	FileID       string
	FileSize     int64
	DownloadSize int64
}

type EvtUploadResource struct {
	ToPeerID peer.ID
	GroupID  string
	FileID   string
	Result   chan<- error
}

type EvtDownloadResource struct {
	PeerID  peer.ID
	GroupID string
	FileID  string
	Result  chan<- error
}
