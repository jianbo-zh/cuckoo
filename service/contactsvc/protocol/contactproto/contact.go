package peer

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/contactproto/ds"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/contactproto/pb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

var log = logging.Logger("peer")

//go:generate protoc --proto_path=$PWD:$PWD/../../.. --go_out=. --go_opt=Mpb/peer.proto=./pb pb/peer.proto

type ContactProto struct {
	host host.Host
	data ds.PeerIface

	emitters struct {
		evtSyncPeers       event.Emitter
		evtApplyAddContact event.Emitter
	}
}

func NewContactProto(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus) (*ContactProto, error) {
	var err error
	contactsvc := ContactProto{
		host: lhost,
		data: ds.Wrap(ids),
	}

	if contactsvc.emitters.evtSyncPeers, err = eventBus.Emitter(&gevent.EvtSyncPeers{}); err != nil {
		return nil, fmt.Errorf("set sync peers emitter error: %w", err)
	}

	if contactsvc.emitters.evtApplyAddContact, err = eventBus.Emitter(&gevent.EvtApplyAddContact{}); err != nil {
		return nil, fmt.Errorf("set apply add contact emitter error: %w", err)
	}

	sub, err := eventBus.Subscribe([]any{new(gevent.EvtHostBootComplete), new(gevent.EvtReceivePeerStream)})
	if err != nil {
		return nil, fmt.Errorf("subscribe boot complete error: %w", err)

	} else {
		go contactsvc.handleSubscribe(context.Background(), sub)
	}

	return &contactsvc, nil
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
					c.emitters.evtSyncPeers.Emit(gevent.EvtSyncPeers{
						ContactIDs: contactIDs,
					})
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

			default:
				log.Warnf("undefined event type: %T", ev)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (c *ContactProto) ApplyAddContact(ctx context.Context, peer0 *types.Peer, content string) error {

	if err := c.data.AddContact(ctx, &pb.ContactMsg{
		PeerId:   []byte(peer0.ID),
		Name:     peer0.Name,
		Avatar:   peer0.Avatar,
		AddTs:    time.Now().Unix(),
		AccessTs: time.Now().Unix(),
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

	if err := c.data.AddContact(ctx, &pb.ContactMsg{
		PeerId:   []byte(peer0.ID),
		Name:     peer0.Name,
		Avatar:   peer0.Avatar,
		AddTs:    time.Now().Unix(),
		AccessTs: time.Now().Unix(),
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
	if err := c.emitters.evtSyncPeers.Emit(gevent.EvtSyncPeers{
		ContactIDs: []peer.ID{peer0.ID},
	}); err != nil {
		return fmt.Errorf("emit sync peer error: %w", err)
	}

	return nil
}

func (c *ContactProto) GetContact(ctx context.Context, peerID peer.ID) (*ContactEntiy, error) {
	contact, err := c.data.GetContact(ctx, peerID)
	if err != nil {
		return nil, err
	}

	return &ContactEntiy{
		PeerID:   peer.ID(contact.PeerId),
		Name:     contact.Name,
		Avatar:   contact.Avatar,
		AddTs:    contact.AddTs,
		AccessTs: contact.AccessTs,
	}, nil
}

func (c *ContactProto) GetContactsByPeerIDs(ctx context.Context, peerIDs []peer.ID) ([]*pb.ContactMsg, error) {
	return c.data.GetContactsByIDs(ctx, peerIDs)
}

func (c *ContactProto) GetContacts(ctx context.Context) ([]*ContactEntiy, error) {
	contacts, err := c.data.GetContacts(ctx)
	if err != nil {
		return nil, err
	}

	var peers []*ContactEntiy
	for _, peeri := range contacts {
		peers = append(peers, &ContactEntiy{
			PeerID:   peer.ID(peeri.PeerId),
			Name:     peeri.Name,
			Avatar:   peeri.Avatar,
			AddTs:    peeri.AddTs,
			AccessTs: peeri.AccessTs,
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
