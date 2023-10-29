package fileproto

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/service/filesvc/protobuf/pb/filepb"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-msgio/pbio"
	"google.golang.org/protobuf/proto"
)

// 文件查询处理
func (f *FileProto) fileQueryHandler(stream network.Stream) {
	defer stream.Close()

	var msg pb.FileQuery
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("pbio read file query msg: %w", err)
		return
	}

	exists := false
	_, err := os.Stat(filepath.Join(f.conf.FileDir, msg.FileId))
	if err != nil {
		if os.IsNotExist(err) {
			exists = false

		} else {
			stream.Reset()
			log.Errorf("os.Stat error: %w", err)
			return
		}
	} else {
		exists = true
	}

	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&pb.FileQueryResult{FileId: msg.FileId, Exists: exists}); err != nil {
		log.Errorf("pbio write file query result msg error: %w", err)
		return
	}
}

// resourceUploadIDHandler 资源上传处理器
func (f *FileProto) resourceUploadIDHandler(stream network.Stream) {
	remotePeerID := stream.Conn().RemotePeer()

	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	wt := pbio.NewDelimitedWriter(stream)

	var reqMsg pb.FileUploadRequest
	if err := rd.ReadMsg(&reqMsg); err != nil {
		log.Errorf("pbio.ReadMsg error: %w", err)
		stream.Reset()
		return
	}

	// 判断文件是否已存在
	filePath := path.Join(f.conf.ResourceDir, reqMsg.FileId)
	if _, err := os.Stat(filePath); err == nil {
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

	// 重命名文件
	if err := os.Rename(tmpFilePath, filePath); err != nil {
		log.Errorf("os.Rename error: %w", err)
		stream.Reset()
		return
	}

	// 记录关联资源
	sessionID := mytype.ContactSessionID(remotePeerID)
	if err := f.data.SaveSessionResource(context.Background(), sessionID.String(), reqMsg.FileId); err != nil {
		log.Errorf("data.SaveSessionResource error: %w", err)
		stream.Reset()
		return
	}

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
	log.Infoln("download resource handler")

	defer stream.Close()

	// 接收请求
	var reqMsg pb.DownloadResourceRequest
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
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
	defer stream.Close()

	for {
		// 读取请求分片
		var msg pb.FileDownloadChunk
		rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
		if err := rd.ReadMsg(&msg); err != nil {
			log.Errorf("pbio read download chunk error: %v", err)
			return
		}

		// 检查文件是否存在
		file := filepath.Join(f.conf.FileDir, msg.FileId)
		fileInfo, err := os.Stat(file)
		if err != nil {
			stream.Reset()
			log.Errorf("os.Stat error: %v", err)
			return
		}

		if len(msg.FileId) == 0 || msg.ChunkIndex < 0 || msg.ChunkTotal < 0 || msg.ChunkIndex >= msg.ChunkTotal {
			stream.Reset()
			log.Errorf("param error chunkIndex egt chunkTotal")
			return
		}

		// 计算分片偏移量
		chunkIndex := msg.ChunkIndex
		chunkTotal := msg.ChunkTotal
		fileSize := fileInfo.Size()

		chunkSize := fileSize / chunkTotal
		offset := chunkIndex * chunkSize
		limit := chunkSize
		if chunkIndex+1 == chunkTotal { // 最后一段
			limit = fileSize - offset
		}

		ofile, err := os.Open(file)
		if err != nil {
			stream.Reset()
			log.Errorf("os.Open error: %v", err)
			return
		}
		defer ofile.Close()

		// 计算分片hash
		ofile.Seek(offset, io.SeekStart)
		hashSum := md5.New()
		if size, err := io.CopyN(hashSum, ofile, limit); err != nil {
			stream.Reset()
			log.Errorf("io.CopyN error: %v", err)
			return

		} else if size != limit {
			stream.Reset()
			log.Errorf("io.CopyN size error")
			return
		}

		// 发送分片信息
		chunkHeader := pb.FileChunkInfo{
			FileId:    msg.FileId,
			ChunkHash: fmt.Sprintf("%x", hashSum.Sum(nil)),
			ChunkSize: limit,
		}
		bs, err := proto.Marshal(&chunkHeader)
		if err != nil {
			stream.Reset()
			log.Errorf("proto.Marshal error: %v", err)
			return
		}

		// base64 编码，因为pb本身有换行
		bs64 := []byte(base64.StdEncoding.EncodeToString(bs))
		bufWriter := bufio.NewWriter(stream)
		if size, err := bufWriter.Write(bs64); err != nil {
			log.Errorf("bufWriter write error: %v", err)
			return

		} else if size != len(bs64) {
			stream.Reset()
			log.Errorf("bufWriter write size error")
			return
		}

		// 发送头信息分隔符
		if err := bufWriter.WriteByte('\n'); err != nil {
			log.Errorf("bufWriter write byte error")
			return
		}

		// 发送具体分片
		ofile.Seek(offset, io.SeekStart)
		if size, err := io.CopyN(bufWriter, ofile, limit); err != nil {
			log.Errorf("io.CopyN file error: %v", err)
			return

		} else if size != limit {
			stream.Reset()
			log.Errorf("io.CopyN size error")
			return
		}
		bufWriter.Flush()

		log.Debugln("send file chunk finish")
	}

}
