package myevent

import (
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

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

type EvtSendResource struct {
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

type EvtRecordSessionAttachment struct {
	SessionID  string
	ResourceID string
	File       *mytype.FileInfo
	Result     chan<- error
}
