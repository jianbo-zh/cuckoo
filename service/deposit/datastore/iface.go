package datastore

import (
	ipfsds "github.com/ipfs/go-datastore"
)

type DepositIface interface {
	ipfsds.Batching
}

// 离线消息相关协议

// 存储消息
// 1. 用户
// 2. 群组

// 存储周期（3天）
// 用户取走就删除
// 群取走不消失

// 获取消息
// 1. 获取自己的消息列表
// 2. 获取群消息列表（消息ID）
// 3. 获取群消息详情（消息详情）
