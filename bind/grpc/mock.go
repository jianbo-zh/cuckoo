package service

import (
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
)

var account proto.Account
var groups []*proto.GroupFull
var groupMembers []*proto.GroupMember
var groupMessages []*proto.GroupMessage
var contacts []*proto.Contact
var contactMessages []*proto.ContactMessage

func init() {
	account = proto.Account{}
	groups = make([]*proto.GroupFull, 0)
	groupMembers = make([]*proto.GroupMember, 0)
	groupMessages = make([]*proto.GroupMessage, 0)
	contacts = make([]*proto.Contact, 0)
	contactMessages = make([]*proto.ContactMessage, 0)

	groupMessages = append(groupMessages, &proto.GroupMessage{
		Id:      "id",
		GroupId: "groupID",
		Sender: &proto.Peer{
			Id:     "peerID-8081",
			Avatar: "md5_f4b3ae325c43e3fb08c0c7fbbc57ea63.jpg",
			Name:   "name-8081",
		},
		MsgType:    "text",
		MimeType:   "text/plain",
		Payload:    []byte("hello 你好！"),
		CreateTime: time.Now().Unix(),
	})
}
