package service

import (
	"time"

	"github.com/jianbo-zh/dchat/bind/grpc/proto"
)

var account proto.Account
var groups []*proto.Group
var groupMembers []*proto.Contact
var groupMessages []*proto.GroupMessage
var contacts []*proto.Contact
var contactMessages []*proto.ContactMessage
var systemMessage []*proto.SystemMessage

func init() {
	account = proto.Account{}
	groups = make([]*proto.Group, 0)
	groupMembers = make([]*proto.Contact, 0)
	groupMessages = make([]*proto.GroupMessage, 0)
	contacts = make([]*proto.Contact, 0)
	contactMessages = make([]*proto.ContactMessage, 0)
	systemMessage = make([]*proto.SystemMessage, 0)
	systemMessage = append(systemMessage, &proto.SystemMessage{
		ID:      "id",
		GroupID: "",
		Sender: &proto.Contact{
			PeerID: "PeerID1",
			Avatar: "avatar1",
			Name:   "name1",
			Alias:  "alias1",
		},
		Receiver: &proto.Contact{
			PeerID: "PeerID",
			Avatar: "avatar",
			Name:   "name",
			Alias:  "alias",
		},
		Content:    "content",
		CreateTime: int32(time.Now().Unix()),
	})
}
