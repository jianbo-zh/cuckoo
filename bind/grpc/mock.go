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
var systemMessage []*proto.SystemMessage

func init() {
	account = proto.Account{
		PeerID:                  "peerID",
		Avatar:                  "md5_4c6c2a483e5f2d06140c7313de06448f.jpg",
		Name:                    "name",
		AddContactWithoutReview: true,
		JoinGroupWithoutReview:  true,
	}
	groups = make([]*proto.GroupFull, 0)
	groupMembers = make([]*proto.GroupMember, 0)
	groupMessages = make([]*proto.GroupMessage, 0)
	contacts = make([]*proto.Contact, 0)
	contactMessages = make([]*proto.ContactMessage, 0)
	systemMessage = make([]*proto.SystemMessage, 0)
	systemMessage = append(systemMessage, &proto.SystemMessage{
		ID:      "id",
		GroupID: "",
		Sender: &proto.Contact{
			PeerID: "PeerID1",
			Avatar: "md5_4c6c2a483e5f2d06140c7313de06448f.jpg",
			Name:   "name1",
			Alias:  "alias1",
		},
		Receiver: &proto.Contact{
			PeerID: "PeerID",
			Avatar: "avatar",
			Name:   "name",
			Alias:  "alias",
		},
		SystemOperate: proto.SystemOperate_APPLY_ADD_CONTACT,
		Content:       "content",
		CreateTime:    int32(time.Now().Unix()),
	})

	contactMessages = append(contactMessages, &proto.ContactMessage{
		ID: "id",
		Sender: &proto.Contact{
			PeerID: "peerID-8081",
			Avatar: "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
			Name:   "name-8081",
			Alias:  "",
		},
		Receiver: &proto.Contact{
			PeerID: account.PeerID,
			Avatar: account.Avatar,
			Name:   account.Name,
			Alias:  "",
		},
		MsgType:    proto.MsgType_TEXT_MSG,
		MimeType:   "text/plain",
		Data:       []byte("你好，大傻瓜"),
		CreateTime: time.Now().Unix(),
	})

	groupMessages = append(groupMessages, &proto.GroupMessage{
		ID:      "id",
		GroupID: "groupID",
		Sender: &proto.Contact{
			PeerID: "peerID-8081",
			Avatar: "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
			Name:   "name-8081",
			Alias:  "",
		},
		MsgType:    proto.MsgType_TEXT_MSG,
		MimeType:   "text/plain",
		Data:       []byte("hello 你好！"),
		CreateTime: time.Now().Unix(),
	})

	contacts = append(contacts, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	}, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	}, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	}, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	}, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	}, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	}, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	}, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	}, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	}, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	}, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	}, &proto.Contact{
		PeerID:      "peerID",
		Avatar:      "md5_490ecc5cbb75e4135eabfb2c7a7629bd.jpg",
		Name:        "name-asdf",
		Alias:       "name-alias",
		LastMessage: "ssdfasfsss",
		UpdateTime:  1234567890,
	})
}
