package peer

import (
	"context"
	"time"

	"github.com/jianbo-zh/dchat/service/peer/protocol/message/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

// 联系人相关服务

/*

通讯录-----
查看通讯录
查找联系人
添加进通讯录
发起聊天
查看聊天内容
新消息（未读）

*/

type PeerServiceIface interface {
	Ping(context.Context, peer.ID) (time.Duration, error)
	GetMessages(context.Context, peer.ID) ([]*pb.Message, error)
	SendTextMessage(context.Context, peer.ID, string) error
	SendGroupInviteMessage(context.Context, peer.ID, string) error
}
