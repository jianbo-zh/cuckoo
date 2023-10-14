package fileproto

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	myevent "github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/filepb"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-msgio/pbio"
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
			case myevent.EvtCheckAvatar:
				go f.handleCheckAvatarEvent(ctx, evt)

			case myevent.EvtSendResource:
				go f.handleUploadResourceEvent(ctx, evt)

			case myevent.EvtDownloadResource:
				go f.handleDownloadResourceEvent(ctx, evt)

			case myevent.EvtLogSessionAttachment:
				go f.handleLogSessionAttachmentEvent(ctx, evt)

			case myevent.EvtGetResourceData:
				go f.handleGetResourceDataEvent(ctx, evt)

			case myevent.EvtSaveResourceData:
				go f.handleSaveResourceDataEvent(ctx, evt)
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

func (f *FileProto) handleCheckAvatarEvent(ctx context.Context, evt myevent.EvtCheckAvatar) {
	fmt.Printf("handle check avatar event %v", evt)

	if evt.Avatar != "" && len(evt.PeerIDs) > 0 {
		_, err := os.Stat(filepath.Join(f.conf.ResourceDir, evt.Avatar))
		if err != nil && os.IsNotExist(err) {
			fmt.Println("for download resource")
			for _, peerID := range evt.PeerIDs {
				fmt.Println("download resource", peerID.String(), evt.Avatar)
				if err = f.DownloadResource(ctx, peerID, evt.Avatar); err == nil {
					break
				}
			}
		} else {
			fmt.Println("avatar is exists")
		}
	}
}

// handleUploadResourceEvent 处理上传资源事件
func (f *FileProto) handleUploadResourceEvent(ctx context.Context, evt myevent.EvtSendResource) {

	var resultErr error
	defer func() {
		if r := recover(); r != nil {
			evt.Result <- fmt.Errorf("panic: %v", r)
		} else {
			evt.Result <- resultErr
		}
		close(evt.Result)
	}()

	fmt.Println("handle send file event: ", evt)

	filePath := path.Join(f.conf.ResourceDir, evt.FileID)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		resultErr = fmt.Errorf("os.Stat error: %w", err)
		return

	} else if fileInfo.Size() <= 0 {
		resultErr = fmt.Errorf("os.Stat file size 0")
		return
	}

	fmt.Println("file exists")

	stream, err := f.host.NewStream(network.WithDialPeerTimeout(ctx, time.Second), evt.ToPeerID, RESOURCE_UPLOAD_ID)
	if err != nil {
		resultErr = fmt.Errorf("host.NewStream error: %w", err)
		return
	}
	defer stream.Close()

	fmt.Println("stream open")

	osfile, err := os.Open(filePath)
	if err != nil {
		resultErr = fmt.Errorf("os.Open file error: %w", err)
		stream.Reset()
		return
	}
	defer osfile.Close()

	fmt.Println("file open")

	// 发送上传请求
	wt := pbio.NewDelimitedWriter(stream)
	reqMsg := pb.FileUploadRequest{
		GroupId:  evt.GroupID,
		FileId:   evt.FileID,
		FileSize: fileInfo.Size(),
	}
	if err = wt.WriteMsg(&reqMsg); err != nil {
		resultErr = fmt.Errorf("pbio.WriteMsg error: %w", err)
		stream.Reset()
		return
	}

	fmt.Println("send upload request: ", reqMsg.String())

	// 接收请求回复
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	var replyMsg pb.FileUploadReply
	if err = rd.ReadMsg(&replyMsg); err != nil {
		resultErr = fmt.Errorf("pbio.ReadMsg error: %w", err)
		stream.Reset()
		return
	}

	fmt.Println("get upload reply: ", replyMsg.String())

	if replyMsg.Exists { // 已存在，则不用上传
		resultErr = nil
		return
	}

	fmt.Println("start send file")

	// 开始发送文件
	bufStream := bufio.NewWriter(stream)
	size, err := bufStream.ReadFrom(osfile)
	if err != nil {
		resultErr = fmt.Errorf("bufStream.ReadFrom error: %w", err)
		stream.Reset()
		return
	}
	bufStream.Flush()
	stream.CloseWrite() // 结束发送文件

	log.Debugf("send file finish %s, size: %d", evt.FileID, size)

	// 检查发送结构
	var resultMsg pb.FileUploadResult
	if err = rd.ReadMsg(&resultMsg); err != nil {
		resultErr = fmt.Errorf("pbio.ReadMsg error: %w", err)
		stream.Reset()
		return
	}

	fmt.Println("receive result msg", resultMsg.String())

	if resultMsg.FileId == evt.FileID && resultMsg.FileSize == fileInfo.Size() && resultMsg.ErrMsg == "" {
		resultErr = nil
		log.Debugf("send file success %s", evt.FileID)

	} else {
		err = fmt.Errorf("send failed, fileId: %s, fileSize: %d, errmsg: %s", resultMsg.FileId, resultMsg.FileSize, resultMsg.ErrMsg)
		resultErr = err
		log.Debugf("send file failed : %s", err.Error())
	}
}

// handleDownloadResourceEvent 处理下载资源事件
func (f *FileProto) handleDownloadResourceEvent(ctx context.Context, evt myevent.EvtDownloadResource) {
	evt.Result <- f.DownloadResource(ctx, evt.PeerID, evt.FileID)
	close(evt.Result)
}
