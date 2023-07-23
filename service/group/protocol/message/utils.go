package message

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
)

func msgID(lamptime uint64, peerID peer.ID) string {
	return fmt.Sprintf("%019d_%s", lamptime, peerID.String())
}
