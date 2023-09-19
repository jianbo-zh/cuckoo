// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: pb/system.proto

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

type SystemMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string           `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	SystemType  string           `protobuf:"bytes,3,opt,name=systemType,proto3" json:"systemType,omitempty"`
	Group       *SystemMsg_Group `protobuf:"bytes,4,opt,name=group,proto3" json:"group,omitempty"`
	FromPeer    *SystemMsg_Peer  `protobuf:"bytes,5,opt,name=fromPeer,proto3" json:"fromPeer,omitempty"`
	ToPeerId    []byte           `protobuf:"bytes,6,opt,name=toPeerId,proto3" json:"toPeerId,omitempty"`
	Content     string           `protobuf:"bytes,7,opt,name=content,proto3" json:"content,omitempty"`
	SystemState string           `protobuf:"bytes,8,opt,name=systemState,proto3" json:"systemState,omitempty"`
	CreateTime  int64            `protobuf:"varint,9,opt,name=createTime,proto3" json:"createTime,omitempty"`
	UpdateTime  int64            `protobuf:"varint,10,opt,name=updateTime,proto3" json:"updateTime,omitempty"`
}

func (x *SystemMsg) Reset() {
	*x = SystemMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_system_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SystemMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SystemMsg) ProtoMessage() {}

func (x *SystemMsg) ProtoReflect() protoreflect.Message {
	mi := &file_pb_system_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SystemMsg.ProtoReflect.Descriptor instead.
func (*SystemMsg) Descriptor() ([]byte, []int) {
	return file_pb_system_proto_rawDescGZIP(), []int{0}
}

func (x *SystemMsg) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *SystemMsg) GetSystemType() string {
	if x != nil {
		return x.SystemType
	}
	return ""
}

func (x *SystemMsg) GetGroup() *SystemMsg_Group {
	if x != nil {
		return x.Group
	}
	return nil
}

func (x *SystemMsg) GetFromPeer() *SystemMsg_Peer {
	if x != nil {
		return x.FromPeer
	}
	return nil
}

func (x *SystemMsg) GetToPeerId() []byte {
	if x != nil {
		return x.ToPeerId
	}
	return nil
}

func (x *SystemMsg) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *SystemMsg) GetSystemState() string {
	if x != nil {
		return x.SystemState
	}
	return ""
}

func (x *SystemMsg) GetCreateTime() int64 {
	if x != nil {
		return x.CreateTime
	}
	return 0
}

func (x *SystemMsg) GetUpdateTime() int64 {
	if x != nil {
		return x.UpdateTime
	}
	return 0
}

type SystemMsg_Group struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name     string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Avatar   string `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`
	Lamptime uint64 `protobuf:"varint,4,opt,name=lamptime,proto3" json:"lamptime,omitempty"`
}

func (x *SystemMsg_Group) Reset() {
	*x = SystemMsg_Group{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_system_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SystemMsg_Group) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SystemMsg_Group) ProtoMessage() {}

func (x *SystemMsg_Group) ProtoReflect() protoreflect.Message {
	mi := &file_pb_system_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SystemMsg_Group.ProtoReflect.Descriptor instead.
func (*SystemMsg_Group) Descriptor() ([]byte, []int) {
	return file_pb_system_proto_rawDescGZIP(), []int{0, 0}
}

func (x *SystemMsg_Group) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *SystemMsg_Group) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SystemMsg_Group) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *SystemMsg_Group) GetLamptime() uint64 {
	if x != nil {
		return x.Lamptime
	}
	return 0
}

type SystemMsg_Peer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PeerId []byte `protobuf:"bytes,1,opt,name=peerId,proto3" json:"peerId,omitempty"`
	Name   string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Avatar string `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`
}

func (x *SystemMsg_Peer) Reset() {
	*x = SystemMsg_Peer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_system_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SystemMsg_Peer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SystemMsg_Peer) ProtoMessage() {}

func (x *SystemMsg_Peer) ProtoReflect() protoreflect.Message {
	mi := &file_pb_system_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SystemMsg_Peer.ProtoReflect.Descriptor instead.
func (*SystemMsg_Peer) Descriptor() ([]byte, []int) {
	return file_pb_system_proto_rawDescGZIP(), []int{0, 1}
}

func (x *SystemMsg_Peer) GetPeerId() []byte {
	if x != nil {
		return x.PeerId
	}
	return nil
}

func (x *SystemMsg_Peer) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SystemMsg_Peer) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

var File_pb_system_proto protoreflect.FileDescriptor

var file_pb_system_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x70, 0x62, 0x2f, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x09, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x70, 0x62, 0x22, 0xe9, 0x03, 0x0a,
	0x09, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x4d, 0x73, 0x67, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x79,
	0x73, 0x74, 0x65, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x12, 0x30, 0x0a, 0x05, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x73, 0x79, 0x73, 0x74,
	0x65, 0x6d, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x4d, 0x73, 0x67, 0x2e,
	0x47, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x35, 0x0a, 0x08,
	0x66, 0x72, 0x6f, 0x6d, 0x50, 0x65, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19,
	0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x79, 0x73, 0x74, 0x65,
	0x6d, 0x4d, 0x73, 0x67, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x52, 0x08, 0x66, 0x72, 0x6f, 0x6d, 0x50,
	0x65, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x74, 0x6f, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x74, 0x6f, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12,
	0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x79, 0x73,
	0x74, 0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x75,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x1a, 0x5f, 0x0a, 0x05, 0x47,
	0x72, 0x6f, 0x75, 0x70, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74,
	0x61, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72,
	0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6d, 0x70, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x08, 0x6c, 0x61, 0x6d, 0x70, 0x74, 0x69, 0x6d, 0x65, 0x1a, 0x4a, 0x0a, 0x04,
	0x50, 0x65, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x65, 0x65, 0x72, 0x49, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pb_system_proto_rawDescOnce sync.Once
	file_pb_system_proto_rawDescData = file_pb_system_proto_rawDesc
)

func file_pb_system_proto_rawDescGZIP() []byte {
	file_pb_system_proto_rawDescOnce.Do(func() {
		file_pb_system_proto_rawDescData = protoimpl.X.CompressGZIP(file_pb_system_proto_rawDescData)
	})
	return file_pb_system_proto_rawDescData
}

var file_pb_system_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_pb_system_proto_goTypes = []interface{}{
	(*SystemMsg)(nil),       // 0: system.pb.SystemMsg
	(*SystemMsg_Group)(nil), // 1: system.pb.SystemMsg.Group
	(*SystemMsg_Peer)(nil),  // 2: system.pb.SystemMsg.Peer
}
var file_pb_system_proto_depIdxs = []int32{
	1, // 0: system.pb.SystemMsg.group:type_name -> system.pb.SystemMsg.Group
	2, // 1: system.pb.SystemMsg.fromPeer:type_name -> system.pb.SystemMsg.Peer
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_pb_system_proto_init() }
func file_pb_system_proto_init() {
	if File_pb_system_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pb_system_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SystemMsg); i {
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
		file_pb_system_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SystemMsg_Group); i {
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
		file_pb_system_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SystemMsg_Peer); i {
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
			RawDescriptor: file_pb_system_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pb_system_proto_goTypes,
		DependencyIndexes: file_pb_system_proto_depIdxs,
		MessageInfos:      file_pb_system_proto_msgTypes,
	}.Build()
	File_pb_system_proto = out.File
	file_pb_system_proto_rawDesc = nil
	file_pb_system_proto_goTypes = nil
	file_pb_system_proto_depIdxs = nil
}
