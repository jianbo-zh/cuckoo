package contactsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	contactproto "github.com/jianbo-zh/dchat/service/contactsvc/protocol/contactproto"
	message "github.com/jianbo-zh/dchat/service/contactsvc/protocol/messageproto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
)

var _ ContactServiceIface = (*ContactSvc)(nil)

type ContactSvc struct {
	msgProto     *message.PeerMessageProto
	contactProto *contactproto.ContactProto
}

func NewContactService(ctx context.Context, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus, accountGetter mytype.AccountGetter) (*ContactSvc, error) {

	var err error

	contactsvc := &ContactSvc{}

	contactsvc.contactProto, err = contactproto.NewContactProto(lhost, ids, ebus, accountGetter)
	if err != nil {
		return nil, fmt.Errorf("contactproto.NewPeerSvc error: %s", err.Error())
	}

	contactsvc.msgProto, err = message.NewMessageSvc(lhost, ids, ebus)
	if err != nil {
		return nil, fmt.Errorf("message.NewMessageSvc error: %s", err.Error())
	}

	return contactsvc, nil
}

func (c *ContactSvc) Close() {}

func (c *ContactSvc) GetContactsByPeerIDs(ctx context.Context, peerIDs []peer.ID) ([]mytype.Contact, error) {

	var contacts []mytype.Contact
	result, err := c.contactProto.GetContactsByPeerIDs(ctx, peerIDs)
	if err != nil {
		return nil, err
	}

	for _, pi := range result {
		contacts = append(contacts, mytype.Contact{
			ID:     peer.ID(pi.Id),
			Name:   pi.Name,
			Avatar: pi.Avatar,
		})
	}

	return contacts, nil
}

func (c *ContactSvc) GetContacts(ctx context.Context) ([]mytype.Contact, error) {
	var contacts []mytype.Contact

	result, err := c.contactProto.GetContacts(ctx)
	if err != nil {
		return nil, err
	}

	for _, pi := range result {
		contacts = append(contacts, mytype.Contact{
			ID:     pi.ID,
			Name:   pi.Name,
			Avatar: pi.Avatar,
		})
	}

	return contacts, nil
}

func (c *ContactSvc) GetContactSessions(ctx context.Context) ([]mytype.ContactSession, error) {
	var contacts []mytype.ContactSession

	result, err := c.contactProto.GetContacts(ctx)
	if err != nil {
		return nil, err
	}

	for _, pi := range result {
		contacts = append(contacts, mytype.ContactSession{
			ID:     pi.ID,
			Name:   pi.Name,
			Avatar: pi.Avatar,
		})
	}

	return contacts, nil
}

func (c *ContactSvc) GetContact(ctx context.Context, peerID peer.ID) (*mytype.Contact, error) {
	return c.contactProto.GetContact(ctx, peerID)
}

func (c *ContactSvc) SetContactName(ctx context.Context, peerID peer.ID, name string) error {
	return c.contactProto.SetContactName(ctx, peerID, name)
}
func (c *ContactSvc) ApplyAddContact(ctx context.Context, peer0 *mytype.Peer, content string) error {
	return c.contactProto.ApplyAddContact(ctx, peer0, content)
}
func (c *ContactSvc) AgreeAddContact(ctx context.Context, peer0 *mytype.Peer) error {
	return c.contactProto.AgreeAddContact(ctx, peer0)
}

func (c *ContactSvc) DeleteContact(ctx context.Context, peerID peer.ID) error {
	return c.contactProto.DeleteContact(ctx, peerID)
}
