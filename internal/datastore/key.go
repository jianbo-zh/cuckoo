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
const contactKeyPrefix = "/dchat/peer/"
const contactSessionKeyPrefix = "/dchat/peer/session/"

type ContactDsKey struct{}

func (c *ContactDsKey) SessionPrefix() string {
	return contactSessionKeyPrefix
}
func (c *ContactDsKey) ContactPrefix(peerID peer.ID) string {
	return contactKeyPrefix + peerID.String() + "/"
}

func (c *ContactDsKey) MsgPrefix(peerID peer.ID) string {
	return contactKeyPrefix + peerID.String() + "/message/"
}

func (c *ContactDsKey) SessionKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(contactSessionKeyPrefix + peerID.String())
}

func (c *ContactDsKey) DetailKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(contactKeyPrefix + peerID.String() + "/detail")
}

func (c *ContactDsKey) StateKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(contactKeyPrefix + peerID.String() + "/state")
}

func (c *ContactDsKey) MsgLogPrefix(peerID peer.ID) string {
	return contactKeyPrefix + peerID.String() + "/message/logs/"
}

func (c *ContactDsKey) MsgLogKey(peerID peer.ID, msgID string) ipfsds.Key {
	return ipfsds.NewKey(contactKeyPrefix + peerID.String() + "/message/logs/" + msgID)
}

func (c *ContactDsKey) MsgHeadKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(contactKeyPrefix + peerID.String() + "/message/head")
}

func (c *ContactDsKey) MsgTailKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(contactKeyPrefix + peerID.String() + "/message/tail")
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

func (g *GroupDsKey) CreatorKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/creator")
}

func (g *GroupDsKey) StateKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/state")
}

func (g *GroupDsKey) NameKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/name")
}

func (g *GroupDsKey) AvatarKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/avatar")
}

func (g *GroupDsKey) NoticeKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/notice")
}

func (g *GroupDsKey) CreateTimeKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/createtime")
}

func (g *GroupDsKey) AutoJoinGroupKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/autojoingroup")
}

func (g *GroupDsKey) DepositPeerIDKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/depositpeerid")
}

func (g *GroupDsKey) AliasKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/alias")
}

func (g *GroupDsKey) MembersKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/members") // 正式成员
}

func (g *GroupDsKey) AgreeMembersKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/agree_members") // 包括准成员（包括群主同意入群的）
}

func (g *GroupDsKey) AdminLamptimeKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/admin/lamptime")
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
	return ipfsds.NewKey(groupKeyPrefix + groupID + "/network/lamptime")
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

func (d *DepositDsKey) GroupMsgLogPrefix(groupID string) string {
	return depositGroupPrefix + groupID + "/message/logs/"
}

func (d *DepositDsKey) PeerMsgLogKey(peerID peer.ID, msgID string) ipfsds.Key {
	return ipfsds.NewKey(depositPeerPrefix + peerID.String() + "/message/logs/" + msgID)
}

func (d *DepositDsKey) PeerLastIDKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(depositPeerPrefix + peerID.String() + "/lastid")
}

func (d *DepositDsKey) GroupMsgLogKey(groupID string, msgID string) ipfsds.Key {
	return ipfsds.NewKey(depositGroupPrefix + groupID + "/message/logs/" + msgID)
}

func (d *DepositDsKey) GroupLastIDKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(depositGroupPrefix + groupID + "/lastid")
}

/**
 * file 文件服务
 */
const FilePrefix = "/dchat/file/"

type FileDsKey struct{}

func (d *FileDsKey) TablePrefix() string {
	return FilePrefix + "/table/"
}

func (d *FileDsKey) FileKey(hashAlgo, hashValue string) ipfsds.Key {
	return ipfsds.NewKey(FilePrefix + "/table/" + hashAlgo + "_" + hashValue)
}
