package fileproto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	myevent "github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/libp2p/go-libp2p/core/event"
)

func (f *FileProto) subscribeHandler(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}

			switch evt := e.(type) {
			case myevent.EvSyncResource:
				go f.handleCheckAvatarEvent(ctx, evt)

			case myevent.EvtLogSessionAttachment:
				go f.handleLogSessionAttachmentEvent(ctx, evt)

			case myevent.EvtGetResourceData:
				go f.handleGetResourceDataEvent(ctx, evt)

			case myevent.EvtSaveResourceData:
				go f.handleSaveResourceDataEvent(ctx, evt)

			case myevent.EvtClearSessionResources:
				go f.handleClearSessionResourcesEvent(ctx, evt)

			case myevent.EvtClearSessionFiles:
				go f.handleClearSessionFilesEvent(ctx, evt)
			}

		case <-ctx.Done():
			return
		}
	}
}

// handleGetResourceDataEvent 读取资源数据
func (f *FileProto) handleGetResourceDataEvent(ctx context.Context, evt myevent.EvtGetResourceData) {
	var result myevent.GetResourceDataResult
	defer func() {
		evt.Result <- &result
		close(evt.Result)
	}()

	file := filepath.Join(f.conf.ResourceDir, evt.ResourceID)
	if _, err := os.Stat(file); err != nil {
		result.Error = fmt.Errorf("os.Stat error: %w", err)
		return
	}

	bs, err := os.ReadFile(file)
	if err != nil {
		result.Error = fmt.Errorf("os.ReadFile error: %w", err)
		return
	}

	result.Data = bs
}

// handleClearSessionResourcesEvent 删除会话资源
func (f *FileProto) handleClearSessionResourcesEvent(ctx context.Context, evt myevent.EvtClearSessionResources) {
	var resultErr error

	defer func() {
		evt.Result <- resultErr
		close(evt.Result)
	}()

	// 查找资源
	resourceIDs, err := f.data.GetSessionResourceIDs(ctx, evt.SessionID)
	if err != nil {
		resultErr = fmt.Errorf("data.GetSessionResourceIDs error: %w", err)
		return
	}
	if len(resourceIDs) > 0 {
		if err := f.DeleteSessionResources(ctx, evt.SessionID, resourceIDs); err != nil {
			resultErr = fmt.Errorf("DeleteSessionResources error: %w", err)
			return
		}
	}
}

// handleClearSessionFilesEvent 删除会话文件
func (f *FileProto) handleClearSessionFilesEvent(ctx context.Context, evt myevent.EvtClearSessionFiles) {
	var resultErr error

	defer func() {
		evt.Result <- resultErr
		close(evt.Result)
	}()

	// 查找文件
	fileIDs, err := f.data.GetSessionFileIDs(ctx, evt.SessionID)
	if err != nil {
		resultErr = fmt.Errorf("data.GetSessionFileIDs error: %w", err)
		return
	}

	// 删除文件
	if len(fileIDs) > 0 {
		if err := f.DeleteSessionFiles(ctx, evt.SessionID, fileIDs); err != nil {
			resultErr = fmt.Errorf("DeleteSessionFiles error: %w", err)
			return
		}
	}
}

// handleSaveResourceDataEvent 保存资源
func (f *FileProto) handleSaveResourceDataEvent(ctx context.Context, evt myevent.EvtSaveResourceData) {
	var resultErr error

	defer func() {
		evt.Result <- resultErr
		close(evt.Result)
	}()

	file := filepath.Join(f.conf.ResourceDir, evt.ResourceID)
	if _, err := os.Stat(file); err == nil {
		// exists
		return

	} else if !os.IsNotExist(err) {
		resultErr = fmt.Errorf("os.Stat error: %w", err)
		return
	}

	if err := os.WriteFile(file, evt.Data, 0755); err != nil {
		resultErr = fmt.Errorf("os.WriteFile error: %w", err)
		return
	}

	if err := f.data.SaveSessionResource(ctx, evt.SessionID, evt.ResourceID); err != nil {
		resultErr = fmt.Errorf("save session resource error: %w", err)
		return
	}
}

func (f *FileProto) handleLogSessionAttachmentEvent(ctx context.Context, evt myevent.EvtLogSessionAttachment) {
	var resultErr error

	defer func() {
		evt.Result <- resultErr
		close(evt.Result)
	}()

	if evt.ResourceID != "" {
		if err := f.data.SaveSessionResource(ctx, evt.SessionID, evt.ResourceID); err != nil {
			resultErr = fmt.Errorf("save session resource error: %w", err)
			return
		}
	}

	if evt.File != nil {
		if err := f.data.SaveSessionFile(ctx, evt.SessionID, encodeFile(evt.File)); err != nil {
			resultErr = fmt.Errorf("data.SaveSessionFile error: %w", err)
			return
		}
	}
}

func (f *FileProto) handleCheckAvatarEvent(ctx context.Context, evt myevent.EvSyncResource) {
	fmt.Printf("handle check avatar event %v", evt)

	if evt.ResourceID != "" && len(evt.PeerIDs) > 0 {
		_, err := os.Stat(filepath.Join(f.conf.ResourceDir, evt.ResourceID))
		if err != nil && os.IsNotExist(err) {
			fmt.Println("for download resource")
			for _, peerID := range evt.PeerIDs {
				fmt.Println("download resource", peerID.String(), evt.ResourceID)
				if err = f.DownloadResource(ctx, peerID, evt.ResourceID); err == nil {
					break
				}
			}
		} else {
			fmt.Println("avatar is exists")
		}
	}
}
