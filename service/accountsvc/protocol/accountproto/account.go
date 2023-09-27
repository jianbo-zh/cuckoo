package accountproto

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/protocol"
	"github.com/jianbo-zh/dchat/internal/types"
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

var StreamTimeout = 1 * time.Minute

var OnlineTimeMap = make(map[peer.ID]int64)

const (
	ID            = protocol.AccountID_v100
	ONLINE_ID     = protocol.AccountOnlineID_v100
	AVATAR_ID     = protocol.AccountAvatarID_v100
	maxMsgSize    = 4 * 1024 // 4K
	OfflineSecond = 60       // 60s
)

type AccountProto struct {
	host      host.Host
	data      ds.AccountIface
	avatarDir string

	emitter struct {
		evtAccountPeerChange           event.Emitter
		evtAccountDepositServiceChange event.Emitter
	}
}

func NewAccountProto(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus, avatarDir string) (*AccountProto, error) {
	var err error
	accountProto := AccountProto{
		host:      lhost,
		data:      ds.Wrap(ids),
		avatarDir: avatarDir,
	}

	lhost.SetStreamHandler(ID, accountProto.GetPeerHandler)
	lhost.SetStreamHandler(ONLINE_ID, accountProto.onlineHandler)
	lhost.SetStreamHandler(AVATAR_ID, accountProto.DownloadPeerAvatarHandler)

	if accountProto.emitter.evtAccountPeerChange, err = eventBus.Emitter(&gevent.EvtAccountPeerChange{}); err != nil {
		return nil, fmt.Errorf("set account peer change emitter error: %w", err)
	}

	if accountProto.emitter.evtAccountDepositServiceChange, err = eventBus.Emitter(&gevent.EvtAccountDepositServiceChange{}); err != nil {
		return nil, fmt.Errorf("set account deposit service change emitter error: %w", err)
	}

	return &accountProto, nil
}

func (a *AccountProto) GetOnlineState(ctx context.Context, peerIDs []peer.ID) (map[peer.ID]bool, error) {

	nowtime := time.Now().Unix()
	result := make(map[peer.ID]bool)

	if len(peerIDs) == 0 {
		return result, nil
	}

	// 检查缓存
	var maybeOfflinePeers []peer.ID
	for _, peerID := range peerIDs {
		if nowtime-OnlineTimeMap[peerID] < OfflineSecond {
			result[peerID] = true
		} else {
			maybeOfflinePeers = append(maybeOfflinePeers, peerID)
		}
	}
	if len(maybeOfflinePeers) == 0 { // 都在线
		return result, nil
	}

	onlineCh := make(chan types.PeerState, len(maybeOfflinePeers))

	st := time.Now()
	var wg sync.WaitGroup
	wg.Add(len(maybeOfflinePeers))
	for _, peerID := range maybeOfflinePeers {
		go a.goCheckOnline(ctx, &wg, peerID, onlineCh)
	}
	wg.Wait()
	close(onlineCh)

	log.Debugf("check online use time: %d", time.Since(st).Milliseconds())

	for res := range onlineCh {
		if res.IsOnline {
			result[res.PeerID] = true
			OnlineTimeMap[res.PeerID] = time.Now().Unix()

		} else {
			result[res.PeerID] = false
		}
	}

	return result, nil
}

func (a *AccountProto) goCheckOnline(ctx context.Context, wg *sync.WaitGroup, peerID peer.ID, onlineCh chan<- types.PeerState) {
	defer wg.Done()

	peerState := types.PeerState{
		PeerID:   peerID,
		IsOnline: false,
	}

	stream, err := a.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, time.Second), ""), peerID, ONLINE_ID)
	if err != nil {
		onlineCh <- peerState
		log.Debugln("host new stream error: %v", err)
		return
	}
	defer stream.Close()

	wt := pbio.NewDelimitedWriter(stream)
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)

	stream.SetDeadline(time.Now().Add(2 * time.Second))
	if err = wt.WriteMsg(&pb.Online{IsOnline: true}); err != nil {
		onlineCh <- peerState
		log.Errorf("pbio write msg error: %v", err)
		return
	}

	var pbOnline pb.Online
	if err = rd.ReadMsg(&pbOnline); err != nil {
		onlineCh <- peerState
		log.Errorf("pbio read msg error: %v", err)
		return
	}

	peerState.IsOnline = true
	onlineCh <- peerState
}

func (a *AccountProto) onlineHandler(stream network.Stream) {
	defer stream.Close()

	wt := pbio.NewDelimitedWriter(stream)
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)

	var pbOnline pb.Online
	if err := rd.ReadMsg(&pbOnline); err != nil {
		log.Errorf("pbio read msg error: %v", err)
		return
	}

	if err := wt.WriteMsg(&pb.Online{IsOnline: true}); err != nil {
		log.Errorf("pbio wt msg error: %v", err)
		return
	}
}

func (a *AccountProto) InitAccount(ctx context.Context) error {
	_, err := a.data.GetAccount(ctx)
	if err != nil {
		if errors.Is(err, ipfsds.ErrNotFound) {
			if err = a.data.CreateAccount(ctx, &pb.Account{
				Id: []byte(a.host.ID()),
			}); err != nil {
				return fmt.Errorf("data create account error: %w", err)
			}
		} else {
			return fmt.Errorf("data get account error: %w", err)
		}
	}
	return nil
}

func (a *AccountProto) CreateAccount(ctx context.Context, account *pb.Account) (*pb.Account, error) {
	account.Id = []byte(a.host.ID())
	if err := a.data.UpdateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("ds update account error: %w", err)
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

func (a *AccountProto) UpdateAccountName(ctx context.Context, name string) error {
	account, err := a.data.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("data get account error: %w", err)
	}
	if account.Name != name {
		account.Name = name
		if err = a.data.UpdateAccount(ctx, account); err != nil {
			return fmt.Errorf("data update account error: %w", err)
		}

		a.emitter.evtAccountPeerChange.Emit(gevent.EvtAccountPeerChange{
			AccountPeer: DecodeAccountPeer(account),
		})
	}

	return nil
}

func (a *AccountProto) UpdateAccountAvatar(ctx context.Context, avatar string) error {
	account, err := a.data.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("data get account error: %w", err)
	}
	if account.Avatar != avatar {
		account.Avatar = avatar
		if err = a.data.UpdateAccount(ctx, account); err != nil {
			return fmt.Errorf("data update account error: %w", err)
		}

		a.emitter.evtAccountPeerChange.Emit(gevent.EvtAccountPeerChange{
			AccountPeer: DecodeAccountPeer(account),
		})
	}

	return nil
}

func (a *AccountProto) UpdateAccountAutoAddContact(ctx context.Context, autoAddContact bool) error {
	account, err := a.data.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("data get account error: %w", err)
	}
	if account.AutoAddContact != autoAddContact {
		account.AutoAddContact = autoAddContact
		if err = a.data.UpdateAccount(ctx, account); err != nil {
			return fmt.Errorf("data update account error: %w", err)
		}
	}

	return nil
}

func (a *AccountProto) UpdateAccountAutoJoinGroup(ctx context.Context, autoJoinGroup bool) error {
	account, err := a.data.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("data get account error: %w", err)
	}

	if account.AutoJoinGroup != autoJoinGroup {
		account.AutoJoinGroup = autoJoinGroup
		if err = a.data.UpdateAccount(ctx, account); err != nil {
			return fmt.Errorf("data update account error: %w", err)
		}
	}

	return nil
}

func (a *AccountProto) UpdateAccountAutoDepositMessage(ctx context.Context, autoDepositMessage bool) error {
	account, err := a.data.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("data get account error: %w", err)
	}

	if account.AutoDepositMessage != autoDepositMessage {
		account.AutoDepositMessage = autoDepositMessage
		if err = a.data.UpdateAccount(ctx, account); err != nil {
			return fmt.Errorf("data update account error: %w", err)
		}
	}

	return nil
}

func (a *AccountProto) UpdateAccountDepositAddress(ctx context.Context, depositPeerID peer.ID) error {
	account, err := a.data.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("data get account error: %w", err)
	}

	if peer.ID(account.DepositAddress) != depositPeerID {
		account.DepositAddress = []byte(depositPeerID)
		if err = a.data.UpdateAccount(ctx, account); err != nil {
			return fmt.Errorf("data update account error: %w", err)
		}

		a.emitter.evtAccountPeerChange.Emit(gevent.EvtAccountPeerChange{
			AccountPeer: DecodeAccountPeer(account),
		})
	}

	return nil
}

func (a *AccountProto) UpdateAccountEnableDepositService(ctx context.Context, enableDepositService bool) error {
	account, err := a.data.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("data get account error: %w", err)
	}

	if account.EnableDepositService != enableDepositService {
		account.EnableDepositService = enableDepositService
		if err = a.data.UpdateAccount(ctx, account); err != nil {
			return fmt.Errorf("data update account error: %w", err)
		}

		a.emitter.evtAccountDepositServiceChange.Emit(gevent.EvtAccountDepositServiceChange{
			Enable: enableDepositService,
		})
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
		Id:     account.Id,
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
