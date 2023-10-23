package accountproto

import (
	"context"
	"errors"
	"fmt"
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
	ACCOUNT_ID = myprotocol.AccountID_v100
	ONLINE_ID  = myprotocol.AccountOnlineID_v100
)

type AccountProto struct {
	host      myhost.Host
	data      ds.AccountIface
	avatarDir string

	emitter struct {
		evtAccountBaseChange  event.Emitter
		evtSyncAccountMessage event.Emitter
		evtSyncSystemMessage  event.Emitter
	}
}

func NewAccountProto(lhost myhost.Host, ids ipfsds.Batching, eventBus event.Bus, avatarDir string) (*AccountProto, error) {
	var err error
	accountProto := AccountProto{
		host:      lhost,
		data:      ds.Wrap(ids),
		avatarDir: avatarDir,
	}

	lhost.SetStreamHandler(ACCOUNT_ID, accountProto.GetPeerHandler)
	lhost.SetStreamHandler(ONLINE_ID, accountProto.onlineHandler)

	if accountProto.emitter.evtSyncAccountMessage, err = eventBus.Emitter(&myevent.EvtSyncAccountMessage{}); err != nil {
		return nil, fmt.Errorf("set peer online state discover emitter error: %w", err)
	}

	if accountProto.emitter.evtSyncSystemMessage, err = eventBus.Emitter(&myevent.EvtSyncSystemMessage{}); err != nil {
		return nil, fmt.Errorf("set peer online state discover emitter error: %w", err)
	}

	if accountProto.emitter.evtAccountBaseChange, err = eventBus.Emitter(&myevent.EvtAccountBaseChange{}); err != nil {
		return nil, fmt.Errorf("set account base change emitter error: %w", err)
	}

	sub, err := eventBus.Subscribe([]any{new(myevent.EvtHostBootComplete)})
	if err != nil {
		return nil, fmt.Errorf("subscribe boot complete event error: %v", err)
	} else {
		go accountProto.subscribeHandler(context.Background(), sub)
	}

	return &accountProto, nil
}

func (a *AccountProto) subscribeHandler(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				log.Errorln("account proto subscribe out not ok")
				return
			}

			switch evt := e.(type) {
			case myevent.EvtHostBootComplete:
				if evt.IsSucc {
					account, err := a.data.GetAccount(ctx)
					if err != nil {
						log.Errorf("data.GetAccount error: %w", err)
					}

					if len(account.DepositAddress) > 0 {
						// 拉取联系人消息
						if err := a.emitter.evtSyncAccountMessage.Emit(myevent.EvtSyncAccountMessage{
							DepositAddress: peer.ID(account.DepositAddress),
						}); err != nil {
							log.Errorf("emit sync account message error: %w", err)

						} else {
							log.Debugln("emit sync account message: ", peer.ID(account.DepositAddress).String())
						}

						// 拉取系统消息
						if err := a.emitter.evtSyncSystemMessage.Emit(myevent.EvtSyncSystemMessage{
							DepositAddress: peer.ID(account.DepositAddress),
						}); err != nil {
							log.Errorf("emit sync account message error: %w", err)

						} else {
							log.Debugln("emit sync account message: ", peer.ID(account.DepositAddress).String())
						}
					}
				}

				// 只会执行一次，直接退出
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func (a *AccountProto) GetOnlineState(peerIDs []peer.ID) map[peer.ID]mytype.OnlineState {

	return a.host.OnlineStats(peerIDs, 60*time.Second)
}

func (a *AccountProto) CheckOnlineState(ctx context.Context, peerID peer.ID) error {

	stream, err := a.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, time.Second), ""), peerID, ONLINE_ID)
	if err != nil {
		return fmt.Errorf("host new stream error: %w", err)
	}
	defer stream.Close()

	fmt.Println("checkOnlineState stream..")

	wt := pbio.NewDelimitedWriter(stream)
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)

	if err = wt.WriteMsg(&pb.AccountOnline{IsOnline: true}); err != nil {
		return fmt.Errorf("pbio write msg error: %w", err)
	}

	fmt.Println("checkOnlineState writemsg..")

	var pbOnline pb.AccountOnline
	if err = rd.ReadMsg(&pbOnline); err != nil {
		return fmt.Errorf("pbio read msg error: %w", err)
	}
	fmt.Println("checkOnlineState readmsg..")

	return nil
}

func (a *AccountProto) onlineHandler(stream network.Stream) {
	defer stream.Close()

	wt := pbio.NewDelimitedWriter(stream)
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)

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

		a.emitter.evtAccountBaseChange.Emit(myevent.EvtAccountBaseChange{
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

		a.emitter.evtAccountBaseChange.Emit(myevent.EvtAccountBaseChange{
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

		a.emitter.evtAccountBaseChange.Emit(myevent.EvtAccountBaseChange{
			AccountPeer: DecodeAccountPeer(account),
		})
	}

	return nil
}

func (a *AccountProto) GetPeer(ctx context.Context, peerID peer.ID) (*pb.AccountPeer, error) {
	log.Debugln("do get peer")

	stream, err := a.host.NewStream(network.WithUseTransient(network.WithDialPeerTimeout(ctx, mytype.DialTimeout), ""), peerID, ACCOUNT_ID)
	if err != nil {
		return nil, fmt.Errorf("new stream error: %w", err)
	}

	stream.SetDeadline(time.Now().Add(StreamTimeout))
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
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
		Id:             account.Id,
		Name:           account.Name,
		Avatar:         account.Avatar,
		DepositAddress: account.DepositAddress,
	}); err != nil {
		log.Errorf("pbio write peer msg error: %s", err.Error())
		return
	}
}
