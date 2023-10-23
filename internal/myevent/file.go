package myevent

import (
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

type GetResourceDataResult struct {
	Error error
	Data  []byte
}

// EvSyncResource 同步资源文件
type EvtSyncResource struct {
	ResourceID string
	PeerIDs    []peer.ID
}

// EvtLogSessionAttachment 记录会话关联的资源与文件
type EvtLogSessionAttachment struct {
	SessionID  string
	ResourceID string
	File       *mytype.FileInfo
	Result     chan<- error
}

// EvtClearSessionResources 删除会话关联的资源（删除会话消息时用）
type EvtClearSessionResources struct {
	SessionID string
	Result    chan<- error
}

// EvtClearSessionFiles 删除会话关联的文件（删除会话时用）
type EvtClearSessionFiles struct {
	SessionID string
	Result    chan<- error
}

// EvtGetResourceData 获取资源文件数据
type EvtGetResourceData struct {
	ResourceID string
	Result     chan<- *GetResourceDataResult
}

// EvtSaveResourceData 保存资源文件数据
type EvtSaveResourceData struct {
	SessionID  string
	ResourceID string
	Data       []byte
	Result     chan<- error
}
