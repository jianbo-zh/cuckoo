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
	"strings"
	"sync"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	myevent "github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/myprotocol"
	"github.com/jianbo-zh/dchat/internal/mytype"
	ds "github.com/jianbo-zh/dchat/service/filesvc/datastore/ds/fileds"
	pb "github.com/jianbo-zh/dchat/service/filesvc/protobuf/pb/filepb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
	"google.golang.org/protobuf/proto"
)

var log = logging.Logger("fileproto")

var StreamTimeout = 1 * time.Minute

const (
	QUERY_ID    = myprotocol.FileQueryID_v100
	DOWNLOAD_ID = myprotocol.FileDownloadID_v100

	RESOURCE_UPLOAD_ID   = myprotocol.ResourceUploadID_v100
	RESOURCE_DOWNLOAD_ID = myprotocol.ResourceDownloadID_v100

	ServiceName = "peer.message"

	ChunkSize = 50 * 1024 // 256K
)

type FileProto struct {
	host myhost.Host
	data ds.FileIface

	conf config.FileServiceConfig
}

func NewFileProto(conf config.FileServiceConfig, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus) (*FileProto, error) {
	var err error

	file := FileProto{
		host: lhost,
		data: ds.FileWrap(ids),
		conf: conf,
	}

	lhost.SetStreamHandler(QUERY_ID, file.fileQueryHandler)
	lhost.SetStreamHandler(RESOURCE_UPLOAD_ID, file.resourceUploadIDHandler)
	lhost.SetStreamHandler(RESOURCE_DOWNLOAD_ID, file.resourceDownloadIDHandler)
	lhost.SetStreamHandler(DOWNLOAD_ID, file.fileDownloadHandler)

	// 订阅器
	sub, err := ebus.Subscribe([]any{
		new(myevent.EvtLogSessionAttachment), new(myevent.EvtSyncResource),
		new(myevent.EvtGetResourceData), new(myevent.EvtSaveResourceData),
		new(myevent.EvtClearSessionResources), new(myevent.EvtClearSessionFiles),
	})
	if err != nil {
		return nil, fmt.Errorf("ebus subscribe error: %w", err)
	}

	go file.subscribeHandler(context.Background(), sub)

	return &file, nil
}

// func (f *FileProto)

func (f *FileProto) GetSessionFiles(ctx context.Context, sessionID string, keywords string, offset int, limit int) ([]*pb.FileInfo, error) {
	return f.data.GetSessionFiles(ctx, sessionID, keywords, offset, limit)
}

// DeleteSessionResources 删除会话的资源文件
func (f *FileProto) DeleteSessionResources(ctx context.Context, sessionID string, resourceIDs []string) error {
	for _, resourceID := range resourceIDs {
		if err := f.data.RemoveSessionResource(ctx, sessionID, resourceID); err != nil {
			return fmt.Errorf("data.RemoveSessionFile error: %w", err)
		}

		if sessionIDs, err := f.data.GetFileSessionIDs(ctx, resourceID); err != nil {
			return fmt.Errorf("data.GetFileSessionIDs error: %w", err)

		} else if len(sessionIDs) == 0 { // 没有其他会话引用了，则可以删除
			if err := os.Remove(filepath.Join(f.conf.ResourceDir, resourceID)); err != nil {
				return fmt.Errorf("os.Remove error: %w", err)
			}
		}
	}

	return nil
}

// DeleteSessionFiles 删除会话的文件
func (f *FileProto) DeleteSessionFiles(ctx context.Context, sessionID string, fileIDs []string) error {
	for _, fileID := range fileIDs {
		if err := f.data.RemoveSessionFile(ctx, sessionID, fileID); err != nil {
			return fmt.Errorf("data.RemoveSessionFile error: %w", err)
		}

		if sessionIDs, err := f.data.GetFileSessionIDs(ctx, fileID); err != nil {
			return fmt.Errorf("data.GetFileSessionIDs error: %w", err)

		} else if len(sessionIDs) == 0 { // 没有其他会话引用了，则可以删除
			if err := os.Remove(filepath.Join(f.conf.FileDir, fileID)); err != nil {
				return fmt.Errorf("os.Remove error: %w", err)
			}
		}
	}

	return nil
}

// downloadResource  下载资源
func (f *FileProto) DownloadResource(ctx context.Context, peerID peer.ID, fileID string) error {

	// 检查本地有没有
	filePath := path.Join(f.conf.ResourceDir, fileID)
	if _, err := os.Stat(filePath); err == nil { // exists
		return nil

	} else if !os.IsNotExist(err) {
		return fmt.Errorf("os.Stat error: %w", err)
	}

	// 发送下载请求
	stream, err := f.host.NewStream(network.WithUseTransient(ctx, ""), peerID, RESOURCE_DOWNLOAD_ID)
	if err != nil {
		return fmt.Errorf("host.NewStream error: %w", err)
	}
	defer stream.Close()

	reqMsg := pb.DownloadResourceRequest{FileId: fileID}
	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&reqMsg); err != nil {
		stream.Reset()
		return fmt.Errorf("pbio.WriteMsg error: %w", err)
	}

	// 开始接收文件
	tmpFilePath := path.Join(f.conf.ResourceDir, fmt.Sprintf("%s_%s", "tmp", fileID))
	tmpFile, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		stream.Reset()
		return fmt.Errorf("os.OpenFile error: %w", err)
	}

	bufStream := bufio.NewReader(stream)
	_, err = bufStream.WriteTo(tmpFile)
	if err != nil {
		tmpFile.Close()
		stream.Reset()
		return fmt.Errorf("bufStream.WriteTo tmp file error: %w", err)
	}
	tmpFile.Close()

	// 重命名文件
	if err := os.Rename(tmpFilePath, filePath); err != nil {
		stream.Reset()
		return fmt.Errorf("os.Rename error: %w", err)
	}

	return nil
}

func (f *FileProto) DownloadFile(ctx context.Context, sessionID string, fromPeerIDs []peer.ID, file *mytype.FileInfo) error {
	// 检查是否存在，如果存在则不下载
	if _, err := os.Stat(filepath.Join(f.conf.FileDir, file.FileID)); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("os.Stat error: %w", err)
	}

	// 执行下载
	if err := f.downloadFile(ctx, fromPeerIDs, file.FileID, file.FileSize); err != nil {
		return fmt.Errorf("download file error: %w", err)
	}

	// 下载成功则记录
	if err := f.data.SaveSessionFile(ctx, sessionID, encodeFile(file)); err != nil {
		return fmt.Errorf("save session file error: %w", err)
	}

	return nil
}

func (d *FileProto) downloadFile(ctx context.Context, fromPeerIDs []peer.ID, fileID string, fileSize int64) error {

	hostID := d.host.ID()

	// 判断文件是否存在，存在则返回
	filePath := filepath.Join(d.conf.FileDir, fileID)
	if _, err := os.Stat(filePath); err == nil {
		log.Errorln("file exists")
		return nil

	} else if !os.IsNotExist(err) {
		return fmt.Errorf("os.Stat error: %v", err)
	}
	// 查询哪些peer有文件
	resultCh := make(chan peer.ID, len(fromPeerIDs))
	var wg sync.WaitGroup
	for _, peerID := range fromPeerIDs {
		// 排除自己
		if peerID == hostID {
			continue
		}

		wg.Add(1)
		go d.queryFile(ctx, &wg, resultCh, peerID, fileID)
	}
	wg.Wait()
	close(resultCh)

	var peerIDs []peer.ID
	for peerID := range resultCh {
		peerIDs = append(peerIDs, peerID)
	}

	if len(peerIDs) == 0 {
		return fmt.Errorf("no peer have file")
	}

	// 计算文件任务
	chunkTotal := fileSize / ChunkSize
	if (fileSize % ChunkSize) > 0 {
		chunkTotal++
	}
	if chunkTotal <= 0 {
		chunkTotal = 1
	}

	taskCh := make(chan int64, chunkTotal)
	for i := int64(0); i < chunkTotal; i++ {
		taskCh <- i
	}

	mhr := NewMultiHandleResult(chunkTotal)

	// 启动分片下载
	var wg2 sync.WaitGroup
	for _, peerID := range peerIDs {
		wg2.Add(1)
		go d.downloadFileChunk(ctx, &wg2, mhr, taskCh, peerID, fileID, chunkTotal)
	}
	wg2.Wait()
	close(taskCh)

	// 检查任务完成没
	if _, isOk := <-taskCh; isOk {
		// 还有值，则下载失败
		return fmt.Errorf("download failed, task not finish")
	}

	// 分片下载完成，则合并整个文件
	ofile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("os open file error: %w", err)
	}
	defer ofile.Close()

	hashSum := md5.New()
	multiWriter := io.MultiWriter(ofile, hashSum)
	for chunkIndex := int64(0); chunkIndex < chunkTotal; chunkIndex++ {
		ofile2, err := os.Open(filepath.Join(d.conf.TmpDir, chunkFileName(fileID, chunkIndex)))
		if err != nil {
			os.Remove(filePath)
			return fmt.Errorf("os open error: %w", err)
		}

		if _, err = io.Copy(multiWriter, ofile2); err != nil {
			ofile2.Close()
			os.Remove(filePath)
			return fmt.Errorf("io.Copy error: %w", err)
		}
		ofile2.Close()
	}

	calcHash := fmt.Sprintf("%x", hashSum.Sum(nil))
	if !strings.Contains(fileID, calcHash) {
		os.Remove(filePath)
		return fmt.Errorf("file %s hash %s error", fileID, calcHash)
	}

	// 下载完成后，删除分片文件
	for chunkIndex := int64(0); chunkIndex < chunkTotal; chunkIndex++ {
		os.Remove(filepath.Join(d.conf.TmpDir, chunkFileName(fileID, chunkIndex)))
	}

	return nil
}

func (d *FileProto) queryFile(ctx context.Context, wg *sync.WaitGroup, resultCh chan peer.ID, peerID peer.ID, fileID string) {
	defer wg.Done()

	stream, err := d.host.NewStream(network.WithUseTransient(ctx, ""), peerID, QUERY_ID)
	if err != nil {
		log.Errorf("host new stream error: %v", err)
		return
	}
	defer stream.Close()

	wt := pbio.NewDelimitedWriter(stream)
	if err = wt.WriteMsg(&pb.FileQuery{FileId: fileID}); err != nil {
		log.Errorf("pbio write msg error: %v", err)
		return
	}

	var result pb.FileQueryResult
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	if err = rd.ReadMsg(&result); err != nil {
		log.Errorf("pbio read msg error: %v", err)
		return
	}

	if result.Exists {
		resultCh <- peerID
	}
}

// downloadPart 下载分片
func (d *FileProto) downloadFileChunk(ctx context.Context, wg *sync.WaitGroup, mhr *MultiHandleResult,
	taskCh chan int64, peerID peer.ID, fileID string, chunkTotal int64) {

	defer wg.Done()

	stream, err := d.host.NewStream(network.WithUseTransient(ctx, ""), peerID, DOWNLOAD_ID)
	if err != nil {
		log.Errorf("host new stream error: %v", err)
		return
	}
	defer stream.Close()

	wt := pbio.NewDelimitedWriter(stream)
	bufReader := bufio.NewReader(stream)

	for chunkIndex := range taskCh {
		// 发送下载分片请求
		if err = wt.WriteMsg(&pb.FileDownloadChunk{
			FileId:     fileID,
			ChunkIndex: chunkIndex,
			ChunkTotal: chunkTotal,
		}); err != nil {
			// 失败了则还回去
			taskCh <- chunkIndex
			log.Errorf("pbio write msg error: %v", err)
			return
		}

		// 接收分片头信息
		bs64, err := bufReader.ReadBytes('\n')
		if err != nil {
			taskCh <- chunkIndex
			log.Errorf("bufRead stream header error: %v", err)
			return
		}

		// 头信息base64解码
		bs, err := base64.StdEncoding.DecodeString(string(bs64[:len(bs64)-1]))
		if err != nil {
			taskCh <- chunkIndex
			log.Errorf("base64 decode string error: %w", err)
		}

		var chunkInfo pb.FileChunkInfo
		if err := proto.Unmarshal(bs, &chunkInfo); err != nil {
			taskCh <- chunkIndex
			log.Errorf("proto.Unmarshal chunk header error: %v", err)
			return
		}

		// 接收分片数据
		hashSum := md5.New()
		chunkFile := filepath.Join(d.conf.TmpDir, chunkFileName(fileID, chunkIndex))
		ofile, err := os.OpenFile(chunkFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
		if err != nil {
			taskCh <- chunkIndex
			log.Errorf("os open file error: %v", err)
			return
		}
		mulwriter := io.MultiWriter(ofile, hashSum)
		size, err := io.CopyN(mulwriter, bufReader, chunkInfo.ChunkSize)
		// 写完了关闭文件
		ofile.Close()
		if err != nil {
			taskCh <- chunkIndex
			log.Errorf("io copy error: %v", err)
			return
		}

		if size != chunkInfo.ChunkSize || fmt.Sprintf("%x", hashSum.Sum(nil)) != chunkInfo.ChunkHash { // 文件不一致
			taskCh <- chunkIndex
			log.Errorln("chunk file error")
			return
		}

		// 检查是否都下载完成
		if isFinish := mhr.IncreOne(); isFinish {
			return
		}
	}
}

// CopyFileToResource 复制文件到资源目录
func (f *FileProto) CopyFileToResource(ctx context.Context, srcFile string) (string, error) {

	tmpFile, fileID, err := f.calcFileID(ctx, srcFile)
	if err != nil {
		return "", fmt.Errorf("copy to tmp dir error: %w", err)
	}

	// 内部资源目录
	if err := os.Rename(tmpFile, filepath.Join(f.conf.ResourceDir, fileID.String())); err != nil {
		return "", fmt.Errorf("os.Rename error: %w", err)
	}

	return fileID.String(), nil
}

// CopyFileToFile 复制外部文件到内部
func (f *FileProto) CopyFileToFile(ctx context.Context, srcFile string) (string, error) {

	tmpFile, fileID, err := f.calcFileID(ctx, srcFile)
	if err != nil {
		return "", fmt.Errorf("copy to tmp dir error: %w", err)
	}

	// 内部文件目录
	if err := os.Rename(tmpFile, filepath.Join(f.conf.FileDir, fileID.String())); err != nil {
		return "", fmt.Errorf("os.Rename error: %w", err)
	}

	return fileID.String(), nil
}

func (f *FileProto) calcFileID(ctx context.Context, srcFile string) (string, *mytype.FileID, error) {
	// 计算hash
	oSrcFile, err := os.Open(srcFile)
	if err != nil {
		return "", nil, fmt.Errorf("os open file error: %w", err)
	}
	defer oSrcFile.Close()

	fi, err := oSrcFile.Stat()
	if err != nil {
		return "", nil, fmt.Errorf("file state error: %w", err)

	} else if fi.IsDir() {
		return "", nil, fmt.Errorf("file is dir, need file")
	}

	tmpFilePath := filepath.Join(f.conf.TmpDir, fmt.Sprintf("%d.tmp", time.Now().Unix()))

	oTmpFile, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return "", nil, fmt.Errorf("os.OpenFile error: %w", err)
	}
	defer oTmpFile.Close()

	hashSum := md5.New()
	multiWriter := io.MultiWriter(oTmpFile, hashSum)

	size, err := io.Copy(multiWriter, oSrcFile)
	if err != nil {
		os.Remove(tmpFilePath)
		return "", nil, fmt.Errorf("io copy calc file hash error: %w", err)
	}

	if fi.Size() != size {
		os.Remove(tmpFilePath)
		return "", nil, fmt.Errorf("io copy size error, filesize: %d, copysize: %d", fi.Size(), size)
	}

	return tmpFilePath, &mytype.FileID{
		HashAlgo:  "md5",
		HashValue: fmt.Sprintf("%x", hashSum.Sum(nil)),
		Extension: filepath.Ext(srcFile),
	}, nil
}
