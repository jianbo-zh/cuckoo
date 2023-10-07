// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: proto/contact_sync_msg.proto

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

// 同步步骤
type ContactSyncMessage_Type int32

const (
	ContactSyncMessage_SUMMARY    ContactSyncMessage_Type = 0 // 第一步，总摘要        <id1,id2,len>
	ContactSyncMessage_RANGE_HASH ContactSyncMessage_Type = 1 // 第二步，区间hash     <id1, id2, hash>
	ContactSyncMessage_RANGE_IDS  ContactSyncMessage_Type = 2 // 第三步，区间msgid     <id1, id2, msgid...>
	ContactSyncMessage_PUSH_MSG   ContactSyncMessage_Type = 3 // 第四步，发送消息       <msg...>
	ContactSyncMessage_PULL_MSG   ContactSyncMessage_Type = 4 // 第五步，获取消息 <id1, id2 ...>
	ContactSyncMessage_DONE       ContactSyncMessage_Type = 5 // 同步完成
)

// Enum value maps for ContactSyncMessage_Type.
var (
	ContactSyncMessage_Type_name = map[int32]string{
		0: "SUMMARY",
		1: "RANGE_HASH",
		2: "RANGE_IDS",
		3: "PUSH_MSG",
		4: "PULL_MSG",
		5: "DONE",
	}
	ContactSyncMessage_Type_value = map[string]int32{
		"SUMMARY":    0,
		"RANGE_HASH": 1,
		"RANGE_IDS":  2,
		"PUSH_MSG":   3,
		"PULL_MSG":   4,
		"DONE":       5,
	}
)

func (x ContactSyncMessage_Type) Enum() *ContactSyncMessage_Type {
	p := new(ContactSyncMessage_Type)
	*p = x
	return p
}

func (x ContactSyncMessage_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ContactSyncMessage_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_contact_sync_msg_proto_enumTypes[0].Descriptor()
}

func (ContactSyncMessage_Type) Type() protoreflect.EnumType {
	return &file_proto_contact_sync_msg_proto_enumTypes[0]
}

func (x ContactSyncMessage_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ContactSyncMessage_Type.Descriptor instead.
func (ContactSyncMessage_Type) EnumDescriptor() ([]byte, []int) {
	return file_proto_contact_sync_msg_proto_rawDescGZIP(), []int{0, 0}
}

type ContactSyncMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type    ContactSyncMessage_Type `protobuf:"varint,1,opt,name=type,proto3,enum=syncmsg.pb.ContactSyncMessage_Type" json:"type,omitempty"` // 具体步骤
	Payload []byte                  `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`                                    // 具体内容
}

func (x *ContactSyncMessage) Reset() {
	*x = ContactSyncMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_sync_msg_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContactSyncMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContactSyncMessage) ProtoMessage() {}

func (x *ContactSyncMessage) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_sync_msg_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContactSyncMessage.ProtoReflect.Descriptor instead.
func (*ContactSyncMessage) Descriptor() ([]byte, []int) {
	return file_proto_contact_sync_msg_proto_rawDescGZIP(), []int{0}
}

func (x *ContactSyncMessage) GetType() ContactSyncMessage_Type {
	if x != nil {
		return x.Type
	}
	return ContactSyncMessage_SUMMARY
}

func (x *ContactSyncMessage) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

// 第一步，总摘要
type ContactDataSummary struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HeadId   string `protobuf:"bytes,2,opt,name=head_id,json=headId,proto3" json:"head_id,omitempty"`
	TailId   string `protobuf:"bytes,3,opt,name=tail_id,json=tailId,proto3" json:"tail_id,omitempty"`
	Length   int32  `protobuf:"varint,4,opt,name=length,proto3" json:"length,omitempty"`
	Lamptime uint64 `protobuf:"varint,5,opt,name=lamptime,proto3" json:"lamptime,omitempty"`
	IsEnd    bool   `protobuf:"varint,6,opt,name=is_end,json=isEnd,proto3" json:"is_end,omitempty"`
}

func (x *ContactDataSummary) Reset() {
	*x = ContactDataSummary{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_sync_msg_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContactDataSummary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContactDataSummary) ProtoMessage() {}

func (x *ContactDataSummary) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_sync_msg_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContactDataSummary.ProtoReflect.Descriptor instead.
func (*ContactDataSummary) Descriptor() ([]byte, []int) {
	return file_proto_contact_sync_msg_proto_rawDescGZIP(), []int{1}
}

func (x *ContactDataSummary) GetHeadId() string {
	if x != nil {
		return x.HeadId
	}
	return ""
}

func (x *ContactDataSummary) GetTailId() string {
	if x != nil {
		return x.TailId
	}
	return ""
}

func (x *ContactDataSummary) GetLength() int32 {
	if x != nil {
		return x.Length
	}
	return 0
}

func (x *ContactDataSummary) GetLamptime() uint64 {
	if x != nil {
		return x.Lamptime
	}
	return 0
}

func (x *ContactDataSummary) GetIsEnd() bool {
	if x != nil {
		return x.IsEnd
	}
	return false
}

// 第二步，区间HASH
type ContactDataRangeHash struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StartId string `protobuf:"bytes,1,opt,name=start_id,json=startId,proto3" json:"start_id,omitempty"`
	EndId   string `protobuf:"bytes,2,opt,name=end_id,json=endId,proto3" json:"end_id,omitempty"`
	Hash    []byte `protobuf:"bytes,3,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *ContactDataRangeHash) Reset() {
	*x = ContactDataRangeHash{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_sync_msg_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContactDataRangeHash) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContactDataRangeHash) ProtoMessage() {}

func (x *ContactDataRangeHash) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_sync_msg_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContactDataRangeHash.ProtoReflect.Descriptor instead.
func (*ContactDataRangeHash) Descriptor() ([]byte, []int) {
	return file_proto_contact_sync_msg_proto_rawDescGZIP(), []int{2}
}

func (x *ContactDataRangeHash) GetStartId() string {
	if x != nil {
		return x.StartId
	}
	return ""
}

func (x *ContactDataRangeHash) GetEndId() string {
	if x != nil {
		return x.EndId
	}
	return ""
}

func (x *ContactDataRangeHash) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

// 第三步，区间消息ID
type ContactDataRangeIDs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StartId string   `protobuf:"bytes,1,opt,name=start_id,json=startId,proto3" json:"start_id,omitempty"`
	EndId   string   `protobuf:"bytes,2,opt,name=end_id,json=endId,proto3" json:"end_id,omitempty"`
	Ids     []string `protobuf:"bytes,3,rep,name=ids,proto3" json:"ids,omitempty"`
}

func (x *ContactDataRangeIDs) Reset() {
	*x = ContactDataRangeIDs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_sync_msg_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContactDataRangeIDs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContactDataRangeIDs) ProtoMessage() {}

func (x *ContactDataRangeIDs) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_sync_msg_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContactDataRangeIDs.ProtoReflect.Descriptor instead.
func (*ContactDataRangeIDs) Descriptor() ([]byte, []int) {
	return file_proto_contact_sync_msg_proto_rawDescGZIP(), []int{3}
}

func (x *ContactDataRangeIDs) GetStartId() string {
	if x != nil {
		return x.StartId
	}
	return ""
}

func (x *ContactDataRangeIDs) GetEndId() string {
	if x != nil {
		return x.EndId
	}
	return ""
}

func (x *ContactDataRangeIDs) GetIds() []string {
	if x != nil {
		return x.Ids
	}
	return nil
}

// 第五步，获取缺失消息
type ContactDataPullMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ids []string `protobuf:"bytes,1,rep,name=ids,proto3" json:"ids,omitempty"`
}

func (x *ContactDataPullMsg) Reset() {
	*x = ContactDataPullMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_sync_msg_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContactDataPullMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContactDataPullMsg) ProtoMessage() {}

func (x *ContactDataPullMsg) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_sync_msg_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContactDataPullMsg.ProtoReflect.Descriptor instead.
func (*ContactDataPullMsg) Descriptor() ([]byte, []int) {
	return file_proto_contact_sync_msg_proto_rawDescGZIP(), []int{4}
}

func (x *ContactDataPullMsg) GetIds() []string {
	if x != nil {
		return x.Ids
	}
	return nil
}

var File_proto_contact_sync_msg_proto protoreflect.FileDescriptor

var file_proto_contact_sync_msg_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x5f,
	0x73, 0x79, 0x6e, 0x63, 0x5f, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a,
	0x73, 0x79, 0x6e, 0x63, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x62, 0x22, 0xc1, 0x01, 0x0a, 0x12, 0x43,
	0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x53, 0x79, 0x6e, 0x63, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x12, 0x37, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x23, 0x2e, 0x73, 0x79, 0x6e, 0x63, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x62, 0x2e, 0x43, 0x6f, 0x6e,
	0x74, 0x61, 0x63, 0x74, 0x53, 0x79, 0x6e, 0x63, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x22, 0x58, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07,
	0x53, 0x55, 0x4d, 0x4d, 0x41, 0x52, 0x59, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x52, 0x41, 0x4e,
	0x47, 0x45, 0x5f, 0x48, 0x41, 0x53, 0x48, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x52, 0x41, 0x4e,
	0x47, 0x45, 0x5f, 0x49, 0x44, 0x53, 0x10, 0x02, 0x12, 0x0c, 0x0a, 0x08, 0x50, 0x55, 0x53, 0x48,
	0x5f, 0x4d, 0x53, 0x47, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08, 0x50, 0x55, 0x4c, 0x4c, 0x5f, 0x4d,
	0x53, 0x47, 0x10, 0x04, 0x12, 0x08, 0x0a, 0x04, 0x44, 0x4f, 0x4e, 0x45, 0x10, 0x05, 0x22, 0x91,
	0x01, 0x0a, 0x12, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x44, 0x61, 0x74, 0x61, 0x53, 0x75,
	0x6d, 0x6d, 0x61, 0x72, 0x79, 0x12, 0x17, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x49, 0x64, 0x12, 0x17,
	0x0a, 0x07, 0x74, 0x61, 0x69, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x74, 0x61, 0x69, 0x6c, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74,
	0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x12,
	0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6d, 0x70, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x08, 0x6c, 0x61, 0x6d, 0x70, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x15, 0x0a, 0x06, 0x69,
	0x73, 0x5f, 0x65, 0x6e, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x69, 0x73, 0x45,
	0x6e, 0x64, 0x22, 0x5c, 0x0a, 0x14, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x44, 0x61, 0x74,
	0x61, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x48, 0x61, 0x73, 0x68, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x74,
	0x61, 0x72, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x74,
	0x61, 0x72, 0x74, 0x49, 0x64, 0x12, 0x15, 0x0a, 0x06, 0x65, 0x6e, 0x64, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x6e, 0x64, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04,
	0x68, 0x61, 0x73, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68,
	0x22, 0x59, 0x0a, 0x13, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52,
	0x61, 0x6e, 0x67, 0x65, 0x49, 0x44, 0x73, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x72, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74,
	0x49, 0x64, 0x12, 0x15, 0x0a, 0x06, 0x65, 0x6e, 0x64, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x65, 0x6e, 0x64, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x64, 0x73,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x03, 0x69, 0x64, 0x73, 0x22, 0x26, 0x0a, 0x12, 0x43,
	0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x44, 0x61, 0x74, 0x61, 0x50, 0x75, 0x6c, 0x6c, 0x4d, 0x73,
	0x67, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x03,
	0x69, 0x64, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_contact_sync_msg_proto_rawDescOnce sync.Once
	file_proto_contact_sync_msg_proto_rawDescData = file_proto_contact_sync_msg_proto_rawDesc
)

func file_proto_contact_sync_msg_proto_rawDescGZIP() []byte {
	file_proto_contact_sync_msg_proto_rawDescOnce.Do(func() {
		file_proto_contact_sync_msg_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_contact_sync_msg_proto_rawDescData)
	})
	return file_proto_contact_sync_msg_proto_rawDescData
}

var file_proto_contact_sync_msg_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_contact_sync_msg_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_proto_contact_sync_msg_proto_goTypes = []interface{}{
	(ContactSyncMessage_Type)(0), // 0: syncmsg.pb.ContactSyncMessage.Type
	(*ContactSyncMessage)(nil),   // 1: syncmsg.pb.ContactSyncMessage
	(*ContactDataSummary)(nil),   // 2: syncmsg.pb.ContactDataSummary
	(*ContactDataRangeHash)(nil), // 3: syncmsg.pb.ContactDataRangeHash
	(*ContactDataRangeIDs)(nil),  // 4: syncmsg.pb.ContactDataRangeIDs
	(*ContactDataPullMsg)(nil),   // 5: syncmsg.pb.ContactDataPullMsg
}
var file_proto_contact_sync_msg_proto_depIdxs = []int32{
	0, // 0: syncmsg.pb.ContactSyncMessage.type:type_name -> syncmsg.pb.ContactSyncMessage.Type
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proto_contact_sync_msg_proto_init() }
func file_proto_contact_sync_msg_proto_init() {
	if File_proto_contact_sync_msg_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_contact_sync_msg_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContactSyncMessage); i {
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
		file_proto_contact_sync_msg_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContactDataSummary); i {
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
		file_proto_contact_sync_msg_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContactDataRangeHash); i {
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
		file_proto_contact_sync_msg_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContactDataRangeIDs); i {
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
		file_proto_contact_sync_msg_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContactDataPullMsg); i {
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
			RawDescriptor: file_proto_contact_sync_msg_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_contact_sync_msg_proto_goTypes,
		DependencyIndexes: file_proto_contact_sync_msg_proto_depIdxs,
		EnumInfos:         file_proto_contact_sync_msg_proto_enumTypes,
		MessageInfos:      file_proto_contact_sync_msg_proto_msgTypes,
	}.Build()
	File_proto_contact_sync_msg_proto = out.File
	file_proto_contact_sync_msg_proto_rawDesc = nil
	file_proto_contact_sync_msg_proto_goTypes = nil
	file_proto_contact_sync_msg_proto_depIdxs = nil
}