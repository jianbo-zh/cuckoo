package contactproto

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/protocol"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/contactproto/ds"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/contactproto/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

var log = logging.Logger("contact")

var StreamTimeout = 1 * time.Minute

const (
	ID = protocol.ContactID_v100

	ServiceName = "peer.contact"
	maxMsgSize  = 4 * 1024 // 4K
)

type ContactProto struct {
	host host.Host
	data ds.PeerIface

	accountGetter types.AccountGetter

	emitters struct {
		evtSyncPeerMessage event.Emitter
		evtApplyAddContact event.Emitter
	}
}

func NewContactProto(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus, accountGetter types.AccountGetter) (*ContactProto, error) {
	var err error
	contactsvc := ContactProto{
		host:          lhost,
		data:          ds.Wrap(ids),
		accountGetter: accountGetter,
	}

	contactsvc.host.SetStreamHandler(ID, contactsvc.handler)

	if contactsvc.emitters.evtSyncPeerMessage, err = eventBus.Emitter(&gevent.EvtSyncPeerMessage{}); err != nil {
		return nil, fmt.Errorf("set sync peers emitter error: %w", err)
	}

	if contactsvc.emitters.evtApplyAddContact, err = eventBus.Emitter(&gevent.EvtApplyAddContact{}); err != nil {
		return nil, fmt.Errorf("set apply add contact emitter error: %w", err)
	}

	sub, err := eventBus.Subscribe([]any{new(gevent.EvtHostBootComplete), new(gevent.EvtReceivePeerStream), new(gevent.EvtAccountPeerChange)})
	if err != nil {
		return nil, fmt.Errorf("subscribe boot complete error: %w", err)

	} else {
		go contactsvc.handleSubscribe(context.Background(), sub)
	}

	return &contactsvc, nil
}

func (c *ContactProto) handler(stream network.Stream) {

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	defer rd.Close()

	// 读取对方数据
	var msg pb.Peer
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	// 更新本地
	ctx := context.Background()
	if err := c.data.UpdateContact(ctx, &pb.Contact{
		Id:            msg.Id,
		Name:          msg.Name,
		Avatar:        msg.Avatar,
		DepositPeerId: msg.DepositPeerId,
	}); err != nil {
		log.Errorf("data update contact error: %v", err)
		stream.Reset()
		return
	}

	// 发送数据给对方
	account, err := c.accountGetter.GetAccount(ctx)
	if err != nil {
		log.Errorf("get account error: %w", err)
		stream.Reset()
		return
	}

	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&pb.Contact{
		Id:            []byte(account.ID),
		Name:          account.Name,
		Avatar:        account.Avatar,
		DepositPeerId: []byte(account.DepositPeerID),
	}); err != nil {
		log.Errorf("pbio write msg error: %w", err)
		stream.Reset()
		return
	}
}

func (c *ContactProto) goSync(contactID peer.ID, accountPeer types.AccountPeer, isBootSync bool) {

	ctx := context.Background()
	stream, err := c.host.NewStream(ctx, contactID, ID)
	if err != nil {
		log.Errorf("host new stream error: %w", err)
		return
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, maxMsgSize)
	wt := pbio.NewDelimitedWriter(stream)

	// 发送数据给对方
	if err = wt.WriteMsg(&pb.Peer{
		Id:            []byte(accountPeer.ID),
		Name:          accountPeer.Name,
		Avatar:        accountPeer.Avatar,
		DepositPeerId: []byte(accountPeer.DepositPeerID),
	}); err != nil {
		log.Errorf("pbio write msg error: %w", err)
		stream.Reset()
		return
	}

	// 读取对方数据
	var msg pb.Peer
	if err = rd.ReadMsg(&msg); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	// 更新本地数据
	if err = c.data.UpdateContact(ctx, &pb.Contact{
		Id:            msg.Id,
		Name:          msg.Name,
		Avatar:        msg.Avatar,
		DepositPeerId: msg.DepositPeerId,
	}); err != nil {
		log.Errorf("data update contact error: %v", err)
		stream.Reset()
		return
	}

	if isBootSync {
		// 启动时同步，还要触发同步消息事件
		c.emitters.evtSyncPeerMessage.Emit(gevent.EvtSyncPeerMessage{
			ContactID: contactID,
		})
	}
}

func (c *ContactProto) handleSubscribe(ctx context.Context, sub event.Subscription) {
	defer sub.Close()

	for {
		select {
		case e, ok := <-sub.Out():
			if !ok {
				return
			}
			switch ev := e.(type) {
			case gevent.EvtHostBootComplete:
				if !ev.IsSucc {
					log.Warnf("host boot complete but not succ")
					continue
				}

				if contactIDs, err := c.data.GetSessionIDs(ctx); err != nil {
					log.Warnf("get peer ids error: %v", err)
					continue

				} else if len(contactIDs) > 0 {
					account, err := c.accountGetter.GetAccount(context.Background())
					if err != nil {
						log.Errorf("get account error: %w", err)
						continue
					}
					for _, contactID := range contactIDs {
						go c.goSync(contactID, types.AccountPeer{
							ID:            account.ID,
							Name:          account.Name,
							Avatar:        account.Avatar,
							DepositPeerID: account.DepositPeerID,
						}, true)
					}
				}
			case gevent.EvtReceivePeerStream:
				state, err := c.data.GetState(ctx, ev.PeerID)
				if err != nil {
					log.Warnf("get state error: %v", err)
					continue
				}

				if state == "apply" {
					if err := c.data.SetSession(ctx, ev.PeerID); err != nil {
						log.Warnf("set session error: %v", err)
						continue
					}
				}
			case gevent.EvtAccountPeerChange:
				if contactIDs, err := c.data.GetSessionIDs(ctx); err != nil {
					log.Warnf("get peer ids error: %v", err)
					continue

				} else if len(contactIDs) > 0 {
					for _, contactID := range contactIDs {
						go c.goSync(contactID, ev.AccountPeer, false)
					}
				}

			default:
				log.Warnf("undefined event type: %T", ev)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (c *ContactProto) ApplyAddContact(ctx context.Context, peer0 *types.Peer, content string) error {

	if err := c.data.AddContact(ctx, &pb.Contact{
		Id:         []byte(peer0.ID),
		Name:       peer0.Name,
		Avatar:     peer0.Avatar,
		CreateTime: time.Now().Unix(),
		AccessTime: time.Now().Unix(),
	}); err != nil {
		return fmt.Errorf("data.AddContact error: %w", err)
	}

	if err := c.data.SetState(ctx, peer0.ID, "apply"); err != nil {
		return fmt.Errorf("data set session error: %w", err)
	}

	if err := c.emitters.evtApplyAddContact.Emit(gevent.EvtApplyAddContact{
		PeerID:  peer0.ID,
		Content: content,
	}); err != nil {
		return fmt.Errorf("emit apply add contact error: %w", err)
	}

	return nil
}

func (c *ContactProto) AgreeAddContact(ctx context.Context, peer0 *types.Peer) error {

	if err := c.data.AddContact(ctx, &pb.Contact{
		Id:         []byte(peer0.ID),
		Name:       peer0.Name,
		Avatar:     peer0.Avatar,
		CreateTime: time.Now().Unix(),
		AccessTime: time.Now().Unix(),
	}); err != nil {
		return fmt.Errorf("data.AddContact error: %w", err)
	}

	if err := c.data.SetState(ctx, peer0.ID, "normal"); err != nil {
		return fmt.Errorf("data set session error: %w", err)
	}

	if err := c.data.SetSession(ctx, peer0.ID); err != nil {
		return fmt.Errorf("data set session error: %w", err)
	}

	// 启动同步，连接对方
	account, err := c.accountGetter.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("get account error: %w", err)
	}

	go c.goSync(peer0.ID, types.AccountPeer{
		ID:            account.ID,
		Name:          account.Name,
		Avatar:        account.Avatar,
		DepositPeerID: account.DepositPeerID,
	}, false)

	return nil
}

func (c *ContactProto) GetContact(ctx context.Context, peerID peer.ID) (*types.Contact, error) {
	contact, err := c.data.GetContact(ctx, peerID)
	if err != nil {
		return nil, err
	}

	return &types.Contact{
		ID:     peer.ID(contact.Id),
		Name:   contact.Name,
		Avatar: contact.Avatar,
	}, nil
}

func (c *ContactProto) GetContactsByPeerIDs(ctx context.Context, peerIDs []peer.ID) ([]*pb.Contact, error) {
	return c.data.GetContactsByIDs(ctx, peerIDs)
}

func (c *ContactProto) GetContacts(ctx context.Context) ([]*types.Contact, error) {
	contacts, err := c.data.GetContacts(ctx)
	if err != nil {
		return nil, err
	}

	var peers []*types.Contact
	for _, peeri := range contacts {
		peers = append(peers, &types.Contact{
			ID:     peer.ID(peeri.Id),
			Name:   peeri.Name,
			Avatar: peeri.Avatar,
		})
	}

	return peers, nil
}

func (c *ContactProto) SetContactName(ctx context.Context, peerID peer.ID, name string) error {
	contact, err := c.data.GetContact(ctx, peerID)
	if err != nil {
		return fmt.Errorf("ds get contact error: %w", err)
	}

	contact.Name = name
	if err = c.data.UpdateContact(ctx, contact); err != nil {
		return fmt.Errorf("ds update contact error: %w", err)
	}

	return nil
}

func (c *ContactProto) DeleteContact(ctx context.Context, peerID peer.ID) error {
	return c.data.DeleteContact(ctx, peerID)
}
