package datastore

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/core/peer"
)

/**
 * account 账号Key
 */
const accountKey = "/dchat/peer/account"

type AccountDsKey struct{}

func (a *AccountDsKey) Key() ipfsds.Key {
	return ipfsds.NewKey(accountKey)
}

/**
 * contact 联系人Key
 */
const contactKeyPrefix = "/dchat/peer/peer/"
const contactMsgKeyPrefix = "/dchat/peer/message/"

type ContactDsKey struct{}

func (c *ContactDsKey) Prefix() string {
	return contactKeyPrefix
}

func (c *ContactDsKey) ContactKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(contactKeyPrefix + peerID.String())
}

func (c *ContactDsKey) MsgPrefix(peerID peer.ID) string {
	return contactMsgKeyPrefix + peerID.String() + "/message/"
}

func (c *ContactDsKey) MsgLogPrefix(peerID peer.ID) string {
	return contactMsgKeyPrefix + peerID.String() + "/message/logs/"
}

func (c *ContactDsKey) MsgLogKey(peerID peer.ID, msgID string) ipfsds.Key {
	return ipfsds.NewKey(contactMsgKeyPrefix + peerID.String() + "/message/logs/" + msgID)
}

func (c *ContactDsKey) MsgHeadKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(contactMsgKeyPrefix + peerID.String() + "/message/head")
}

func (c *ContactDsKey) MsgTailKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(contactMsgKeyPrefix + peerID.String() + "/message/tail")
}

/**
 * group 群组Key
 */
const groupKeyPrefix = "/dchat/group/"
const groupSessionKeyPrefix = "/dchat/group/session/"

type GroupDsKey struct{}

func (g *GroupDsKey) AdminPrefix(groupID string) string {
	return groupKeyPrefix + groupID + "/"
}

func (g *GroupDsKey) SessionPrefix() string {
	return groupSessionKeyPrefix
}

func (g *GroupDsKey) AdminLogPrefix(groupID string) string {
	return groupKeyPrefix + groupID + "/admin/logs/"
}

func (g *GroupDsKey) MsgPrefix(groupID string) string {
	return groupKeyPrefix + groupID + "/message/"
}

func (g *GroupDsKey) MsgLogPrefix(groupID string) string {
	return groupKeyPrefix + groupID + "/message/logs/"
}

func (g *GroupDsKey) SessionKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupSessionKeyPrefix + groupID)
}

func (g *GroupDsKey) AdminLogKey(groupID string, msgID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/logs/" + msgID)
}

func (g *GroupDsKey) AdminLogHeadKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/head")
}

func (g *GroupDsKey) AdminLogTailKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/tail")
}

func (g *GroupDsKey) CreatorKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/creator")
}

func (g *GroupDsKey) StateKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/state")
}

func (g *GroupDsKey) NameKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/name")
}

func (g *GroupDsKey) LocalNameKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/localname")
}

func (g *GroupDsKey) AvatarKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/avatar")
}

func (g *GroupDsKey) LocalAvatarKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/localavatar")
}

func (g *GroupDsKey) NoticeKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/notice")
}

func (g *GroupDsKey) MsgLogKey(groupID string, msgID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/message/logs/" + msgID)
}

func (g *GroupDsKey) MsgLogHeadKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/message/head")
}

func (g *GroupDsKey) MsgLogTailKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/message/tail")
}

func (g *GroupDsKey) NetworkLamptimeKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/network/lamportime")
}

/**
 * system 系统Key
 */
const systemMessagePrefix = "/dchat/system/message/"

type SystemDsKey struct{}

func (s *SystemDsKey) MsgLogPrefix() string {
	return systemMessagePrefix
}

func (s *SystemDsKey) MsgLogKey(msgID string) ipfsds.Key {
	return ipfsds.NewKey(systemMessagePrefix + msgID)
}

/**
 * deposit 寄存Key
 */
const depositPeerPrefix = "/dchat/deposit/peer/"
const depositGroupPrefix = "/dchat/deposit/group/"

type DepositDsKey struct{}

func (d *DepositDsKey) PeerMsgLogPrefix(peerID peer.ID) string {
	return depositPeerPrefix + peerID.String() + "/message/logs/"
}

func (d *DepositDsKey) PeerMsgLogKey(peerID peer.ID, msgID string) ipfsds.Key {
	return ipfsds.NewKey(depositPeerPrefix + peerID.String() + "/message/logs/" + msgID)
}

func (d *DepositDsKey) PeerLastAckIDKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(depositPeerPrefix + peerID.String() + "/ackid")
}
