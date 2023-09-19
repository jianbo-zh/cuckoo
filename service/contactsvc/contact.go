package contactsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/cuckoo/config"
	"github.com/jianbo-zh/dchat/internal/types"
	contactproto "github.com/jianbo-zh/dchat/service/contactsvc/protocol/contactproto"
	message "github.com/jianbo-zh/dchat/service/contactsvc/protocol/messageproto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var _ ContactServiceIface = (*ContactSvc)(nil)

type ContactSvc struct {
	msgProto     *message.PeerMessageProto
	contactProto *contactproto.ContactProto
}

func NewContactService(ctx context.Context, conf config.ContactServiceConfig, lhost host.Host, ids ipfsds.Batching, ebus event.Bus,
	rdiscvry *drouting.RoutingDiscovery) (*ContactSvc, error) {

	var err error

	contactsvc := &ContactSvc{}

	contactsvc.msgProto, err = message.NewMessageSvc(lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("message.NewMessageSvc error: %s", err.Error())
	}

	contactsvc.contactProto, err = contactproto.NewContactProto(lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("contactproto.NewPeerSvc error: %s", err.Error())
	}

	return contactsvc, nil
}

func (c *ContactSvc) Close() {}

func (c *ContactSvc) GetContactsByPeerIDs(ctx context.Context, peerIDs []peer.ID) ([]types.Contact, error) {

	var contacts []types.Contact
	result, err := c.contactProto.GetContactsByPeerIDs(ctx, peerIDs)
	if err != nil {
		return nil, err
	}

	for _, pi := range result {
		contacts = append(contacts, types.Contact{
			ID:     peer.ID(pi.PeerId),
			Name:   pi.Name,
			Avatar: pi.Avatar,
		})
	}

	return contacts, nil
}

func (c *ContactSvc) GetContacts(ctx context.Context) ([]types.Contact, error) {
	var contacts []types.Contact

	result, err := c.contactProto.GetContacts(ctx)
	if err != nil {
		return nil, err
	}

	for _, pi := range result {
		contacts = append(contacts, types.Contact{
			ID:     pi.PeerID,
			Name:   pi.Name,
			Avatar: pi.Avatar,
		})
	}

	return contacts, nil
}

func (c *ContactSvc) GetContactSessions(ctx context.Context) ([]types.ContactSession, error) {
	var contacts []types.ContactSession

	result, err := c.contactProto.GetContacts(ctx)
	if err != nil {
		return nil, err
	}

	for _, pi := range result {
		contacts = append(contacts, types.ContactSession{
			ID:     pi.PeerID,
			Name:   pi.Name,
			Avatar: pi.Avatar,
		})
	}

	return contacts, nil
}

func (c *ContactSvc) GetContact(ctx context.Context, peerID peer.ID) (*types.Contact, error) {

	contact, err := c.contactProto.GetContact(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("c.contactProto.GetContact error: %w", err)
	}

	return &types.Contact{
		ID:     contact.PeerID,
		Name:   contact.Name,
		Avatar: contact.Avatar,
	}, nil
}

func (c *ContactSvc) SetContactName(ctx context.Context, peerID peer.ID, name string) error {
	return c.contactProto.SetContactName(ctx, peerID, name)
}
func (c *ContactSvc) ApplyAddContact(ctx context.Context, peer0 *types.Peer, content string) error {
	return c.contactProto.ApplyAddContact(ctx, peer0, content)
}
func (c *ContactSvc) AgreeAddContact(ctx context.Context, peer0 *types.Peer) error {
	return c.contactProto.AgreeAddContact(ctx, peer0)
}

func (c *ContactSvc) DeleteContact(ctx context.Context, peerID peer.ID) error {
	return c.contactProto.DeleteContact(ctx, peerID)
}
