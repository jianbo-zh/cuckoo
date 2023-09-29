package myhost

import (
	"time"

	"github.com/jianbo-zh/dchat/internal/mytype"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	// 确定为在线的时长
	MaxOnlineDuration = 60 * time.Second
	// 清理统计缓存间隔
	ClearStatsInterval = 1000
)

// 清理统计缓存计数
var ClearStatsCounting = 0

// PeersOnlineStats 节点在线统计，onlineDuration 指定多少秒内算才在线，最长60秒
func (h *MyHost) OnlineStats(peerIDs []peer.ID, onlineDuration time.Duration) map[peer.ID]mytype.OnlineState {
	nowtime := time.Now()
	peerStats := make(map[peer.ID]mytype.OnlineState)
	h.statsMutex.RLock()
	for _, peerID := range peerIDs {
		if onlineTime, exists := h.onlineMap[peerID]; exists {
			if nowtime.Sub(onlineTime) <= onlineDuration {
				peerStats[peerID] = mytype.OnlineStateOnline
			} else {
				peerStats[peerID] = mytype.OnlineStateUnknown
			}

		} else if onlineTime, exists = h.offlineMap[peerID]; exists {
			if nowtime.Sub(onlineTime) <= onlineDuration {
				peerStats[peerID] = mytype.OnlineStateOffline
			} else {
				peerStats[peerID] = mytype.OnlineStateUnknown
			}
		} else {
			peerStats[peerID] = mytype.OnlineStateUnknown
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
}

func (h *MyHost) offline(peerID peer.ID) {
	h.statsMutex.Lock()
	h.offlineMap[peerID] = time.Now()
	delete(h.onlineMap, peerID)
	h.statsMutex.Unlock()
}
