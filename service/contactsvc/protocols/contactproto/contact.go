package contactproto

import (
	"context"
	"fmt"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/myhost"
	"github.com/jianbo-zh/dchat/internal/myprotocol"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/jianbo-zh/dchat/service/common/datastore/ds/sessionds"
	ds "github.com/jianbo-zh/dchat/service/contactsvc/datastore/ds/contactds"
	pb "github.com/jianbo-zh/dchat/service/contactsvc/protobuf/pb/contactpb"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
)

var log = logging.Logger("cuckoo/contactproto")

var StreamTimeout = 1 * time.Minute

const (
	SYNC_ID  = myprotocol.ContactID_v100
	CHECK_ID = myprotocol.ContactCheckID_v100

	ServiceName = "peer.contact"
)

type ContactProto struct {
	host        myhost.Host
	data        ds.PeerIface
	sessionData sessionds.SessionIface

	accountGetter mytype.AccountGetter

	emitters struct {
		evtSyncPeerMessage event.Emitter
		evtApplyAddContact event.Emitter
		evtSyncResource    event.Emitter
		evtContactAdded    event.Emitter
	}
}

func NewContactProto(lhost myhost.Host, ids ipfsds.Batching, eventBus event.Bus, accountGetter mytype.AccountGetter) (*ContactProto, error) {
	var err error
	contactsvc := ContactProto{
		host:          lhost,
		data:          ds.Wrap(ids),
		sessionData:   sessionds.SessionWrap(ids),
		accountGetter: accountGetter,
	}

	contactsvc.host.SetStreamHandler(SYNC_ID, contactsvc.syncHandler)
	contactsvc.host.SetStreamHandler(CHECK_ID, contactsvc.checkApplyHandler)

	if contactsvc.emitters.evtSyncPeerMessage, err = eventBus.Emitter(&myevent.EvtSyncContactMessage{}); err != nil {
		return nil, fmt.Errorf("set sync peers emitter error: %w", err)
	}

	if contactsvc.emitters.evtApplyAddContact, err = eventBus.Emitter(&myevent.EvtApplyAddContact{}); err != nil {
		return nil, fmt.Errorf("set apply add contact emitter error: %w", err)
	}

	if contactsvc.emitters.evtSyncResource, err = eventBus.Emitter(&myevent.EvtSyncResource{}); err != nil {
		return nil, fmt.Errorf("set check avatar emitter error: %w", err)
	}

	if contactsvc.emitters.evtContactAdded, err = eventBus.Emitter(&myevent.EvtContactAdded{}); err != nil {
		return nil, fmt.Errorf("set check avatar emitter error: %w", err)
	}

	sub, err := eventBus.Subscribe([]any{new(myevent.EvtHostBootComplete), new(myevent.EvtAccountBaseChange)})
	if err != nil {
		return nil, fmt.Errorf("subscribe boot complete error: %w", err)

	} else {
		go contactsvc.handleSubscribe(context.Background(), sub)
	}

	return &contactsvc, nil
}

func (c *ContactProto) syncHandler(stream network.Stream) {
	log.Infoln("sync handler")

	remotePeerID := stream.Conn().RemotePeer()
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	defer rd.Close()

	// 读取对方数据
	var msg pb.ContactPeer
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	log.Debugln("sync handler receive peer msg: ", msg.String())

	// 更新本地
	ctx := context.Background()
	contact, err := c.data.GetContact(ctx, remotePeerID)
	if err != nil {
		log.Errorf("get contact error: %w", err)
		stream.Reset()
		return
	}

	// 触发检查是否需要同步头像
	if err = c.emitters.evtSyncResource.Emit(myevent.EvtSyncResource{
		ResourceID: msg.Avatar,
		PeerIDs:    []peer.ID{remotePeerID},
	}); err != nil {
		log.Errorf("emit check avatar evt error: %w", err)
	}

	contact.Name = msg.Name
	contact.Avatar = msg.Avatar
	contact.DepositAddress = msg.DepositAddress

	if contact.State == pb.ContactState_Apply {
		log.Debugln("set contact state apply -> normal")
		contact.State = pb.ContactState_Normal

		if err = c.data.SetFormal(ctx, remotePeerID); err != nil {
			log.Errorf("data.SetFormal error: %w", err)
			stream.Reset()
			return
		}

		sessionID := mytype.ContactSessionID(remotePeerID)
		if err = c.sessionData.SetSessionID(ctx, sessionID.String()); err != nil {
			log.Errorf("svc set session error: %w", err)
			stream.Reset()
			return
		}

		// 触发新联系人
		c.emitters.evtContactAdded.Emit(myevent.EvtContactAdded{
			ID:             peer.ID(contact.Id),
			Name:           contact.Name,
			Avatar:         contact.Avatar,
			DepositAddress: peer.ID(contact.DepositAddress),
		})
	}

	log.Debugln("update contact %s", contact.String())
	if err := c.data.UpdateContact(ctx, contact); err != nil {
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

	accountMsg := pb.Contact{
		Id:             []byte(account.ID),
		Name:           account.Name,
		Avatar:         account.Avatar,
		DepositAddress: []byte(account.DepositAddress),
	}

	log.Debugln("sync handler send peer: ", accountMsg.String())

	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&accountMsg); err != nil {
		log.Errorf("pbio write msg error: %w", err)
		stream.Reset()
		return
	}
}

func (c *ContactProto) checkApplyHandler(stream network.Stream) {

	remotePeerID := stream.Conn().RemotePeer()

	defer stream.Close()

	// 读取对方数据
	var msg pb.ContactPeer
	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	if err := rd.ReadMsg(&msg); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	// 更新本地
	ctx := context.Background()
	contact, err := c.data.GetContact(ctx, remotePeerID)
	if err != nil {
		log.Errorf("get contact error: %w", err)
		stream.Reset()
		return
	}

	var peerState pb.ContactPeerState
	if contact.State == pb.ContactState_Normal {
		// 已经是好友了，则更新信息
		contact.Name = msg.Name
		contact.Avatar = msg.Avatar
		contact.DepositAddress = msg.DepositAddress
		if err := c.data.UpdateContact(ctx, contact); err != nil {
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

		peerState = pb.ContactPeerState{
			Id:             []byte(account.ID),
			Name:           account.Name,
			Avatar:         account.Avatar,
			DepositAddress: []byte(account.DepositAddress),
			State:          contact.State,
		}

	} else {
		peerState = pb.ContactPeerState{
			State: contact.State,
		}
	}

	wt := pbio.NewDelimitedWriter(stream)
	if err := wt.WriteMsg(&peerState); err != nil {
		log.Errorf("pbio write msg error: %w", err)
		stream.Reset()
		return
	}

}

func (c *ContactProto) syncContact(contactID peer.ID, accountPeer mytype.AccountPeer, isBootSync bool) {

	log.Debugln("sync contact: ", contactID.String(), accountPeer, isBootSync)

	ctx := context.Background()
	stream, err := c.host.NewStream(network.WithUseTransient(ctx, ""), contactID, SYNC_ID)
	if err != nil {
		log.Errorf("host new stream error: %v", err)
		return
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	wt := pbio.NewDelimitedWriter(stream)

	// 发送数据给对方
	sendmsg := pb.ContactPeer{
		Id:             []byte(accountPeer.ID),
		Name:           accountPeer.Name,
		Avatar:         accountPeer.Avatar,
		DepositAddress: []byte(accountPeer.DepositAddress),
	}
	log.Debugln("sync contact send peer: ", sendmsg.String())

	if err = wt.WriteMsg(&sendmsg); err != nil {
		log.Errorf("pbio write msg error: %w", err)
		stream.Reset()
		return
	}

	// 读取对方数据
	var recvmsg pb.ContactPeer
	if err := rd.ReadMsg(&recvmsg); err != nil {
		log.Errorf("pbio read msg error: %v", err)
		stream.Reset()
		return
	}

	log.Debugln("sync contact recv peer: ", peer.ID(recvmsg.Id).String())

	// 更新本地数据
	if err = c.data.UpdateContact(ctx, &pb.Contact{
		Id:             recvmsg.Id,
		Name:           recvmsg.Name,
		Avatar:         recvmsg.Avatar,
		DepositAddress: recvmsg.DepositAddress,
	}); err != nil {
		log.Errorf("data update contact error: %v", err)
		stream.Reset()
		return
	}

	if isBootSync {
		// 启动时同步，还要触发同步消息事件
		if err = c.emitters.evtSyncPeerMessage.Emit(myevent.EvtSyncContactMessage{
			ContactID: contactID,
		}); err != nil {
			log.Errorf("emit sync peer message error: %w", err)

		} else {
			log.Debugln("emit EvtSyncContactMessage ", contactID.String())
		}
	}

	// 触发检查是否需要同步头像
	if err = c.emitters.evtSyncResource.Emit(myevent.EvtSyncResource{
		ResourceID: recvmsg.Avatar,
		PeerIDs:    []peer.ID{contactID},
	}); err != nil {
		log.Errorf("emit check avatar evt error: %w", err)
	}
}

func (c *ContactProto) goCheckApply(peerID peer.ID, accountPeer mytype.AccountPeer) {

	log.Debugln("check apply: ", peerID.String())

	ctx := context.Background()
	stream, err := c.host.NewStream(network.WithUseTransient(ctx, ""), peerID, CHECK_ID)
	if err != nil {
		log.Errorf("host new stream error: %w", err)
		return
	}
	defer stream.Close()

	rd := pbio.NewDelimitedReader(stream, mytype.PbioReaderMaxSizeNormal)
	wt := pbio.NewDelimitedWriter(stream)

	// 发送数据给对方
	sendmsg := pb.ContactPeer{
		Id:             []byte(accountPeer.ID),
		Name:           accountPeer.Name,
		Avatar:         accountPeer.Avatar,
		DepositAddress: []byte(accountPeer.DepositAddress),
	}

	log.Debugln("check apply send peer: ", sendmsg.String())

	if err = wt.WriteMsg(&sendmsg); err != nil {
		log.Errorf("pbio write msg error: %w", err)
		stream.Reset()
		return
	}

	// 读取对方数据
	var recvmsg pb.ContactPeerState
	if err = rd.ReadMsg(&recvmsg); err != nil {
		log.Errorf("pbio read msg error: %w", err)
		stream.Reset()
		return
	}

	log.Debugln("check apply recv peer: ", recvmsg.String())

	if recvmsg.State == pb.ContactState_Normal {
		log.Debugln("peer is contact, set state normal")

		// 对方已经同意了，则更新本地数据
		if err = c.data.UpdateContact(ctx, &pb.Contact{
			Id:             recvmsg.Id,
			Name:           recvmsg.Name,
			Avatar:         recvmsg.Avatar,
			DepositAddress: recvmsg.DepositAddress,
			State:          pb.ContactState_Normal,
			CreateTime:     time.Now().Unix(),
		}); err != nil {
			log.Errorf("data update contact error: %v", err)
			stream.Reset()
			return
		}

		// 设置正常好友
		if err = c.data.SetFormal(ctx, peerID); err != nil {
			log.Errorf("data.SetFormal error: %w", err)
			stream.Reset()
			return
		}

		// 更新会话
		sessionID := mytype.ContactSessionID(peerID)
		if err = c.sessionData.SetSessionID(ctx, sessionID.String()); err != nil {
			log.Errorf("svc set session error: %w", err)
			stream.Reset()
			return
		}

		// 删除申请记录缓存
		if err = c.data.DeleteApply(ctx, peerID); err != nil {
			log.Errorf("data.DeleteApply error: %v", err)
			stream.Reset()
			return
		}

		// 触发新联系人
		c.emitters.evtContactAdded.Emit(myevent.EvtContactAdded{
			ID:             peer.ID(recvmsg.Id),
			Name:           recvmsg.Name,
			Avatar:         recvmsg.Avatar,
			DepositAddress: peer.ID(recvmsg.DepositAddress),
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
			case myevent.EvtHostBootComplete:

				if !ev.IsSucc {
					log.Warnf("host boot complete but not succ")
					continue
				}

				account, err := c.accountGetter.GetAccount(ctx)
				if err != nil {
					log.Errorf("get account error: %w", err)
					continue
				}

				// 检查所有申请是否通过
				if peerIDs, err := c.data.GetApplyIDs(ctx); err != nil {
					log.Errorf("data.GetApplyIDs error: %w", err)

				} else if len(peerIDs) > 0 {
					// 超过7天就算过期了
					expiredts := time.Now().Unix() - (7 * 24 * 60 * 60)
					for _, peerID := range peerIDs {
						log.Debugln("find apply contact ", peerID.String())

						contact, err := c.data.GetContact(ctx, peerID)
						if err != nil {
							log.Errorf("data.GetContact error: %v", err)
							continue
						}

						if contact.CreateTime < expiredts || contact.State == pb.ContactState_Normal { // 申请过期或者好友关系正常
							log.Debugln("apply contact expired, should delete")
							if err = c.data.DeleteApply(ctx, peerID); err != nil {
								log.Errorf("data.DeleteApply error: %w", err)
							}
						} else {
							go c.goCheckApply(peerID, mytype.AccountPeer{
								ID:             account.ID,
								Name:           account.Name,
								Avatar:         account.Avatar,
								DepositAddress: account.DepositAddress,
							})
						}
					}
				} else {
					log.Debugln("no apply contacts")
				}

				// 交换好友最新信息
				if contactIDs, err := c.data.GetFormalIDs(ctx); err != nil {
					log.Warnf("get peer ids error: %v", err)
					continue

				} else if len(contactIDs) > 0 {
					for _, contactID := range contactIDs {
						go c.syncContact(contactID, mytype.AccountPeer{
							ID:             account.ID,
							Name:           account.Name,
							Avatar:         account.Avatar,
							DepositAddress: account.DepositAddress,
						}, true)
					}
				}

			case myevent.EvtAccountBaseChange:
				log.Infoln("EvtAccountBaseChange")

				if contactIDs, err := c.data.GetFormalIDs(ctx); err != nil {
					log.Warnf("get peer ids error: %v", err)
					continue

				} else if len(contactIDs) > 0 {
					for _, contactID := range contactIDs {
						go c.syncContact(contactID, ev.AccountPeer, false)
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

func (c *ContactProto) ApplyAddContact(ctx context.Context, peer0 *mytype.Peer, content string) error {

	if err := c.data.AddContact(ctx, &pb.Contact{
		Id:         []byte(peer0.ID),
		Name:       peer0.Name,
		Avatar:     peer0.Avatar,
		State:      pb.ContactState_Apply,
		CreateTime: time.Now().Unix(),
	}); err != nil {
		return fmt.Errorf("data.AddContact error: %w", err)
	}

	// 本地保存申请记录
	if err := c.data.SetApply(ctx, peer0.ID); err != nil {
		return fmt.Errorf("data.SetApply error: %w", err)
	}

	resultCh := make(chan error, 1)
	if err := c.emitters.evtApplyAddContact.Emit(myevent.EvtApplyAddContact{
		PeerID:      peer0.ID,
		DepositAddr: peer0.DepositAddress,
		Content:     content,
		Result:      resultCh,
	}); err != nil {
		return fmt.Errorf("emit apply add contact error: %w", err)
	}

	if err := <-resultCh; err != nil {
		return fmt.Errorf("send apply add contact error: %w", err)
	}

	return nil
}

func (c *ContactProto) AgreeAddContact(ctx context.Context, peer0 *mytype.Peer) error {

	if err := c.data.AddContact(ctx, &pb.Contact{
		Id:         []byte(peer0.ID),
		Name:       peer0.Name,
		Avatar:     peer0.Avatar,
		State:      pb.ContactState_Normal,
		CreateTime: time.Now().Unix(),
		AccessTime: time.Now().Unix(),
	}); err != nil {
		return fmt.Errorf("data.AddContact error: %w", err)
	}

	// 设置正式联系人
	if err := c.data.SetFormal(ctx, peer0.ID); err != nil {
		return fmt.Errorf("data.SetFormal error: %w", err)
	}

	// 设置会话
	sessionID := mytype.ContactSessionID(peer0.ID)
	if err := c.sessionData.SetSessionID(ctx, sessionID.String()); err != nil {
		return fmt.Errorf("svc set session error: %w", err)
	}

	// 启动同步，连接对方
	account, err := c.accountGetter.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("get account error: %w", err)
	}

	go c.syncContact(peer0.ID, mytype.AccountPeer{
		ID:             account.ID,
		Name:           account.Name,
		Avatar:         account.Avatar,
		DepositAddress: account.DepositAddress,
	}, false)

	return nil
}

func (c *ContactProto) GetContact(ctx context.Context, peerID peer.ID) (*mytype.Contact, error) {
	contact, err := c.data.GetContact(ctx, peerID)
	if err != nil {
		return nil, err
	}

	return &mytype.Contact{
		ID:             peer.ID(contact.Id),
		Name:           contact.Name,
		Avatar:         contact.Avatar,
		DepositAddress: peer.ID(contact.DepositAddress),
	}, nil
}

func (c *ContactProto) GetContactsByPeerIDs(ctx context.Context, peerIDs []peer.ID) ([]*pb.Contact, error) {
	return c.data.GetContactsByIDs(ctx, peerIDs)
}

func (c *ContactProto) GetContacts(ctx context.Context) ([]*mytype.Contact, error) {
	contacts, err := c.data.GetContacts(ctx)
	if err != nil {
		return nil, err
	}

	var peers []*mytype.Contact
	for _, peeri := range contacts {
		peers = append(peers, &mytype.Contact{
			ID:             peer.ID(peeri.Id),
			Name:           peeri.Name,
			Avatar:         peeri.Avatar,
			DepositAddress: peer.ID(peeri.DepositAddress),
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
