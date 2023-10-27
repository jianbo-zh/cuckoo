package myhost

import (
	"time"

	"github.com/jianbo-zh/dchat/internal/myevent"
	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	// 确定为在线的时长
	MaxOnlineDuration = 10 * time.Second
	// 清理统计缓存间隔
	ClearStatsInterval = 1000
)

// 清理统计缓存计数
var ClearStatsCounting = 0

// OnlineState 节点在线状态
func (h *MyHost) OnlineState(peerID peer.ID) mytype.OnlineState {
	// h.statsMutex.RLock()
	// peerState := mytype.OnlineStateUnknown
	// if _, exists := h.onlineMap[peerID]; exists {
	// 	peerState = mytype.OnlineStateOnline

	// } else if _, exists = h.offlineMap[peerID]; exists {
	// 	peerState = mytype.OnlineStateOffline
	// }
	// h.statsMutex.RUnlock()
	return mytype.OnlineStateUnknown

	// return peerState
}

// PeersOnlineStats 节点在线统计，onlineDuration 指定多少秒内算才在线，最长60秒
func (h *MyHost) OnlineStats(peerIDs []peer.ID) map[peer.ID]mytype.OnlineState {
	h.statsMutex.RLock()
	peerStats := make(map[peer.ID]mytype.OnlineState)
	for _, peerID := range peerIDs {
		if _, exists := h.onlineMap[peerID]; exists {
			peerStats[peerID] = mytype.OnlineStateOnline

		} else if _, exists = h.offlineMap[peerID]; exists {
			peerStats[peerID] = mytype.OnlineStateOffline

		} else {
			peerStats[peerID] = mytype.OnlineStateUnknown
		}
	}
	h.statsMutex.RUnlock()

	return peerStats
}

func (h *MyHost) online(peerID peer.ID) {
	isOnline := false

	h.statsMutex.Lock()
	if _, exists := h.onlineMap[peerID]; !exists {
		isOnline = true
	}
	h.onlineMap[peerID] = time.Now()
	delete(h.offlineMap, peerID)

	ClearStatsCounting++
	if ClearStatsCounting%ClearStatsInterval == 0 {
		nowtime := time.Now()
		for pid, ts := range h.onlineMap {
			if nowtime.Sub(ts) > MaxOnlineDuration { // 过期了
				delete(h.onlineMap, pid)
			}
		}

		for pid, ts := range h.offlineMap {
			if nowtime.Sub(ts) > MaxOnlineDuration { // 过期了
				delete(h.offlineMap, pid)
			}
		}
	}
	h.statsMutex.Unlock()

	if isOnline {
		if err := h.emitters.evtPeerStateChanged.Emit(myevent.EvtPeerStateChanged{
			PeerID: peerID,
			Online: true,
		}); err != nil {
			log.Errorf("emit EvtPeerStateChanged error: %w", err)
		}
	}
}

func (h *MyHost) offline(peerID peer.ID) {
	isOffline := false

	h.statsMutex.Lock()
	if _, exists := h.offlineMap[peerID]; !exists {
		isOffline = true
	}
	h.offlineMap[peerID] = time.Now()
	delete(h.onlineMap, peerID)
	h.statsMutex.Unlock()

	if isOffline {
		if err := h.emitters.evtPeerStateChanged.Emit(myevent.EvtPeerStateChanged{
			PeerID: peerID,
			Online: false,
		}); err != nil {
			log.Errorf("emit EvtPeerStateChanged error: %w", err)
		}
	}
}
