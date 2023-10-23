package myhost

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"

	logging "github.com/jianbo-zh/go-log"
	ma "github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("myhost")

type Host interface {
	host.Host

	OnlineState(peerID peer.ID, onlineDuration time.Duration) mytype.OnlineState

	// peer在线状态统计
	OnlineStats(peerIDs []peer.ID, onlineDuration time.Duration) map[peer.ID]mytype.OnlineState
}

type MyHost struct {
	basehost host.Host

	statsMutex sync.RWMutex
	onlineMap  map[peer.ID]time.Time
	offlineMap map[peer.ID]time.Time

	emitters struct {
		evtPeerStateChanged event.Emitter
	}
}

func NewHost(h host.Host, ebus event.Bus) (Host, error) {
	var err error

	lhost := MyHost{
		basehost:   h,
		onlineMap:  make(map[peer.ID]time.Time),
		offlineMap: make(map[peer.ID]time.Time),
	}

	if lhost.emitters.evtPeerStateChanged, err = ebus.Emitter(&myevent.EvtPeerStateChanged{}); err != nil {
		return nil, fmt.Errorf("ebus.Emitter error: %w", err)
	}

	return &lhost, nil

}

func (h *MyHost) ID() peer.ID {
	return h.basehost.ID()
}

func (h *MyHost) Peerstore() peerstore.Peerstore {
	return h.basehost.Peerstore()
}

func (h *MyHost) Addrs() []ma.Multiaddr {
	return h.basehost.Addrs()
}

func (h *MyHost) Network() network.Network {
	return h.basehost.Network()
}

func (h *MyHost) Mux() protocol.Switch {
	return h.basehost.Mux()
}

func (h *MyHost) Connect(ctx context.Context, pi peer.AddrInfo) error {
	return h.basehost.Connect(ctx, pi)
}

func (h *MyHost) SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
	h.basehost.SetStreamHandler(pid, func(stream network.Stream) {
		h.online(stream.Conn().RemotePeer())
		handler(stream)
	})
}

func (h *MyHost) SetStreamHandlerMatch(pid protocol.ID, match func(protocol.ID) bool, handler network.StreamHandler) {
	h.basehost.SetStreamHandlerMatch(pid, match, func(stream network.Stream) {
		h.online(stream.Conn().RemotePeer())
		handler(stream)
	})
}

func (h *MyHost) RemoveStreamHandler(pid protocol.ID) {
	h.basehost.RemoveStreamHandler(pid)
}

func (h *MyHost) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (stream network.Stream, err error) {
	stream, err = h.basehost.NewStream(ctx, p, pids...)
	if err != nil {
		h.offline(p)
	} else {
		h.online(p)
	}
	return
}

func (h *MyHost) ConnManager() connmgr.ConnManager {
	return h.basehost.ConnManager()
}

func (h *MyHost) EventBus() event.Bus {
	return h.basehost.EventBus()
}

func (h *MyHost) Close() error {
	return h.basehost.Close()
}
