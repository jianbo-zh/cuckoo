package fileproto

import (
	"bufio"
	"fmt"
	"os"
	"path"

	pb "github.com/jianbo-zh/dchat/protobuf/pb/filepb"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-msgio/pbio"
)

// 文件查询处理
func (f *FileProto) fileQueryHandler(stream network.Stream) {

}

// resourceUploadIDHandler 资源上传处理器
func (f *FileProto) resourceUploadIDHandler(stream network.Stream) {
	fmt.Println("resourceUploadIDHandler start")

	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	wt := pbio.NewDelimitedWriter(stream)

	var reqMsg pb.FileUploadRequest
	if err := rd.ReadMsg(&reqMsg); err != nil {
		log.Errorf("pbio.ReadMsg error: %w", err)
		stream.Reset()
		return
	}

	fmt.Println("receive upload request: ", reqMsg.String())

	// 判断文件是否已存在
	filePath := path.Join(f.conf.ResourceDir, reqMsg.FileId)
	if _, err := os.Stat(filePath); err == nil {
		fmt.Println("file exists")
		if err := wt.WriteMsg(&pb.FileUploadReply{Exists: true}); err != nil {
			log.Errorf("pbio.WriteMsg error: %w", err)
			stream.Reset()
		}
		return

	} else if !os.IsNotExist(err) {
		log.Errorf("os.Stat error: %w", err)
		stream.Reset()
		return
	}

	// 告知不存在，可以开始上传
	if err := wt.WriteMsg(&pb.FileUploadReply{Exists: false}); err != nil {
		log.Errorf("pbio.WriteMsg error: %w", err)
		stream.Reset()
		return
	}

	// 接收临时文件
	tmpFilePath := path.Join(f.conf.ResourceDir, fmt.Sprintf("%s_%s", "tmp", reqMsg.FileId))
	tmpFile, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		log.Errorf("os.OpenFile error: %w", err)
		stream.Reset()
		return
	}

	fmt.Println("start receive file")

	bfrd := bufio.NewReader(stream)
	size, err := bfrd.WriteTo(tmpFile)
	if err != nil {
		log.Errorf("bufWriter.ReadFrom error: %w", err)
		tmpFile.Close()
		stream.Reset()
		return
	}
	tmpFile.Close()

	log.Debugln("receive file finish %s, fileSize: %d", reqMsg.FileId, size)
	fmt.Println("receive file finish")

	// 重命名文件
	if err := os.Rename(tmpFilePath, filePath); err != nil {
		log.Errorf("os.Rename error: %w", err)
		stream.Reset()
		return
	}
	fmt.Println("rename file finish")

	// 回复接收成功
	if err := wt.WriteMsg(&pb.FileUploadResult{
		FileId:   reqMsg.FileId,
		FileSize: reqMsg.FileSize,
		ErrMsg:   "",
	}); err != nil {
		log.Errorf("pbio.WriteMsg error: %w", err)
		stream.Reset()
		return
	}

	log.Debugln("receive file success ", reqMsg.FileId)
}

// resourceDownloadIDHandler 资源下载处理器
func (f *FileProto) resourceDownloadIDHandler(stream network.Stream) {
	log.Debugln("download resource start")

	defer stream.Close()

	// 接收请求
	var reqMsg pb.DownloadResourceRequest
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	if err := rd.ReadMsg(&reqMsg); err != nil {
		log.Errorf("pbio read request msg error: %w", err)
		stream.Reset()
		return
	}

	filePath := path.Join(f.conf.ResourceDir, reqMsg.FileId)
	if _, err := os.Stat(filePath); err != nil {
		log.Errorf("os.Stat %s error: %w", filePath, err)
		stream.Reset()
		return
	}

	// // todo: 鉴权
	// wt := pbio.NewDelimitedWriter(stream)
	// if err := wt.WriteMsg(&pb.DownloadResourceReply{Error: ""}); err != nil {
	// 	log.Errorf("pbio write reply msg error: %w", err)
	// 	stream.Reset()
	// 	return
	// }

	// 开始发送文件
	osfile, err := os.Open(filePath)
	if err != nil {
		log.Errorf("os.Open %s error: %w", filePath, err)
		stream.Reset()
		return
	}
	defer osfile.Close()

	bufStream := bufio.NewWriter(stream)
	size, err := bufStream.ReadFrom(osfile)
	if err != nil {
		log.Errorf("bufStream.ReadFrom error: %w", err)
		stream.Reset()
		return
	}
	bufStream.Flush()

	log.Debugf("send file: %s success, sendSize: %d", reqMsg.FileId, size)
}

// 文件下载处理
func (f *FileProto) fileDownloadHandler(stream network.Stream) {

}
