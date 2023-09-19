package accountproto

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/protocol"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto/ds"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

var log = logging.Logger("account-protocol")

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/account.proto=./pb pb/account.proto

var StreamTimeout = 1 * time.Minute

const (
	ID         = protocol.AccountID_v100
	AVATAR_ID  = protocol.AccountAvatarID_v100
	maxMsgSize = 4 * 1024 // 4K
)

type AccountProto struct {
	host      host.Host
	data      ds.AccountIface
	avatarDir string
}

func NewAccountProto(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus, avatarDir string) (*AccountProto, error) {
	accountProto := AccountProto{
		host:      lhost,
		data:      ds.Wrap(ids),
		avatarDir: avatarDir,
	}

	lhost.SetStreamHandler(ID, accountProto.GetPeerHandler)
	lhost.SetStreamHandler(AVATAR_ID, accountProto.DownloadPeerAvatarHandler)

	return &accountProto, nil
}

func (a *AccountProto) CreateAccount(ctx context.Context, account *pb.Account) (*pb.Account, error) {
	_, err := a.data.GetAccount(ctx)
	if err == nil {
		return nil, fmt.Errorf("account is exists")

	} else if !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("get account error: %w", err)
	}

	account.PeerId = []byte(a.host.ID())
	err = a.data.CreateAccount(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("ds create account error: %w", err)
	}

	return account, nil
}

func (a *AccountProto) GetAccount(ctx context.Context) (*pb.Account, error) {
	account, err := a.data.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("ds get account error: %w", err)
	}

	return account, nil
}

func (a *AccountProto) UpdateAccount(ctx context.Context, account *pb.Account) error {
	if err := a.data.UpdateAccount(ctx, account); err != nil {
		return fmt.Errorf("ds update account error: %w", err)
	}
	return nil
}

func (a *AccountProto) GetPeer(ctx context.Context, peerID peer.ID) (*pb.Peer, error) {
	log.Debugln("do get peer")

	stream, err := a.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, time.Second), ""), peerID, ID)
	if err != nil {
		return nil, fmt.Errorf("new stream error: %w", err)
	}

	stream.SetDeadline(time.Now().Add(StreamTimeout))
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rd.Close()

	var peerInfo pb.Peer
	if err = rd.ReadMsg(&peerInfo); err != nil {
		stream.Reset()
		return nil, fmt.Errorf("pbio read msg error: %w", err)
	}

	return &peerInfo, nil
}

func (a *AccountProto) DownloadPeerAvatar(ctx context.Context, peerID peer.ID, avatar string) error {
	log.Debugln("do download peer avatar")

	// 检查文件在不在
	avatarPath := path.Join(a.avatarDir, avatar)
	if fi, err := os.Stat(avatarPath); err == nil {
		if fi.Size() > 0 {
			log.Warnln("avatar file exists")
			return nil
		}

	} else if !os.IsNotExist(err) {
		return fmt.Errorf("os stat error: %w", err)
	}

	// 打开输入流下载文件
	stream, err := a.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, time.Second), ""), peerID, AVATAR_ID)
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
	file, err := os.OpenFile(avatarPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("os.OpenFile error: %w", err)
	}
	defer file.Close()

	bfrd := bufio.NewReader(stream)
	if _, err := bfrd.WriteTo(file); err != nil {
		return fmt.Errorf("bufio write to file error: %s", err.Error())
	}

	return nil
}

// PeerHandler 查询Peer信息
func (a *AccountProto) GetPeerHandler(stream network.Stream) {
	log.Debugln("handle get peer")

	wt := pbio.NewDelimitedWriter(stream)
	defer wt.Close()

	account, err := a.data.GetAccount(context.Background())
	if err != nil {
		log.Errorf("get account error: %s", err.Error())
		return
	}

	if err := wt.WriteMsg(&pb.Peer{
		PeerId: account.PeerId,
		Name:   account.Name,
		Avatar: account.Avatar,
	}); err != nil {
		log.Errorf("pbio write peer msg error: %s", err.Error())
		return
	}
}

// AvatarHandler 下载账号头像
func (a *AccountProto) DownloadPeerAvatarHandler(stream network.Stream) {
	log.Debugln("handle download peer avatar")

	defer stream.Close()
	bfrd := bufio.NewReader(stream)
	avatar, err := bfrd.ReadString('\n')
	if err != nil {
		log.Errorf("bufio read avatar error: %s", err.Error())
		return
	}
	avatarPath := path.Join(a.avatarDir, strings.TrimSpace(avatar))
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
