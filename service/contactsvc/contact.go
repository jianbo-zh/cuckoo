package contactsvc

import (
	"context"
	"fmt"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocols/contactmsgproto"
	"github.com/jianbo-zh/dchat/service/contactsvc/protocols/contactproto"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
)

var log = logging.Logger("cuckoo/contactsvc")

var _ ContactServiceIface = (*ContactSvc)(nil)

type ContactSvc struct {
	host myhost.Host

	accountGetter mytype.AccountGetter

	msgProto     *contactmsgproto.PeerMessageProto
	contactProto *contactproto.ContactProto

	emitters struct {
		evtLogSessionAttachment      event.Emitter
		evtPushDepositContactMessage event.Emitter
		evtClearSessionResources     event.Emitter
		evtClearSessionFiles         event.Emitter
		evtClearSession              event.Emitter
		evtDeleteSession             event.Emitter
	}
}

func NewContactService(ctx context.Context, lhost myhost.Host, ids ipfsds.Batching, ebus event.Bus, accountGetter mytype.AccountGetter) (*ContactSvc, error) {

	var err error

	svc := &ContactSvc{
		host:          lhost,
		accountGetter: accountGetter,
	}

	if svc.contactProto, err = contactproto.NewContactProto(lhost, ids, ebus, accountGetter); err != nil {
		return nil, fmt.Errorf("contactproto.NewContactProto error: %w", err)
	}

	if svc.msgProto, err = contactmsgproto.NewMessageSvc(lhost, ids, ebus); err != nil {
		return nil, fmt.Errorf("contactmsgproto.NewMessageSvc error: %w", err)
	}

	// 触发器：关联会话附件
	if svc.emitters.evtLogSessionAttachment, err = ebus.Emitter(&myevent.EvtLogSessionAttachment{}); err != nil {
		return nil, fmt.Errorf("set send resource request emitter error: %w", err)
	}

	// 触发器：发送离线消息
	if svc.emitters.evtPushDepositContactMessage, err = ebus.Emitter(&myevent.EvtPushDepositContactMessage{}); err != nil {
		return nil, fmt.Errorf("set pull deposit msg emitter error: %v", err)
	}

	// 触发器：删除会话资源
	if svc.emitters.evtClearSessionResources, err = ebus.Emitter(&myevent.EvtClearSessionResources{}); err != nil {
		return nil, fmt.Errorf("set send resource request emitter error: %w", err)
	}

	// 触发器：删除会话文件
	if svc.emitters.evtClearSessionFiles, err = ebus.Emitter(&myevent.EvtClearSessionFiles{}); err != nil {
		return nil, fmt.Errorf("set send resource request emitter error: %w", err)
	}

	// 触发器：清空会话
	if svc.emitters.evtClearSession, err = ebus.Emitter(&myevent.EvtClearSession{}); err != nil {
		return nil, fmt.Errorf("set clear session emitter error: %w", err)
	}

	// 触发器：删除会话
	if svc.emitters.evtDeleteSession, err = ebus.Emitter(&myevent.EvtDeleteSession{}); err != nil {
		return nil, fmt.Errorf("set delete session emitter error: %w", err)
	}

	return svc, nil
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

// DeleteContact 删除联系人（附件、消息、会话）
func (c *ContactSvc) DeleteContact(ctx context.Context, contactID peer.ID) error {

	sessionID := mytype.ContactSessionID(contactID)

	// 删除联系人会话
	if err := c.contactProto.DeleteContact(ctx, contactID); err != nil {
		return fmt.Errorf("proto.DeleteContact error: %w", err)
	}

	// 删除会话
	resultCh := make(chan error, 1)
	if err := c.emitters.evtDeleteSession.Emit(myevent.EvtDeleteSession{
		SessionID: sessionID.String(),
		Result:    resultCh,
	}); err != nil {
		return fmt.Errorf("emit delete session error: %w", err)
	}
	if err := <-resultCh; err != nil {
		return fmt.Errorf("delete session error: %w", err)
	}

	// 删除会话资源
	resultCh1 := make(chan error, 1)
	if err := c.emitters.evtClearSessionResources.Emit(myevent.EvtClearSessionResources{
		SessionID: sessionID.String(),
		Result:    resultCh1,
	}); err != nil {
		return fmt.Errorf("emit clear session resources error: %w", err)
	}
	if err := <-resultCh1; err != nil {
		return fmt.Errorf("clear session resources error: %w", err)
	}

	// 删除会话文件
	resultCh2 := make(chan error, 1)
	if err := c.emitters.evtClearSessionFiles.Emit(myevent.EvtClearSessionFiles{
		SessionID: sessionID.String(),
		Result:    resultCh2,
	}); err != nil {
		return fmt.Errorf("emit clear session files error: %w", err)
	}
	if err := <-resultCh2; err != nil {
		return fmt.Errorf("clear session files error: %w", err)
	}

	return nil
}
