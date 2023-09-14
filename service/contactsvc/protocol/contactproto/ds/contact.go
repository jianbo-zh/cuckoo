package ds

import (
	"context"
	"fmt"
	"strings"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/internal/datastore"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocol/contactproto/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

var _ PeerIface = (*PeerDS)(nil)

var contactDsKey = &datastore.ContactDsKey{}

type PeerDS struct {
	ipfsds.Batching
}

func Wrap(b ipfsds.Batching) *PeerDS {
	return &PeerDS{Batching: b}
}

func (p *PeerDS) AddContact(ctx context.Context, info *pb.ContactMsg) error {
	value, err := proto.Marshal(info)
	if err != nil {
		return err
	}

	return p.Put(ctx, contactDsKey.ContactKey(peer.ID(info.PeerId)), value)
}

func (p *PeerDS) GetContact(ctx context.Context, peerID peer.ID) (*pb.ContactMsg, error) {

	value, err := p.Get(ctx, contactDsKey.ContactKey(peer.ID(peerID)))
	if err != nil {
		return nil, err
	}

	var info pb.ContactMsg
	if err := proto.Unmarshal(value, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (p *PeerDS) GetContactsByIDs(ctx context.Context, peerIDs []peer.ID) ([]*pb.ContactMsg, error) {
	var contacts []*pb.ContactMsg
	for _, peerID := range peerIDs {
		value, err := p.Get(ctx, contactDsKey.ContactKey(peerID))
		if err != nil {
			return nil, err
		}

		var contact pb.ContactMsg
		if err := proto.Unmarshal(value, &contact); err != nil {
			return nil, err
		}
		contacts = append(contacts, &contact)
	}

	return contacts, nil
}

func (p *PeerDS) GetContacts(ctx context.Context) ([]*pb.ContactMsg, error) {
	results, err := p.Query(ctx, query.Query{
		Prefix: contactDsKey.Prefix(),
	})
	if err != nil {
		return nil, err
	}

	var peeris []*pb.ContactMsg
	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		var peeri pb.ContactMsg
		if err := proto.Unmarshal(result.Entry.Value, &peeri); err != nil {
			return nil, err
		}

		peeris = append(peeris, &peeri)
	}

	return peeris, nil
}

func (p *PeerDS) GetContactIDs(ctx context.Context) ([]peer.ID, error) {
	results, err := p.Query(ctx, query.Query{
		Prefix:   contactDsKey.Prefix(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	var peerIDs []peer.ID
	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		peerID, err := peer.Decode(strings.TrimLeft(result.Entry.Key, contactDsKey.Prefix()))
		if err != nil {
			return nil, err
		}

		peerIDs = append(peerIDs, peerID)
	}

	return peerIDs, nil
}

func (p *PeerDS) UpdateContact(ctx context.Context, info *pb.ContactMsg) error {
	value, err := proto.Marshal(info)
	if err != nil {
		return err
	}

	return p.Put(ctx, contactDsKey.ContactKey(peer.ID(info.PeerId)), value)
}

func (p *PeerDS) DeleteContact(ctx context.Context, peerID peer.ID) error {
	// delete message
	results, err := p.Query(ctx, query.Query{
		Prefix:   contactDsKey.MsgPrefix(peerID),
		KeysOnly: true,
	})
	if err != nil {
		return fmt.Errorf("ds query error: %w", err)
	}

	for result := range results.Next() {
		if result.Error != nil {
			return fmt.Errorf("ds result next error: %w", err)
		}

		if err = p.Delete(ctx, ipfsds.NewKey(result.Key)); err != nil {
			return fmt.Errorf("ds delete key error: %w", err)
		}
	}

	// delete contact
	return p.Delete(ctx, contactDsKey.ContactKey(peerID))
}
