package contactds

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/internal/datastore"
	pb "github.com/jianbo-zh/dchat/service/contactsvc/protobuf/pb/contactpb"
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

func (p *PeerDS) AddContact(ctx context.Context, info *pb.Contact) error {
	value, err := proto.Marshal(info)
	if err != nil {
		return err
	}

	return p.Put(ctx, contactDsKey.DetailKey(peer.ID(info.Id)), value)
}

func (p *PeerDS) GetContact(ctx context.Context, peerID peer.ID) (*pb.Contact, error) {

	value, err := p.Get(ctx, contactDsKey.DetailKey(peer.ID(peerID)))
	if err != nil {
		return nil, err
	}

	var contact pb.Contact
	if err := proto.Unmarshal(value, &contact); err != nil {
		return nil, err
	}

	return &contact, nil
}

func (p *PeerDS) GetContactsByIDs(ctx context.Context, peerIDs []peer.ID) ([]*pb.Contact, error) {
	var contacts []*pb.Contact
	for _, peerID := range peerIDs {
		value, err := p.Get(ctx, contactDsKey.DetailKey(peerID))
		if err != nil {
			return nil, fmt.Errorf("get contact detail error: %w", err)
		}

		var contact pb.Contact
		if err := proto.Unmarshal(value, &contact); err != nil {
			return nil, fmt.Errorf("proto unmarshal error: %w", err)
		}
		contacts = append(contacts, &contact)
	}

	return contacts, nil
}

func (p *PeerDS) GetContacts(ctx context.Context) ([]*pb.Contact, error) {
	results, err := p.Query(ctx, query.Query{
		Prefix:   contactDsKey.ListPrefix(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	var contacts []*pb.Contact
	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}

		peerID, err := peer.Decode(strings.TrimPrefix(result.Key, contactDsKey.ListPrefix()))
		if err != nil {
			return nil, fmt.Errorf("peer decode error: %w", err)
		}

		value, err := p.Get(ctx, contactDsKey.DetailKey(peerID))
		if err != nil {
			return nil, fmt.Errorf("get contact detail error: %w", err)
		}

		var contact pb.Contact
		if err = proto.Unmarshal(value, &contact); err != nil {
			return nil, fmt.Errorf("proto unmarshal error: %w", err)
		}

		contacts = append(contacts, &contact)
	}

	return contacts, nil
}

func (p *PeerDS) UpdateContact(ctx context.Context, info *pb.Contact) error {
	value, err := proto.Marshal(info)
	if err != nil {
		return err
	}

	return p.Put(ctx, contactDsKey.DetailKey(peer.ID(info.Id)), value)
}

// DeleteContact 删除联系人
func (p *PeerDS) DeleteContact(ctx context.Context, peerID peer.ID) error {

	if err := p.Delete(ctx, contactDsKey.ListKey(peerID)); err != nil {
		return fmt.Errorf("delete formal error: %w", err)
	}

	results, err := p.Query(ctx, query.Query{
		Prefix:   contactDsKey.ContactPrefix(peerID),
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

	return nil
}

// SetApply 设置新申请
func (p *PeerDS) SetApply(ctx context.Context, peerID peer.ID) error {
	err := p.Put(ctx, contactDsKey.ApplyKey(peerID), []byte(strconv.FormatInt(time.Now().Unix(), 10)))
	if err != nil {
		return fmt.Errorf("set apply error: %w", err)
	}

	return nil
}

// DeleteApply 删除申请
func (p *PeerDS) DeleteApply(ctx context.Context, peerID peer.ID) error {
	if err := p.Delete(ctx, contactDsKey.ApplyKey(peerID)); err != nil {
		return fmt.Errorf("delete apply error: %w", err)
	}

	return nil
}

// GetApplyIDs 获取会话ID列表
func (p *PeerDS) GetApplyIDs(ctx context.Context) ([]peer.ID, error) {
	results, err := p.Query(ctx, query.Query{
		Prefix:   contactDsKey.ApplyPrefix(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var peerIDs []peer.ID
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("ds result next error: %w", err)
		}

		peerID, err := peer.Decode(strings.TrimPrefix(result.Key, contactDsKey.ApplyPrefix()))
		if err != nil {
			return nil, fmt.Errorf("peer decode error: %w", err)
		}

		peerIDs = append(peerIDs, peerID)
	}

	return peerIDs, nil
}

// SetSession 设置新联系人
func (p *PeerDS) SetFormal(ctx context.Context, peerID peer.ID) error {
	err := p.Put(ctx, contactDsKey.ListKey(peerID), []byte(strconv.FormatInt(time.Now().Unix(), 10)))
	if err != nil {
		return fmt.Errorf("set formal error: %w", err)
	}

	return nil
}

// GetSessionIDs 获取联系人ID列表
func (p *PeerDS) GetFormalIDs(ctx context.Context) ([]peer.ID, error) {
	results, err := p.Query(ctx, query.Query{
		Prefix:   contactDsKey.ListPrefix(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var peerIDs []peer.ID
	for result := range results.Next() {
		if result.Error != nil {
			return nil, fmt.Errorf("ds result next error: %w", err)
		}

		peerID, err := peer.Decode(strings.TrimPrefix(result.Key, contactDsKey.ListPrefix()))
		if err != nil {
			return nil, fmt.Errorf("peer decode error: %w", err)
		}

		peerIDs = append(peerIDs, peerID)
	}

	return peerIDs, nil
}
