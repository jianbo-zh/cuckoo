// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: pb/group_msg.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id         string          `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	GroupId    string          `protobuf:"bytes,2,opt,name=groupId,proto3" json:"groupId,omitempty"`
	Member     *Message_Member `protobuf:"bytes,3,opt,name=member,proto3" json:"member,omitempty"`
	MsgType    string          `protobuf:"bytes,4,opt,name=msgType,proto3" json:"msgType,omitempty"`
	MimeType   string          `protobuf:"bytes,5,opt,name=mimeType,proto3" json:"mimeType,omitempty"`
	Payload    []byte          `protobuf:"bytes,6,opt,name=payload,proto3" json:"payload,omitempty"`
	CreateTime int64           `protobuf:"varint,7,opt,name=createTime,proto3" json:"createTime,omitempty"`
	Lamportime uint64          `protobuf:"varint,8,opt,name=lamportime,proto3" json:"lamportime,omitempty"`
	Signature  []byte          `protobuf:"bytes,9,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_group_msg_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_pb_group_msg_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_pb_group_msg_proto_rawDescGZIP(), []int{0}
}

func (x *Message) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Message) GetGroupId() string {
	if x != nil {
		return x.GroupId
	}
	return ""
}

func (x *Message) GetMember() *Message_Member {
	if x != nil {
		return x.Member
	}
	return nil
}

func (x *Message) GetMsgType() string {
	if x != nil {
		return x.MsgType
	}
	return ""
}

func (x *Message) GetMimeType() string {
	if x != nil {
		return x.MimeType
	}
	return ""
}

func (x *Message) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *Message) GetCreateTime() int64 {
	if x != nil {
		return x.CreateTime
	}
	return 0
}

func (x *Message) GetLamportime() uint64 {
	if x != nil {
		return x.Lamportime
	}
	return 0
}

func (x *Message) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type Message_Member struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name   string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Avatar string `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`
}

func (x *Message_Member) Reset() {
	*x = Message_Member{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_group_msg_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message_Member) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message_Member) ProtoMessage() {}

func (x *Message_Member) ProtoReflect() protoreflect.Message {
	mi := &file_pb_group_msg_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message_Member.ProtoReflect.Descriptor instead.
func (*Message_Member) Descriptor() ([]byte, []int) {
	return file_pb_group_msg_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Message_Member) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Message_Member) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Message_Member) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

var File_pb_group_msg_proto protoreflect.FileDescriptor

var file_pb_group_msg_proto_rawDesc = []byte{
	0x0a, 0x12, 0x70, 0x62, 0x2f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x6d, 0x73, 0x67, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x62,
	0x22, 0xdb, 0x02, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07,
	0x67, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x67,
	0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x12, 0x32, 0x0a, 0x06, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x2e, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x4d, 0x65, 0x6d, 0x62,
	0x65, 0x72, 0x52, 0x06, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x73,
	0x67, 0x54, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x73, 0x67,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6d, 0x69, 0x6d, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6d, 0x69, 0x6d, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x6c, 0x61,
	0x6d, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a,
	0x6c, 0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69,
	0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x1a, 0x44, 0x0a, 0x06, 0x4d, 0x65, 0x6d, 0x62,
	0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pb_group_msg_proto_rawDescOnce sync.Once
	file_pb_group_msg_proto_rawDescData = file_pb_group_msg_proto_rawDesc
)

func file_pb_group_msg_proto_rawDescGZIP() []byte {
	file_pb_group_msg_proto_rawDescOnce.Do(func() {
		file_pb_group_msg_proto_rawDescData = protoimpl.X.CompressGZIP(file_pb_group_msg_proto_rawDescData)
	})
	return file_pb_group_msg_proto_rawDescData
}

var file_pb_group_msg_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pb_group_msg_proto_goTypes = []interface{}{
	(*Message)(nil),        // 0: message.pb.Message
	(*Message_Member)(nil), // 1: message.pb.Message.Member
}
var file_pb_group_msg_proto_depIdxs = []int32{
	1, // 0: message.pb.Message.member:type_name -> message.pb.Message.Member
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_pb_group_msg_proto_init() }
func file_pb_group_msg_proto_init() {
	if File_pb_group_msg_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pb_group_msg_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pb_group_msg_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message_Member); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pb_group_msg_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pb_group_msg_proto_goTypes,
		DependencyIndexes: file_pb_group_msg_proto_depIdxs,
		MessageInfos:      file_pb_group_msg_proto_msgTypes,
	}.Build()
	File_pb_group_msg_proto = out.File
	file_pb_group_msg_proto_rawDesc = nil
	file_pb_group_msg_proto_goTypes = nil
	file_pb_group_msg_proto_depIdxs = nil
}
