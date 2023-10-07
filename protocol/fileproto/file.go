package fileproto

import (
	"bufio"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	ds "github.com/jianbo-zh/dchat/datastore/ds/fileds"
	myevent "github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/myprotocol"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/filepb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

var log = logging.Logger("file")

var StreamTimeout = 1 * time.Minute

const (
	QUERY_ID    = myprotocol.FileQueryID_v100
	DOWNLOAD_ID = myprotocol.FileDownloadID_v100

	RESOURCE_UPLOAD_ID   = myprotocol.ResourceUploadID_v100
	RESOURCE_DOWNLOAD_ID = myprotocol.ResourceDownloadID_v100

	ServiceName = "peer.message"
	maxMsgSize  = 4 * 1024 // 4K

	ChunkSize = 256 * 1024 // 256K
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
	sub, err := ebus.Subscribe([]any{new(myevent.EvtDownloadFile), new(myevent.EvtDownloadResource), new(myevent.EvtCheckAvatar), new(myevent.EvtUploadResource)})
	if err != nil {
		return nil, fmt.Errorf("ebus subscribe error: %w", err)
	}

	go file.subscribeHandler(context.Background(), sub)

	return &file, nil
}

func (f *FileProto) SendPeerFile(ctx context.Context, peerID peer.ID, file string) error {

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
	stream, err := f.host.NewStream(network.WithDialPeerTimeout(ctx, time.Second), peerID, RESOURCE_DOWNLOAD_ID)
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

	// 等待鉴权结果
	var replyMsg pb.DownloadResourceReply
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	if err := rd.ReadMsg(&replyMsg); err != nil {
		stream.Reset()
		return fmt.Errorf("pbio.ReadMsg error: %w", err)
	}
	if replyMsg.Error != "" {
		stream.Reset()
		return fmt.Errorf("download reply: %s", replyMsg.Error)
	}

	// 开始接收文件
	tmpFilePath := path.Join(f.conf.ResourceDir, fmt.Sprintf("%s_%s", "tmp", fileID))
	tmpFile, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		stream.Reset()
		return fmt.Errorf("os.OpenFile error: %w", err)
	}

	bufStream := bufio.NewReader(stream)
	if _, err := bufStream.WriteTo(tmpFile); err != nil {
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

func (f *FileProto) CalcFileID(ctx context.Context, filePath string) (*mytype.FileID, error) {
	// 计算hash
	ofile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("os open file error: %w", err)
	}
	defer ofile.Close()

	fi, err := ofile.Stat()
	if err != nil {
		return nil, fmt.Errorf("file state error: %w", err)

	} else if fi.IsDir() {
		return nil, fmt.Errorf("file is dir, need file")
	}

	hashSum := md5.New()

	size, err := io.Copy(hashSum, ofile)
	if err != nil {
		return nil, fmt.Errorf("io copy calc file hash error: %w", err)
	}

	if fi.Size() != size {
		return nil, fmt.Errorf("io copy size error, filesize: %d, copysize: %d", fi.Size(), size)
	}

	fileHash := mytype.FileID{
		HashAlgo:  "md5",
		HashValue: fmt.Sprintf("%x", hashSum.Sum(nil)),
	}

	// 存储结果
	err = f.data.SaveFile(ctx, filePath, fi.Size(), fileHash.HashAlgo, fileHash.HashValue)
	if err != nil {
		return nil, fmt.Errorf("data save file error: %w", err)
	}

	return &fileHash, nil
}

func (d *FileProto) downloadFile(peerIDs []peer.ID, fName string, fSize int64, hashAlgo string, hashValue string) {

	cacheDir := ""
	downloadDir := ""

	fileCacheDir := filepath.Join(cacheDir, fmt.Sprintf("%s%s", hashAlgo, hashValue))

	// 判断文件是否存在，存在则返回
	filePath := filepath.Join(downloadDir, fName)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		log.Errorln("file exists error")
	}

	resultCh := make(chan peer.ID, len(peerIDs))

	var wg sync.WaitGroup
	for _, peerID := range peerIDs {
		wg.Add(1)
		go d.query(&wg, peerID, hashAlgo, hashValue, resultCh)
	}
	wg.Wait()

	// 计算文件任务
	chunkSum := fSize / ChunkSize
	if (fSize % ChunkSize) > 0 {
		chunkSum++
	}

	fileCh := make(chan int64, chunkSum)
	for i := int64(0); i < chunkSum; i++ {
		fileCh <- i
	}

	// exists file peer
	var wg2 sync.WaitGroup
	for peerID := range resultCh {
		wg2.Add(1)
		go d.downloadPart(&wg2, peerID, hashAlgo, hashValue, ChunkSize, fileCh, fileCacheDir)
	}
	wg2.Wait()

	if _, isOk := <-fileCh; isOk {
		// 还有值，则下载失败
		log.Errorln("download failed, task not finish")
		return
	}

	// 分片下载完成，则合并整个文件
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		log.Errorln("file exists error")
		return
	}

	ofile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Errorf("os open file error: %w", err)
		return
	}
	defer ofile.Close()

	hashSum := md5.New()
	multiWriter := io.MultiWriter(ofile, hashSum)
	for i := int64(0); i < chunkSum; i++ {
		chunkFile := filepath.Join(fileCacheDir, fmt.Sprintf("%d.part", i))
		ofile2, err := os.Open(chunkFile)
		if err != nil {
			os.Remove(filePath)

			log.Errorf("os open error: %w", err)
			return
		}

		if _, err = io.Copy(multiWriter, ofile2); err != nil {
			ofile2.Close()
			os.Remove(filePath)

			log.Errorf("io copy error: %w", err)
			return
		}
		ofile2.Close()
	}

	if fmt.Sprintf("%x", hashSum.Sum(nil)) != hashValue {
		os.Remove(filePath)

		log.Errorln("file hash error")
		return
	}

	// 下载完成后，删除分片文件
	for i := int64(0); i < chunkSum; i++ {
		os.Remove(filepath.Join(fileCacheDir, fmt.Sprintf("%d.part", i)))
	}
}

func (d *FileProto) query(wg *sync.WaitGroup, peerID peer.ID, hashAlgo string, hashValue string, resultCh chan peer.ID) {
	defer wg.Done()

	ctx := context.Background()
	stream, err := d.host.NewStream(network.WithDialPeerTimeout(ctx, mytype.DialTimeout), peerID, QUERY_ID)
	if err != nil {
		log.Errorf("host new stream error: %v", err)
		return
	}
	defer stream.Close()

	wt := pbio.NewDelimitedWriter(stream)
	if err = wt.WriteMsg(&pb.FileQuery{HashAlgo: hashAlgo, HashValue: hashValue}); err != nil {
		log.Errorf("pbio write msg error: %v", err)
		return
	}

	var result pb.FileQueryResult
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	if err = rd.ReadMsg(&result); err != nil {
		log.Errorf("pbio read msg error: %v", err)
		return
	}

	resultCh <- peerID

}

func (d *FileProto) downloadPart(wg *sync.WaitGroup, peerID peer.ID, hashAlgo string, hashValue string, chunkSize int64, fileCh chan int64, fileCacheDir string) {
	defer wg.Done()
	ctx := context.Background()
	stream, err := d.host.NewStream(network.WithDialPeerTimeout(ctx, mytype.DialTimeout), peerID, DOWNLOAD_ID)
	if err != nil {
		log.Errorf("host new stream error: %w", err)
		return
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	wt := pbio.NewDelimitedWriter(stream)

	for index := range fileCh {
		msg := pb.FileDownloadChunk{
			HashAlgo:   hashAlgo,
			HashValue:  hashValue,
			ChunkIndex: index,
			ChunkSize:  chunkSize,
		}
		if err = wt.WriteMsg(&msg); err != nil {
			// 失败了则还回去
			fileCh <- index

			log.Errorf("pbio write msg error: %w", err)
			return
		}

		chunkFile := filepath.Join(fileCacheDir, fmt.Sprintf("%d.part", index))
		ofile, err := os.OpenFile(chunkFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
		if err != nil {
			// 失败了则还回去
			fileCh <- index

			log.Errorf("os open file error: %w", err)
			return
		}

		var chunkInfo pb.FileChunkInfo
		if err = rd.ReadMsg(&chunkInfo); err != nil {
			fileCh <- index
			ofile.Close()

			log.Errorf("pbio read msg error: %w", err)
			return
		}

		hashSum := md5.New()

		mulwriter := io.MultiWriter(ofile, hashSum)
		size, err := io.CopyN(mulwriter, stream, chunkInfo.ChunkSize)
		if err != nil {
			fileCh <- index
			ofile.Close()

			log.Errorf("io copy error: %w", err)
			return
		}

		if size != chunkInfo.ChunkSize || fmt.Sprintf("%x", hashSum.Sum(nil)) != chunkInfo.ChunkHash { // 文件不一致
			fileCh <- index
			ofile.Close()

			log.Errorln("chunk file error")
			return
		}

		ofile.Close()
	}
}
