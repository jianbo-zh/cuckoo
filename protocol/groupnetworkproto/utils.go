package groupnetworkproto

import "github.com/libp2p/go-libp2p/core/peer"

func connKey(peerID1 peer.ID, peerID2 peer.ID) ConnKey {
	if peerID1.String() <= peerID2.String() {
		return peerID1.String() + "_" + peerID2.String()
	}
	return peerID2.String() + "_" + peerID1.String()
}
