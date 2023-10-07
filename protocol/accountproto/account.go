package accountproto

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	ds "github.com/jianbo-zh/dchat/datastore/ds/accountds"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/myprotocol"
	"github.com/jianbo-zh/dchat/internal/mytype"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/accountpb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

var log = logging.Logger("account-protocol")

var StreamTimeout = 1 * time.Minute

const (
	ID         = myprotocol.AccountID_v100
	ONLINE_ID  = myprotocol.AccountOnlineID_v100
	maxMsgSize = 4 * 1024 // 4K
)

type AccountProto struct {
	host      myhost.Host
	data      ds.AccountIface
	avatarDir string

	emitter struct {
		evtAccountPeerChange   event.Emitter
		evtOnlineStateDiscover event.Emitter
	}
}

func NewAccountProto(lhost myhost.Host, ids ipfsds.Batching, eventBus event.Bus, avatarDir string) (*AccountProto, error) {
	var err error
	accountProto := AccountProto{
		host:      lhost,
		data:      ds.Wrap(ids),
		avatarDir: avatarDir,
	}

	lhost.SetStreamHandler(ID, accountProto.GetPeerHandler)
	lhost.SetStreamHandler(ONLINE_ID, accountProto.onlineHandler)

	if accountProto.emitter.evtAccountPeerChange, err = eventBus.Emitter(&myevent.EvtAccountPeerChange{}); err != nil {
		return nil, fmt.Errorf("set account peer change emitter error: %w", err)
	}

	if accountProto.emitter.evtOnlineStateDiscover, err = eventBus.Emitter(&myevent.EvtOnlineStateDiscover{}); err != nil {
		return nil, fmt.Errorf("set peer online state discover emitter error: %w", err)
	}

	return &accountProto, nil
}

func (a *AccountProto) GetOnlineState(peerIDs []peer.ID) map[peer.ID]mytype.OnlineState {

	go a.goCheckOnlineState(peerIDs, 30*time.Second)

	return a.host.OnlineStats(peerIDs, 60*time.Second)
}

func (a *AccountProto) goCheckOnlineState(peerIDs []peer.ID, onlineDuration time.Duration) {

	peersOnlineState := make(map[peer.ID]bool)
	if len(peerIDs) == 0 {
		return
	}

	var checkOnlinePeerIDs []peer.ID
	onlineStats := a.host.OnlineStats(peerIDs, onlineDuration)
	for _, peerID := range peerIDs {
		if onlineStats[peerID] == mytype.OnlineStateOnline {
			peersOnlineState[peerID] = true
		} else {
			checkOnlinePeerIDs = append(checkOnlinePeerIDs, peerID)
		}
	}

	if len(checkOnlinePeerIDs) > 0 {
		onlineCh := make(chan mytype.PeerState, len(checkOnlinePeerIDs))

		st := time.Now()
		var wg sync.WaitGroup
		wg.Add(len(checkOnlinePeerIDs))
		for _, peerID := range checkOnlinePeerIDs {
			go a.goCheckOnline(&wg, peerID, onlineCh)
		}
		wg.Wait()
		close(onlineCh)

		log.Debugf("check online use time: %d", time.Since(st).Milliseconds())

		for res := range onlineCh {
			if res.IsOnline {
				peersOnlineState[res.PeerID] = true
			} else {
				peersOnlineState[res.PeerID] = false
			}
		}

		// 触发在线更新事件发送到前端
		err := a.emitter.evtOnlineStateDiscover.Emit(myevent.EvtOnlineStateDiscover{
			OnlineState: peersOnlineState,
		})
		if err != nil {
			log.Errorf("emit peer online state discover error: %v", err)
		}
	}
}

func (a *AccountProto) goCheckOnline(wg *sync.WaitGroup, peerID peer.ID, onlineCh chan<- mytype.PeerState) {
	defer wg.Done()

	peerState := mytype.PeerState{
		PeerID:   peerID,
		IsOnline: false, // default: offline
	}

	ctx := context.Background()
	stream, err := a.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, time.Second), ""), peerID, ONLINE_ID)
	if err != nil {
		onlineCh <- peerState
		log.Debugln("host new stream error: ", err.Error())
		return
	}
	defer stream.Close()

	wt := pbio.NewDelimitedWriter(stream)
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)

	stream.SetDeadline(time.Now().Add(2 * time.Second))
	if err = wt.WriteMsg(&pb.AccountOnline{IsOnline: true}); err != nil {
		onlineCh <- peerState
		log.Errorf("pbio write msg error: %v", err)
		return
	}

	var pbOnline pb.AccountOnline
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

	var pbOnline pb.AccountOnline
	if err := rd.ReadMsg(&pbOnline); err != nil {
		log.Errorf("pbio read msg error: %v", err)
		return
	}

	if err := wt.WriteMsg(&pb.AccountOnline{IsOnline: true}); err != nil {
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

		a.emitter.evtAccountPeerChange.Emit(myevent.EvtAccountPeerChange{
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

		a.emitter.evtAccountPeerChange.Emit(myevent.EvtAccountPeerChange{
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

		a.emitter.evtAccountPeerChange.Emit(myevent.EvtAccountPeerChange{
			AccountPeer: DecodeAccountPeer(account),
		})
	}

	return nil
}

func (a *AccountProto) GetPeer(ctx context.Context, peerID peer.ID) (*pb.AccountPeer, error) {
	log.Debugln("do get peer")

	stream, err := a.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, mytype.DialTimeout), ""), peerID, ID)
	if err != nil {
		return nil, fmt.Errorf("new stream error: %w", err)
	}

	stream.SetDeadline(time.Now().Add(StreamTimeout))
	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rd.Close()

	var peerInfo pb.AccountPeer
	if err = rd.ReadMsg(&peerInfo); err != nil {
		stream.Reset()
		return nil, fmt.Errorf("pbio read msg error: %w", err)
	}

	return &peerInfo, nil
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

	if err := wt.WriteMsg(&pb.AccountPeer{
		Id:     account.Id,
		Name:   account.Name,
		Avatar: account.Avatar,
	}); err != nil {
		log.Errorf("pbio write peer msg error: %s", err.Error())
		return
	}
}
