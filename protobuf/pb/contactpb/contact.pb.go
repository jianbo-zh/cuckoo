// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: proto/contact.proto

package contactpb

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

type ContactState int32

const (
	ContactState_None   ContactState = 0
	ContactState_Apply  ContactState = 1
	ContactState_Normal ContactState = 2
)

// Enum value maps for ContactState.
var (
	ContactState_name = map[int32]string{
		0: "None",
		1: "Apply",
		2: "Normal",
	}
	ContactState_value = map[string]int32{
		"None":   0,
		"Apply":  1,
		"Normal": 2,
	}
)

func (x ContactState) Enum() *ContactState {
	p := new(ContactState)
	*p = x
	return p
}

func (x ContactState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ContactState) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_contact_proto_enumTypes[0].Descriptor()
}

func (ContactState) Type() protoreflect.EnumType {
	return &file_proto_contact_proto_enumTypes[0]
}

func (x ContactState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ContactState.Descriptor instead.
func (ContactState) EnumDescriptor() ([]byte, []int) {
	return file_proto_contact_proto_rawDescGZIP(), []int{0}
}

type Contact struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id             []byte       `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name           string       `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Avatar         string       `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`
	DepositAddress []byte       `protobuf:"bytes,4,opt,name=depositAddress,proto3" json:"depositAddress,omitempty"`
	State          ContactState `protobuf:"varint,5,opt,name=state,proto3,enum=contact.pb.ContactState" json:"state,omitempty"`
	CreateTime     int64        `protobuf:"varint,6,opt,name=createTime,proto3" json:"createTime,omitempty"`
	AccessTime     int64        `protobuf:"varint,7,opt,name=accessTime,proto3" json:"accessTime,omitempty"`
}

func (x *Contact) Reset() {
	*x = Contact{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Contact) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Contact) ProtoMessage() {}

func (x *Contact) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Contact.ProtoReflect.Descriptor instead.
func (*Contact) Descriptor() ([]byte, []int) {
	return file_proto_contact_proto_rawDescGZIP(), []int{0}
}

func (x *Contact) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Contact) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Contact) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *Contact) GetDepositAddress() []byte {
	if x != nil {
		return x.DepositAddress
	}
	return nil
}

func (x *Contact) GetState() ContactState {
	if x != nil {
		return x.State
	}
	return ContactState_None
}

func (x *Contact) GetCreateTime() int64 {
	if x != nil {
		return x.CreateTime
	}
	return 0
}

func (x *Contact) GetAccessTime() int64 {
	if x != nil {
		return x.AccessTime
	}
	return 0
}

type ContactPeer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id             []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name           string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Avatar         string `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`
	DepositAddress []byte `protobuf:"bytes,4,opt,name=depositAddress,proto3" json:"depositAddress,omitempty"`
}

func (x *ContactPeer) Reset() {
	*x = ContactPeer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContactPeer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContactPeer) ProtoMessage() {}

func (x *ContactPeer) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContactPeer.ProtoReflect.Descriptor instead.
func (*ContactPeer) Descriptor() ([]byte, []int) {
	return file_proto_contact_proto_rawDescGZIP(), []int{1}
}

func (x *ContactPeer) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *ContactPeer) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ContactPeer) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *ContactPeer) GetDepositAddress() []byte {
	if x != nil {
		return x.DepositAddress
	}
	return nil
}

type ContactPeerState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id             []byte       `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name           string       `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Avatar         string       `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`
	DepositAddress []byte       `protobuf:"bytes,4,opt,name=depositAddress,proto3" json:"depositAddress,omitempty"`
	State          ContactState `protobuf:"varint,5,opt,name=state,proto3,enum=contact.pb.ContactState" json:"state,omitempty"`
}

func (x *ContactPeerState) Reset() {
	*x = ContactPeerState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContactPeerState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContactPeerState) ProtoMessage() {}

func (x *ContactPeerState) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContactPeerState.ProtoReflect.Descriptor instead.
func (*ContactPeerState) Descriptor() ([]byte, []int) {
	return file_proto_contact_proto_rawDescGZIP(), []int{2}
}

func (x *ContactPeerState) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *ContactPeerState) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ContactPeerState) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *ContactPeerState) GetDepositAddress() []byte {
	if x != nil {
		return x.DepositAddress
	}
	return nil
}

func (x *ContactPeerState) GetState() ContactState {
	if x != nil {
		return x.State
	}
	return ContactState_None
}

var File_proto_contact_proto protoreflect.FileDescriptor

var file_proto_contact_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x2e, 0x70,
	0x62, 0x22, 0xdd, 0x01, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x12, 0x26, 0x0a, 0x0e, 0x64, 0x65, 0x70,
	0x6f, 0x73, 0x69, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x0e, 0x64, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x12, 0x2e, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x18, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x2e, 0x70, 0x62, 0x2e, 0x43, 0x6f,
	0x6e, 0x74, 0x61, 0x63, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d,
	0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x54, 0x69, 0x6d, 0x65, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x54, 0x69, 0x6d,
	0x65, 0x22, 0x71, 0x0a, 0x0b, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x50, 0x65, 0x65, 0x72,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x12, 0x26, 0x0a, 0x0e,
	0x64, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x64, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x22, 0xa6, 0x01, 0x0a, 0x10, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74,
	0x50, 0x65, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a,
	0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61,
	0x76, 0x61, 0x74, 0x61, 0x72, 0x12, 0x26, 0x0a, 0x0e, 0x64, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74,
	0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x64,
	0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x2e, 0x0a,
	0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x18, 0x2e, 0x63,
	0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x2e, 0x70, 0x62, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63,
	0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2a, 0x2f, 0x0a,
	0x0c, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x08, 0x0a,
	0x04, 0x4e, 0x6f, 0x6e, 0x65, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x41, 0x70, 0x70, 0x6c, 0x79,
	0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x4e, 0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x10, 0x02, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_contact_proto_rawDescOnce sync.Once
	file_proto_contact_proto_rawDescData = file_proto_contact_proto_rawDesc
)

func file_proto_contact_proto_rawDescGZIP() []byte {
	file_proto_contact_proto_rawDescOnce.Do(func() {
		file_proto_contact_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_contact_proto_rawDescData)
	})
	return file_proto_contact_proto_rawDescData
}

var file_proto_contact_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_contact_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_proto_contact_proto_goTypes = []interface{}{
	(ContactState)(0),        // 0: contact.pb.ContactState
	(*Contact)(nil),          // 1: contact.pb.Contact
	(*ContactPeer)(nil),      // 2: contact.pb.ContactPeer
	(*ContactPeerState)(nil), // 3: contact.pb.ContactPeerState
}
var file_proto_contact_proto_depIdxs = []int32{
	0, // 0: contact.pb.Contact.state:type_name -> contact.pb.ContactState
	0, // 1: contact.pb.ContactPeerState.state:type_name -> contact.pb.ContactState
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proto_contact_proto_init() }
func file_proto_contact_proto_init() {
	if File_proto_contact_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_contact_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Contact); i {
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
		file_proto_contact_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContactPeer); i {
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
		file_proto_contact_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContactPeerState); i {
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
			RawDescriptor: file_proto_contact_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_contact_proto_goTypes,
		DependencyIndexes: file_proto_contact_proto_depIdxs,
		EnumInfos:         file_proto_contact_proto_enumTypes,
		MessageInfos:      file_proto_contact_proto_msgTypes,
	}.Build()
	File_proto_contact_proto = out.File
	file_proto_contact_proto_rawDesc = nil
	file_proto_contact_proto_goTypes = nil
	file_proto_contact_proto_depIdxs = nil
}