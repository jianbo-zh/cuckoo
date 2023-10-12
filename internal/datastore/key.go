package datastore

import (
	ipfsds "github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/core/peer"
)

/*
account 账号Key
---------------------------------------
/cuckoo/account
---------------------------------------
*/
const accountKey = "/cuckoo/account"

type AccountDsKey struct{}

func (a *AccountDsKey) Key() ipfsds.Key {
	return ipfsds.NewKey(accountKey)
}

/*
session 会话Key
---------------------------------------
/cuckoo/session/list/<sessionID>
/cuckoo/session/<sessionID>/lastmsg
/cuckoo/session/<sessionID>/unreads
---------------------------------------
*/
const sessionPrefix = "/cuckoo/session/"
const sessionListPrefix = "/cuckoo/session/list/"

type SessionDsKey struct{}

func (s *SessionDsKey) ListPrefix() string {
	return sessionListPrefix
}

func (s *SessionDsKey) ListKey(sessionID string) ipfsds.Key {
	return ipfsds.NewKey(sessionListPrefix + sessionID)
}

func (s *SessionDsKey) LastMsgKey(sessionID string) ipfsds.Key {
	return ipfsds.NewKey(sessionPrefix + sessionID + "/lastmsg")
}

func (s *SessionDsKey) UnreadsKey(sessionID string) ipfsds.Key {
	return ipfsds.NewKey(sessionPrefix + sessionID + "/unreads")
}

/*
contact 联系人Key
---------------------------------------
/cuckoo/contact/list/<contactID>
/cuckoo/contact/apply/<contactID>
/cuckoo/contact/<contactID>/info
/cuckoo/contact/<contactID>/state
/cuckoo/contact/<contactID>/message/head
/cuckoo/contact/<contactID>/message/tail
/cuckoo/contact/<contactID>/message/logs/<logID>
---------------------------------------
*/
const contactKeyPrefix = "/cuckoo/contact/"
const contactApplyKeyPrefix = "/cuckoo/contact/apply/"
const contactListPrefix = "/cuckoo/contact/list/"

type ContactDsKey struct{}

func (c *ContactDsKey) ListPrefix() string {
	return contactListPrefix
}

func (c *ContactDsKey) ApplyPrefix() string {
	return contactApplyKeyPrefix
}

func (c *ContactDsKey) ContactPrefix(peerID peer.ID) string {
	return contactKeyPrefix + peerID.String() + "/"
}

func (c *ContactDsKey) MsgPrefix(peerID peer.ID) string {
	return contactKeyPrefix + peerID.String() + "/message/"
}

func (c *ContactDsKey) ListKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(contactListPrefix + peerID.String())
}

func (c *ContactDsKey) ApplyKey(peerID peer.ID) ipfsds.Key {
	return ipfsds.NewKey(contactApplyKeyPrefix + peerID.String())
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

/*
group 群组Key
---------------------------------------
/cuckoo/group/list/<groupID>
/cuckoo/group/<groupID>/admin/alias
/cuckoo/group/<groupID>/admin/creator
/cuckoo/group/<groupID>/admin/name
/cuckoo/group/<groupID>/admin/avatar
/cuckoo/group/<groupID>/admin/state
/cuckoo/group/<groupID>/admin/notice
/cuckoo/group/<groupID>/admin/createtime
/cuckoo/group/<groupID>/admin/autojoingroup
/cuckoo/group/<groupID>/admin/depositpeerid
/cuckoo/group/<groupID>/admin/members
/cuckoo/group/<groupID>/admin/agree_members
/cuckoo/group/<groupID>/admin/lamptime
/cuckoo/group/<groupID>/admin/logs/<logID>
/cuckoo/group/<groupID/message/head
/cuckoo/group/<groupID/message/tail
/cuckoo/group/<groupID/message/logs/<logID>
/cuckoo/group/<groupID/network/lamptime
---------------------------------------
*/
const groupKeyPrefix = "/cuckoo/group/"
const groupListPrefix = "/cuckoo/group/list/"

type GroupDsKey struct{}

func (g *GroupDsKey) AdminPrefix(groupID string) string {
	return groupKeyPrefix + groupID + "/"
}

func (g *GroupDsKey) ListPrefix() string {
	return groupListPrefix
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

func (g *GroupDsKey) ListKey(groupID string) ipfsds.Key {
	return ipfsds.NewKey(groupListPrefix + groupID)
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

func (g *GroupDsKey) DepositAddressKey(groupID string) ipfsds.Key {
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

/*
system 系统Key
---------------------------------------
/cuckoo/systemmsg/<logID>
---------------------------------------
*/
const systemMessagePrefix = "/cuckoo/systemmsg/"

type SystemDsKey struct{}

func (s *SystemDsKey) MsgLogPrefix() string {
	return systemMessagePrefix
}

func (s *SystemDsKey) MsgLogKey(msgID string) ipfsds.Key {
	return ipfsds.NewKey(systemMessagePrefix + msgID)
}

/*
deposit 寄存Key
---------------------------------------
/cuckoo/deposit/peer/<peerID>/lastid
/cuckoo/deposit/peer/<peerID>/message/logs/<logID>
/cuckoo/deposit/group/<groupID>/lastid
/cuckoo/deposit/group/<groupID>/message/logs/<logID>
---------------------------------------
*/
const depositPeerPrefix = "/cuckoo/deposit/peer/"
const depositGroupPrefix = "/cuckoo/deposit/group/"

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

/*
file 文件服务
---------------------------------------
/cuckoo/filetable/file/<fileID>/<sessionID>
/cuckoo/filetable/session/<sessionID>/resource/<fileID>
/cuckoo/filetable/session/<sessionID>/file/<fileID>
/cuckoo/filetable/session/<sessionID>/filetime/<fileID>
---------------------------------------
*/
const FilePrefix = "/cuckoo/filetable/file/"
const SessionPrefix = "/cuckoo/filetable/session/"

type FileDsKey struct{}

func (d *FileDsKey) FileSessionPrefix(fileID string) string {
	return FilePrefix + fileID + "/"
}

func (d *FileDsKey) FileSessionKey(fileID string, sessionID string) ipfsds.Key {
	return ipfsds.NewKey(FilePrefix + fileID + "/" + sessionID)
}

func (d *FileDsKey) SessionResourcePrefix(sessionID string) string {
	return SessionPrefix + sessionID + "/resource/"
}

func (d *FileDsKey) SessionResourceKey(sessionID string, fileID string) ipfsds.Key {
	return ipfsds.NewKey(SessionPrefix + sessionID + "/resource/" + fileID)
}

func (d *FileDsKey) SessionFilePrefix(sessionID string) string {
	return SessionPrefix + sessionID + "/file/"
}

func (d *FileDsKey) SessionFileKey(sessionID string, fileID string) ipfsds.Key {
	return ipfsds.NewKey(SessionPrefix + sessionID + "/file/" + fileID)
}

func (d *FileDsKey) SessionFileTimePrefix(sessionID string) string {
	return SessionPrefix + sessionID + "/filetime/"
}

func (d *FileDsKey) SessionFileTimeKey(sessionID string, fileID string) ipfsds.Key {
	return ipfsds.NewKey(SessionPrefix + sessionID + "/filetime/" + fileID)
}
