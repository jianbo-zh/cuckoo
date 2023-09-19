package network

import (
	"crypto/sha256"
	"math/big"
	"sort"

	"github.com/libp2p/go-libp2p/core/peer"
)

type ConnAlgo struct {
	peerIndex  int
	orderQueue []peer.ID
}

func NewConnAlgo(hostID peer.ID, otherPeerIDs []peer.ID) *ConnAlgo {
	unimap := make(map[uint64]peer.ID, len(otherPeerIDs)+1)

	hash := sha256.Sum256([]byte(hostID))
	peerBint := big.NewInt(0).SetBytes(hash[:]).Uint64()
	unimap[peerBint] = hostID

	for _, pid := range otherPeerIDs {
		hash := sha256.Sum256([]byte(pid))
		bint := big.NewInt(0).SetBytes(hash[:]).Uint64()

		if _, exists := unimap[bint]; !exists {
			unimap[bint] = pid
		}
	}

	orderUints := make([]uint64, len(unimap))
	i := 0
	for val := range unimap {
		orderUints[i] = val
		i++
	}

	sort.Slice(orderUints, func(i, j int) bool {
		return orderUints[i] < orderUints[j]
	})

	var peerIndex int
	orderQueue := make([]peer.ID, len(orderUints))
	for i, bint := range orderUints {
		if bint == peerBint {
			peerIndex = i
		}
		orderQueue[i] = unimap[bint]
	}

	return &ConnAlgo{
		peerIndex:  peerIndex,
		orderQueue: orderQueue,
	}
}

func (cq *ConnAlgo) GetClosestPeers() (peers []peer.ID) {
	size := len(cq.orderQueue)
	if size <= 1 {
		return
	} else if size <= 3 {
		for i, peerID := range cq.orderQueue {
			if cq.peerIndex != i {
				peers = append(peers, peerID)
			}
		}
	} else if size <= 100 {
		// 3条线：前后2条，90度一条
		fi := (cq.peerIndex + 1) % size
		bi := (cq.peerIndex - 1 + size) % size
		i90 := cq.peerIndex + (size / 2)
		peers = []peer.ID{
			cq.orderQueue[fi],
			cq.orderQueue[bi],
			cq.orderQueue[i90],
		}

	} else {
		// 4条线：前后2条，60度2条
		fi := (cq.peerIndex + 1) % size
		bi := (cq.peerIndex - 1 + size) % size
		fi60 := (cq.peerIndex + (size / 3)) % size
		bi60 := (cq.peerIndex - (size / 3) + size) % size
		peers = []peer.ID{
			cq.orderQueue[fi],
			cq.orderQueue[bi],
			cq.orderQueue[fi60],
			cq.orderQueue[bi60],
		}
	}

	return peers
}
