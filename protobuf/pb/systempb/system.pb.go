// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: proto/system.proto

package systempb

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

type SystemMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string               `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	SystemType  string               `protobuf:"bytes,3,opt,name=systemType,proto3" json:"systemType,omitempty"`
	Group       *SystemMessage_Group `protobuf:"bytes,4,opt,name=group,proto3" json:"group,omitempty"`
	FromPeer    *SystemMessage_Peer  `protobuf:"bytes,5,opt,name=fromPeer,proto3" json:"fromPeer,omitempty"`
	ToPeerId    []byte               `protobuf:"bytes,6,opt,name=toPeerId,proto3" json:"toPeerId,omitempty"`
	Content     string               `protobuf:"bytes,7,opt,name=content,proto3" json:"content,omitempty"`
	SystemState string               `protobuf:"bytes,8,opt,name=systemState,proto3" json:"systemState,omitempty"`
	CreateTime  int64                `protobuf:"varint,9,opt,name=createTime,proto3" json:"createTime,omitempty"`
	UpdateTime  int64                `protobuf:"varint,10,opt,name=updateTime,proto3" json:"updateTime,omitempty"`
}

func (x *SystemMessage) Reset() {
	*x = SystemMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_system_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SystemMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SystemMessage) ProtoMessage() {}

func (x *SystemMessage) ProtoReflect() protoreflect.Message {
	mi := &file_proto_system_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SystemMessage.ProtoReflect.Descriptor instead.
func (*SystemMessage) Descriptor() ([]byte, []int) {
	return file_proto_system_proto_rawDescGZIP(), []int{0}
}

func (x *SystemMessage) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *SystemMessage) GetSystemType() string {
	if x != nil {
		return x.SystemType
	}
	return ""
}

func (x *SystemMessage) GetGroup() *SystemMessage_Group {
	if x != nil {
		return x.Group
	}
	return nil
}

func (x *SystemMessage) GetFromPeer() *SystemMessage_Peer {
	if x != nil {
		return x.FromPeer
	}
	return nil
}

func (x *SystemMessage) GetToPeerId() []byte {
	if x != nil {
		return x.ToPeerId
	}
	return nil
}

func (x *SystemMessage) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *SystemMessage) GetSystemState() string {
	if x != nil {
		return x.SystemState
	}
	return ""
}

func (x *SystemMessage) GetCreateTime() int64 {
	if x != nil {
		return x.CreateTime
	}
	return 0
}

func (x *SystemMessage) GetUpdateTime() int64 {
	if x != nil {
		return x.UpdateTime
	}
	return 0
}

type SystemMessage_Group struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name     string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Avatar   string `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`
	Lamptime uint64 `protobuf:"varint,4,opt,name=lamptime,proto3" json:"lamptime,omitempty"`
}

func (x *SystemMessage_Group) Reset() {
	*x = SystemMessage_Group{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_system_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SystemMessage_Group) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SystemMessage_Group) ProtoMessage() {}

func (x *SystemMessage_Group) ProtoReflect() protoreflect.Message {
	mi := &file_proto_system_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SystemMessage_Group.ProtoReflect.Descriptor instead.
func (*SystemMessage_Group) Descriptor() ([]byte, []int) {
	return file_proto_system_proto_rawDescGZIP(), []int{0, 0}
}

func (x *SystemMessage_Group) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *SystemMessage_Group) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SystemMessage_Group) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *SystemMessage_Group) GetLamptime() uint64 {
	if x != nil {
		return x.Lamptime
	}
	return 0
}

type SystemMessage_Peer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PeerId []byte `protobuf:"bytes,1,opt,name=peerId,proto3" json:"peerId,omitempty"`
	Name   string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Avatar string `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`
}

func (x *SystemMessage_Peer) Reset() {
	*x = SystemMessage_Peer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_system_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SystemMessage_Peer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SystemMessage_Peer) ProtoMessage() {}

func (x *SystemMessage_Peer) ProtoReflect() protoreflect.Message {
	mi := &file_proto_system_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SystemMessage_Peer.ProtoReflect.Descriptor instead.
func (*SystemMessage_Peer) Descriptor() ([]byte, []int) {
	return file_proto_system_proto_rawDescGZIP(), []int{0, 1}
}

func (x *SystemMessage_Peer) GetPeerId() []byte {
	if x != nil {
		return x.PeerId
	}
	return nil
}

func (x *SystemMessage_Peer) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SystemMessage_Peer) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

var File_proto_system_proto protoreflect.FileDescriptor

var file_proto_system_proto_rawDesc = []byte{
	0x0a, 0x12, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x70, 0x62, 0x22,
	0xf5, 0x03, 0x0a, 0x0d, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x34, 0x0a, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1e, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x79, 0x73,
	0x74, 0x65, 0x6d, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x52, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x39, 0x0a, 0x08, 0x66, 0x72, 0x6f, 0x6d, 0x50,
	0x65, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x73, 0x79, 0x73, 0x74,
	0x65, 0x6d, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x52, 0x08, 0x66, 0x72, 0x6f, 0x6d, 0x50, 0x65,
	0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x74, 0x6f, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x74, 0x6f, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x18,
	0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x79, 0x73, 0x74,
	0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73,
	0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x75, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x1a, 0x5f, 0x0a, 0x05, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61,
	0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x12,
	0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6d, 0x70, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x08, 0x6c, 0x61, 0x6d, 0x70, 0x74, 0x69, 0x6d, 0x65, 0x1a, 0x4a, 0x0a, 0x04, 0x50,
	0x65, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x65, 0x65, 0x72, 0x49, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_system_proto_rawDescOnce sync.Once
	file_proto_system_proto_rawDescData = file_proto_system_proto_rawDesc
)

func file_proto_system_proto_rawDescGZIP() []byte {
	file_proto_system_proto_rawDescOnce.Do(func() {
		file_proto_system_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_system_proto_rawDescData)
	})
	return file_proto_system_proto_rawDescData
}

var file_proto_system_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_proto_system_proto_goTypes = []interface{}{
	(*SystemMessage)(nil),       // 0: system.pb.SystemMessage
	(*SystemMessage_Group)(nil), // 1: system.pb.SystemMessage.Group
	(*SystemMessage_Peer)(nil),  // 2: system.pb.SystemMessage.Peer
}
var file_proto_system_proto_depIdxs = []int32{
	1, // 0: system.pb.SystemMessage.group:type_name -> system.pb.SystemMessage.Group
	2, // 1: system.pb.SystemMessage.fromPeer:type_name -> system.pb.SystemMessage.Peer
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proto_system_proto_init() }
func file_proto_system_proto_init() {
	if File_proto_system_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_system_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SystemMessage); i {
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
		file_proto_system_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SystemMessage_Group); i {
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
		file_proto_system_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SystemMessage_Peer); i {
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
			RawDescriptor: file_proto_system_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_system_proto_goTypes,
		DependencyIndexes: file_proto_system_proto_depIdxs,
		MessageInfos:      file_proto_system_proto_msgTypes,
	}.Build()
	File_proto_system_proto = out.File
	file_proto_system_proto_rawDesc = nil
	file_proto_system_proto_goTypes = nil
	file_proto_system_proto_depIdxs = nil
}
