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
	"strings"
	"sync"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	ds "github.com/jianbo-zh/dchat/datastore/ds/fileds"
	myevent "github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/internal/protocol"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/filepb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

var log = logging.Logger("download")

var StreamTimeout = 1 * time.Minute

const (
	QUERY_ID    = protocol.FileQueryID_v100
	DOWNLOAD_ID = protocol.FileDownloadID_v100
	AVATAR_ID   = protocol.AvatarDownloadID_v100

	ServiceName = "peer.message"
	maxMsgSize  = 4 * 1024 // 4K

	ChunkSize = 256 * 1024 // 256K
)

type DownloadProto struct {
	host myhost.Host
	data ds.DownloadIface

	conf config.FileServiceConfig

	emitters struct {
		evtDownloadResult event.Emitter
	}
}

func NewDownloadProto(conf config.FileServiceConfig, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus) (*DownloadProto, error) {
	var err error
	download := DownloadProto{
		host: lhost,
		data: ds.DownloadWrap(ids),
		conf: conf,
	}

	lhost.SetStreamHandler(QUERY_ID, download.fileQueryHandler)
	lhost.SetStreamHandler(DOWNLOAD_ID, download.fileDownloadHandler)
	lhost.SetStreamHandler(AVATAR_ID, download.avatarDownloadHandler)

	// 触发器
	download.emitters.evtDownloadResult, err = ebus.Emitter(&myevent.EvtDownloadResult{})
	if err != nil {
		return nil, fmt.Errorf("ebus emitter error: %w", err)
	}

	// 订阅器
	sub, err := ebus.Subscribe([]any{new(myevent.EvtDownloadRequest), new(myevent.EvtCheckAvatar)})
	if err != nil {
		return nil, fmt.Errorf("ebus subscribe error: %w", err)
	}

	go download.subscribeHandler(context.Background(), sub)

	return &download, nil
}

func (d *DownloadProto) subscribeHandler(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}

			switch evt := e.(type) {
			case myevent.EvtDownloadRequest:
				log.Debugln("receive download request")
				go d.goDownload(evt.FromPeerIDs, evt.FileName, evt.FileSize, evt.HashAlgo, evt.HashValue)

			case myevent.EvtCheckAvatar:
				if evt.Avatar != "" && len(evt.PeerIDs) > 0 {
					if _, err := os.Stat(filepath.Join(d.conf.ResourceDir, evt.Avatar)); os.IsNotExist(err) {
						// 头像不存在则执行下载
						go func() {
							for _, peerID := range evt.PeerIDs {
								if err = d.AvatarDownload(ctx, peerID, evt.Avatar); err == nil {
									break
								}
							}
						}()
					}

				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func (d *DownloadProto) goDownload(peerIDs []peer.ID, fName string, fSize int64, hashAlgo string, hashValue string) {

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
		go d.goQuery(&wg, peerID, hashAlgo, hashValue, resultCh)
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
		go d.goDownloadPart(&wg2, peerID, hashAlgo, hashValue, ChunkSize, fileCh, fileCacheDir)
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

type PeerQueryResult struct {
	PeerID     peer.ID
	FileExists bool
}

func (d *DownloadProto) goQuery(wg *sync.WaitGroup, peerID peer.ID, hashAlgo string, hashValue string, resultCh chan peer.ID) {
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

func (d *DownloadProto) goDownloadPart(wg *sync.WaitGroup, peerID peer.ID, hashAlgo string, hashValue string, chunkSize int64, fileCh chan int64, fileCacheDir string) {
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

// 文件查询处理
func (d *DownloadProto) fileQueryHandler(stream network.Stream) {

}

// 文件下载处理
func (d *DownloadProto) fileDownloadHandler(stream network.Stream) {

}

func (d *DownloadProto) AvatarDownload(ctx context.Context, peerID peer.ID, avatar string) error {
	log.Debugln("do download peer avatar")

	// 检查文件在不在
	avatarPath := path.Join(d.conf.ResourceDir, avatar)
	if fi, err := os.Stat(avatarPath); err == nil {
		if fi.Size() > 0 {
			log.Warnln("avatar file exists")
			return nil
		}

	} else if !os.IsNotExist(err) {
		return fmt.Errorf("os stat error: %w", err)
	}

	// 打开输入流下载文件
	stream, err := d.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, time.Second), ""), peerID, AVATAR_ID)
	if err != nil {
		return fmt.Errorf("new stream error: %w", err)
	}
	defer stream.Close()

	bfwt := bufio.NewWriter(stream)
	if _, err := bfwt.WriteString(fmt.Sprintf("%s\n", avatar)); err != nil {
		return fmt.Errorf("bufio write string error: %w", err)
	}
	bfwt.Flush()

	// 打开文件，接收下载文件
	avatarTmpPath := avatarPath + ".tmp"
	if _, err := os.Stat(avatarTmpPath); err == nil {
		if err = os.Remove(avatarTmpPath); err != nil {
			return fmt.Errorf("os remove error: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("os stat error: %w", err)
	}

	file, err := os.OpenFile(avatarTmpPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("os.OpenFile error: %w", err)
	}

	bfrd := bufio.NewReader(stream)
	if _, err := bfrd.WriteTo(file); err != nil {
		file.Close()
		return fmt.Errorf("bufio write to file error: %s", err.Error())
	}

	file.Close()
	if err = os.Rename(avatarTmpPath, avatarPath); err != nil {
		return fmt.Errorf("os rename error: %w", err)
	}

	return nil
}

// AvatarHandler 下载账号头像
func (d *DownloadProto) avatarDownloadHandler(stream network.Stream) {
	log.Debugln("handle download peer avatar")

	defer stream.Close()
	bfrd := bufio.NewReader(stream)
	avatar, err := bfrd.ReadString('\n')
	if err != nil {
		log.Errorf("bufio read avatar error: %s", err.Error())
		return
	}
	avatarPath := path.Join(d.conf.ResourceDir, strings.TrimSpace(avatar))
	file, err := os.Open(avatarPath)
	if err != nil {
		log.Errorf("os open avatar error: %s", err.Error())
		return
	}
	defer file.Close()

	bfrw := bufio.NewWriter(stream)
	_, err = bfrw.ReadFrom(file)
	if err != nil {
		log.Errorf("bufio read from file error: %s", err.Error())
		return
	}
	bfrw.Flush()
}
