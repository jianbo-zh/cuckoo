package msgsync

import (
	"github.com/jianbo-zh/dchat/datastore"
	logging "github.com/jianbo-zh/go-log"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)

// 消息同步相关协议

var log = logging.Logger("message")

const (
	ID = "/dchat/peer/msgsync/1.0.0"

	ServiceName = "peer.msgsync"
)

type PeerMsgSyncService struct {
	host      host.Host
	datastore datastore.PeerIface
}

func NewPeerMsgSyncService(h host.Host, ds datastore.PeerIface) *PeerMsgSyncService {
	kas := &PeerMsgSyncService{
		host:      h,
		datastore: ds,
	}
	h.SetStreamHandler(ID, kas.Handler)
	return kas
}

func (p *PeerMsgSyncService) Handler(s network.Stream) {

}
