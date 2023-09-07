package accountproto

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/host"
	"github.com/jianbo-zh/dchat/protocolid"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto/ds"
	"github.com/jianbo-zh/dchat/service/accountsvc/protocol/accountproto/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ipeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

var log = logging.Logger("account")

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/account.proto=./pb pb/account.proto

var StreamTimeout = 1 * time.Minute

const (
	ID         = protocolid.ACCOUNT_ID
	AVATAR_ID  = protocolid.ACCOUNT_SYNC_AVATAR_ID
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

	lhost.SetStreamHandler(ID, accountProto.PeerHandler)
	lhost.SetStreamHandler(AVATAR_ID, accountProto.PeerAvatarHandler)

	return &accountProto, nil
}

func (a *AccountProto) CreateAccount(ctx context.Context, account Account) (*Account, error) {

	_, err := a.data.GetAccount(ctx)
	if !errors.Is(err, ipfsds.ErrNotFound) {
		return nil, fmt.Errorf("account is exists")
	}

	rAccount := Account{
		PeerID:         a.host.ID(),
		Name:           account.Name,
		Avatar:         account.Avatar,
		AutoAddContact: account.AutoAddContact,
		AutoJoinGroup:  account.AutoJoinGroup,
	}

	err = a.data.CreateAccount(ctx, &pb.AccountMsg{
		PeerId:         []byte(rAccount.PeerID),
		Name:           rAccount.Name,
		Avatar:         rAccount.Avatar,
		AutoAddContact: rAccount.AutoAddContact,
		AutoJoinGroup:  rAccount.AutoJoinGroup,
	})
	if err != nil {
		return nil, fmt.Errorf("a.data.CreateAccount error: %w", err)
	}

	return &rAccount, nil
}

func (a *AccountProto) GetAccount(ctx context.Context) (*Account, error) {
	pAccount, err := a.data.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("a.data.GetAccount error: %w", err)
	}

	return &Account{
		PeerID:         ipeer.ID(pAccount.PeerId),
		Name:           pAccount.Name,
		Avatar:         pAccount.Avatar,
		AutoAddContact: pAccount.AutoAddContact,
		AutoJoinGroup:  pAccount.AutoJoinGroup,
	}, nil
}

func (a *AccountProto) UpdateAccount(ctx context.Context, account Account) error {

	return a.data.UpdateAccount(ctx, &pb.AccountMsg{
		PeerId:         []byte(account.PeerID),
		Name:           account.Name,
		Avatar:         account.Avatar,
		AutoAddContact: account.AutoAddContact,
		AutoJoinGroup:  account.AutoJoinGroup,
	})
}

func (a *AccountProto) GetPeer(ctx context.Context, peerID peer.ID) (*Peer, error) {

	stream, err := a.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, time.Second), ""), peerID, ID)
	if err != nil {
		return nil, fmt.Errorf("a.host.NewStream error: %w", err)
	}

	stream.SetDeadline(time.Now().Add(StreamTimeout))
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rd.Close()

	var peerInfo pb.PeerMsg
	if err = rd.ReadMsg(&peerInfo); err != nil {
		stream.Reset()
		return nil, fmt.Errorf("rd.ReadMsg error: %w", err)
	}

	return &Peer{
		PeerID: peer.ID(peerInfo.PeerId),
		Name:   peerInfo.Name,
		Avatar: peerInfo.Avatar,
	}, nil
}

func (a *AccountProto) DownloadPeerAvatar(ctx context.Context, peerID peer.ID, avatar string) error {

	avatarPath := path.Join(a.avatarDir, avatar)

	// 检查文件在不在
	if _, err := os.Stat(avatarPath); err == nil {
		log.Warnf("os.Stat avatar file exists")
		return nil

	} else if !os.IsNotExist(err) {
		return fmt.Errorf("os.Stat error: %w", err)
	}

	// 打开输入流下载文件
	stream, err := a.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, time.Second), ""), peerID, AVATAR_ID)
	if err != nil {
		return fmt.Errorf("a.host.NewStream error: %w", err)
	}
	defer stream.Close()

	bfwt := bufio.NewWriter(stream)
	if _, err = bfwt.WriteString(avatar); err != nil {
		return fmt.Errorf("bfwt.WriteString error: %w", err)
	}

	// 打开文件，接收下载文件
	file, err := os.OpenFile(avatarPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("os.OpenFile error: %w", err)
	}
	defer file.Close()

	bfrd := bufio.NewReader(stream)

	if n, err := bfrd.WriteTo(file); err != nil {
		return fmt.Errorf("bfrd.WriteTo error: %s", err.Error())
	} else {
		log.Debugf("bfrd.WriteTo size :%d", n)
	}

	return nil
}

// PeerHandler 查询Peer信息
func (a *AccountProto) PeerHandler(stream network.Stream) {

	wt := pbio.NewDelimitedWriter(stream)
	defer wt.Close()

	account, err := a.data.GetAccount(context.Background())
	if err != nil {
		log.Errorf("a.data.GetAccount error: %s", err.Error())
		return
	}

	if err := wt.WriteMsg(&pb.PeerMsg{
		PeerId: account.PeerId,
		Name:   account.Name,
		Avatar: account.Avatar,
	}); err != nil {
		log.Errorf("wt.WriteMsg error: %s", err.Error())
		return
	}
}

// AvatarHandler 下载账号头像
func (a *AccountProto) PeerAvatarHandler(stream network.Stream) {

	defer stream.Close()
	bfrd := bufio.NewReader(stream)
	avatar, err := bfrd.ReadString('\n')
	if err != nil {
		log.Errorf("bfrd.ReadString error: %s", err.Error())
	}

	avatarPath := path.Join(a.avatarDir, avatar)
	file, err := os.Open(avatarPath)
	if err != nil {
		return
	}
	defer file.Close()

	bfrw := bufio.NewWriter(stream)
	n, err := bfrw.ReadFrom(file)
	if err != nil {
		log.Errorf("bfrw.ReadFrom error: %s", err.Error())
	}
	bfrw.Flush()
	log.Debugf("bfrw.ReadFrom size: %d", n)
}
