package myhost

import (
	"time"

	"github.com/jianbo-zh/dchat/internal/types"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	// 确定为在线的时长
	OnlineDuration = 30 * time.Second
	// 清理统计缓存间隔
	ClearStatsInterval = 1000
)

// 清理统计缓存计数
var ClearStatsCounting = 0

// PeersOnlineStats 节点在线统计
func (h *MyHost) PeersOnlineStats(peerIDs []peer.ID) map[peer.ID]types.OnlineState {
	peerStats := make(map[peer.ID]types.OnlineState)
	h.statsMutex.RLock()
	for _, peerID := range peerIDs {
		if _, exists := h.onlineMap[peerID]; exists {
			peerStats[peerID] = types.OnlineStateOnline

		} else if _, exists = h.offlineMap[peerID]; exists {
			peerStats[peerID] = types.OnlineStateOffline

		} else {
			peerStats[peerID] = types.OnlineStateUnknown
		}
	}
	h.statsMutex.RUnlock()
	return peerStats
}

func (h *MyHost) online(peerID peer.ID) {
	h.statsMutex.Lock()

	h.onlineMap[peerID] = time.Now()
	delete(h.offlineMap, peerID)

	ClearStatsCounting++
	if ClearStatsCounting%ClearStatsInterval == 0 {
		nowtime := time.Now()
		for pid, ts := range h.onlineMap {
			if nowtime.Sub(ts) > OnlineDuration { // 过期了
				delete(h.onlineMap, pid)
			}
		}

		for pid, ts := range h.offlineMap {
			if nowtime.Sub(ts) > OnlineDuration { // 过期了
				delete(h.offlineMap, pid)
			}
		}
	}

	h.statsMutex.Unlock()
}

func (h *MyHost) offline(peerID peer.ID) {
	h.statsMutex.Lock()
	h.offlineMap[peerID] = time.Now()
	delete(h.onlineMap, peerID)
	h.statsMutex.Unlock()
}
