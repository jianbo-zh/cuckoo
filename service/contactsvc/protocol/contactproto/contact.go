package peer

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	gevent "github.com/jianbo-zh/dchat/event"
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
	host      host.Host
	data      ds.PeerIface
	avatarDir string

	emitters struct {
		evtSyncPeers event.Emitter
	}
}

func NewContactProto(lhost host.Host, ids ipfsds.Batching, eventBus event.Bus, avatarDir string) (*ContactProto, error) {
	var err error
	contactsvc := ContactProto{
		host:      lhost,
		data:      ds.Wrap(ids),
		avatarDir: avatarDir,
	}

	if contactsvc.emitters.evtSyncPeers, err = eventBus.Emitter(&gevent.EvtSyncPeers{}); err != nil {
		return nil, fmt.Errorf("set sync peers emitter error: %v", err)
	}

	sub, err := eventBus.Subscribe([]any{new(gevent.EvtHostBootComplete)})
	if err != nil {
		return nil, fmt.Errorf("subscribe boot complete error: %v", err)

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
					return
				}

				if contactIDs, err := c.data.GetContactIDs(ctx); err != nil {
					log.Warnf("get peer ids error: %v", err)
					return

				} else if len(contactIDs) > 0 {
					c.emitters.evtSyncPeers.Emit(gevent.EvtSyncPeers{
						PeerIDs: contactIDs,
					})
				}

			default:
				log.Warnf("undefined event type: %T", ev)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (c *ContactProto) AddContact(ctx context.Context, peerID peer.ID, name string, avatar string) error {

	if err := c.data.AddContact(ctx, &pb.ContactMsg{
		PeerId:   []byte(peerID),
		Name:     name,
		Avatar:   avatar,
		AddTs:    time.Now().Unix(),
		AccessTs: time.Now().Unix(),
	}); err != nil {
		return fmt.Errorf("data.AddContact error: %w", err)
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

func (c *ContactProto) DeleteContact(ctx context.Context, peerID peer.ID) error {
	return c.data.DeleteContact(ctx, peerID)
}
