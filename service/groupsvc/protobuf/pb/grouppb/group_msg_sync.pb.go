// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: proto/group_msg_sync.proto

package grouppb

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
type GroupSyncMessage_Type int32

const (
	GroupSyncMessage_INIT       GroupSyncMessage_Type = 0 // 初始化，传递groupID过去
	GroupSyncMessage_SUMMARY    GroupSyncMessage_Type = 1 // 第一步，总摘要        <id1,id2,len>
	GroupSyncMessage_RANGE_HASH GroupSyncMessage_Type = 2 // 第二步，区间hash     <id1, id2, hash>
	GroupSyncMessage_RANGE_IDS  GroupSyncMessage_Type = 3 // 第三步，区间msgid     <id1, id2, msgid...>
	GroupSyncMessage_PUSH_MSG   GroupSyncMessage_Type = 4 // 第四步，发送消息       <msg...>
	GroupSyncMessage_PULL_MSG   GroupSyncMessage_Type = 5 // 第五步，获取消息 <id1, id2 ...>
	GroupSyncMessage_DONE       GroupSyncMessage_Type = 6 // 同步完成
)

// Enum value maps for GroupSyncMessage_Type.
var (
	GroupSyncMessage_Type_name = map[int32]string{
		0: "INIT",
		1: "SUMMARY",
		2: "RANGE_HASH",
		3: "RANGE_IDS",
		4: "PUSH_MSG",
		5: "PULL_MSG",
		6: "DONE",
	}
	GroupSyncMessage_Type_value = map[string]int32{
		"INIT":       0,
		"SUMMARY":    1,
		"RANGE_HASH": 2,
		"RANGE_IDS":  3,
		"PUSH_MSG":   4,
		"PULL_MSG":   5,
		"DONE":       6,
	}
)

func (x GroupSyncMessage_Type) Enum() *GroupSyncMessage_Type {
	p := new(GroupSyncMessage_Type)
	*p = x
	return p
}

func (x GroupSyncMessage_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (GroupSyncMessage_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_group_msg_sync_proto_enumTypes[0].Descriptor()
}

func (GroupSyncMessage_Type) Type() protoreflect.EnumType {
	return &file_proto_group_msg_sync_proto_enumTypes[0]
}

func (x GroupSyncMessage_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GroupSyncMessage_Type.Descriptor instead.
func (GroupSyncMessage_Type) EnumDescriptor() ([]byte, []int) {
	return file_proto_group_msg_sync_proto_rawDescGZIP(), []int{0, 0}
}

type GroupSyncMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupId string                `protobuf:"bytes,1,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`                // 组ID
	Type    GroupSyncMessage_Type `protobuf:"varint,2,opt,name=type,proto3,enum=sync.pb.GroupSyncMessage_Type" json:"type,omitempty"` // 具体步骤
	Payload []byte                `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`                               // 具体内容
}

func (x *GroupSyncMessage) Reset() {
	*x = GroupSyncMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_group_msg_sync_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupSyncMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupSyncMessage) ProtoMessage() {}

func (x *GroupSyncMessage) ProtoReflect() protoreflect.Message {
	mi := &file_proto_group_msg_sync_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupSyncMessage.ProtoReflect.Descriptor instead.
func (*GroupSyncMessage) Descriptor() ([]byte, []int) {
	return file_proto_group_msg_sync_proto_rawDescGZIP(), []int{0}
}

func (x *GroupSyncMessage) GetGroupId() string {
	if x != nil {
		return x.GroupId
	}
	return ""
}

func (x *GroupSyncMessage) GetType() GroupSyncMessage_Type {
	if x != nil {
		return x.Type
	}
	return GroupSyncMessage_INIT
}

func (x *GroupSyncMessage) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

// 第一步，总摘要
type GroupMessageSummary struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupId  string `protobuf:"bytes,1,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
	HeadId   string `protobuf:"bytes,2,opt,name=head_id,json=headId,proto3" json:"head_id,omitempty"`
	TailId   string `protobuf:"bytes,3,opt,name=tail_id,json=tailId,proto3" json:"tail_id,omitempty"`
	Length   int32  `protobuf:"varint,4,opt,name=length,proto3" json:"length,omitempty"`
	Lamptime uint64 `protobuf:"varint,5,opt,name=lamptime,proto3" json:"lamptime,omitempty"`
	IsEnd    bool   `protobuf:"varint,6,opt,name=is_end,json=isEnd,proto3" json:"is_end,omitempty"`
}

func (x *GroupMessageSummary) Reset() {
	*x = GroupMessageSummary{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_group_msg_sync_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupMessageSummary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupMessageSummary) ProtoMessage() {}

func (x *GroupMessageSummary) ProtoReflect() protoreflect.Message {
	mi := &file_proto_group_msg_sync_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupMessageSummary.ProtoReflect.Descriptor instead.
func (*GroupMessageSummary) Descriptor() ([]byte, []int) {
	return file_proto_group_msg_sync_proto_rawDescGZIP(), []int{1}
}

func (x *GroupMessageSummary) GetGroupId() string {
	if x != nil {
		return x.GroupId
	}
	return ""
}

func (x *GroupMessageSummary) GetHeadId() string {
	if x != nil {
		return x.HeadId
	}
	return ""
}

func (x *GroupMessageSummary) GetTailId() string {
	if x != nil {
		return x.TailId
	}
	return ""
}

func (x *GroupMessageSummary) GetLength() int32 {
	if x != nil {
		return x.Length
	}
	return 0
}

func (x *GroupMessageSummary) GetLamptime() uint64 {
	if x != nil {
		return x.Lamptime
	}
	return 0
}

func (x *GroupMessageSummary) GetIsEnd() bool {
	if x != nil {
		return x.IsEnd
	}
	return false
}

// 第二步，区间HASH
type GroupMessageRangeHash struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupId string `protobuf:"bytes,1,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
	StartId string `protobuf:"bytes,2,opt,name=start_id,json=startId,proto3" json:"start_id,omitempty"`
	EndId   string `protobuf:"bytes,3,opt,name=end_id,json=endId,proto3" json:"end_id,omitempty"`
	Hash    []byte `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *GroupMessageRangeHash) Reset() {
	*x = GroupMessageRangeHash{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_group_msg_sync_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupMessageRangeHash) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupMessageRangeHash) ProtoMessage() {}

func (x *GroupMessageRangeHash) ProtoReflect() protoreflect.Message {
	mi := &file_proto_group_msg_sync_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupMessageRangeHash.ProtoReflect.Descriptor instead.
func (*GroupMessageRangeHash) Descriptor() ([]byte, []int) {
	return file_proto_group_msg_sync_proto_rawDescGZIP(), []int{2}
}

func (x *GroupMessageRangeHash) GetGroupId() string {
	if x != nil {
		return x.GroupId
	}
	return ""
}

func (x *GroupMessageRangeHash) GetStartId() string {
	if x != nil {
		return x.StartId
	}
	return ""
}

func (x *GroupMessageRangeHash) GetEndId() string {
	if x != nil {
		return x.EndId
	}
	return ""
}

func (x *GroupMessageRangeHash) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

// 第三步，区间消息ID
type GroupMessageRangeIDs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupId string   `protobuf:"bytes,1,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
	StartId string   `protobuf:"bytes,2,opt,name=start_id,json=startId,proto3" json:"start_id,omitempty"`
	EndId   string   `protobuf:"bytes,3,opt,name=end_id,json=endId,proto3" json:"end_id,omitempty"`
	Ids     []string `protobuf:"bytes,4,rep,name=ids,proto3" json:"ids,omitempty"`
}

func (x *GroupMessageRangeIDs) Reset() {
	*x = GroupMessageRangeIDs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_group_msg_sync_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupMessageRangeIDs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupMessageRangeIDs) ProtoMessage() {}

func (x *GroupMessageRangeIDs) ProtoReflect() protoreflect.Message {
	mi := &file_proto_group_msg_sync_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupMessageRangeIDs.ProtoReflect.Descriptor instead.
func (*GroupMessageRangeIDs) Descriptor() ([]byte, []int) {
	return file_proto_group_msg_sync_proto_rawDescGZIP(), []int{3}
}

func (x *GroupMessageRangeIDs) GetGroupId() string {
	if x != nil {
		return x.GroupId
	}
	return ""
}

func (x *GroupMessageRangeIDs) GetStartId() string {
	if x != nil {
		return x.StartId
	}
	return ""
}

func (x *GroupMessageRangeIDs) GetEndId() string {
	if x != nil {
		return x.EndId
	}
	return ""
}

func (x *GroupMessageRangeIDs) GetIds() []string {
	if x != nil {
		return x.Ids
	}
	return nil
}

// 第五步，获取缺失消息
type GroupMessagePullMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ids []string `protobuf:"bytes,1,rep,name=ids,proto3" json:"ids,omitempty"`
}

func (x *GroupMessagePullMsg) Reset() {
	*x = GroupMessagePullMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_group_msg_sync_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupMessagePullMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupMessagePullMsg) ProtoMessage() {}

func (x *GroupMessagePullMsg) ProtoReflect() protoreflect.Message {
	mi := &file_proto_group_msg_sync_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupMessagePullMsg.ProtoReflect.Descriptor instead.
func (*GroupMessagePullMsg) Descriptor() ([]byte, []int) {
	return file_proto_group_msg_sync_proto_rawDescGZIP(), []int{4}
}

func (x *GroupMessagePullMsg) GetIds() []string {
	if x != nil {
		return x.Ids
	}
	return nil
}

var File_proto_group_msg_sync_proto protoreflect.FileDescriptor

var file_proto_group_msg_sync_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x6d, 0x73,
	0x67, 0x5f, 0x73, 0x79, 0x6e, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x73, 0x79,
	0x6e, 0x63, 0x2e, 0x70, 0x62, 0x22, 0xdf, 0x01, 0x0a, 0x10, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x53,
	0x79, 0x6e, 0x63, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x49, 0x64, 0x12, 0x32, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x73, 0x79, 0x6e, 0x63, 0x2e, 0x70, 0x62, 0x2e, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x53, 0x79, 0x6e, 0x63, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x22, 0x62, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x49,
	0x4e, 0x49, 0x54, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x55, 0x4d, 0x4d, 0x41, 0x52, 0x59,
	0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x52, 0x41, 0x4e, 0x47, 0x45, 0x5f, 0x48, 0x41, 0x53, 0x48,
	0x10, 0x02, 0x12, 0x0d, 0x0a, 0x09, 0x52, 0x41, 0x4e, 0x47, 0x45, 0x5f, 0x49, 0x44, 0x53, 0x10,
	0x03, 0x12, 0x0c, 0x0a, 0x08, 0x50, 0x55, 0x53, 0x48, 0x5f, 0x4d, 0x53, 0x47, 0x10, 0x04, 0x12,
	0x0c, 0x0a, 0x08, 0x50, 0x55, 0x4c, 0x4c, 0x5f, 0x4d, 0x53, 0x47, 0x10, 0x05, 0x12, 0x08, 0x0a,
	0x04, 0x44, 0x4f, 0x4e, 0x45, 0x10, 0x06, 0x22, 0xad, 0x01, 0x0a, 0x13, 0x47, 0x72, 0x6f, 0x75,
	0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x12,
	0x19, 0x0a, 0x08, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x68, 0x65,
	0x61, 0x64, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x68, 0x65, 0x61,
	0x64, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x61, 0x69, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x61, 0x69, 0x6c, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06,
	0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6c, 0x65,
	0x6e, 0x67, 0x74, 0x68, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6d, 0x70, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x6c, 0x61, 0x6d, 0x70, 0x74, 0x69, 0x6d, 0x65,
	0x12, 0x15, 0x0a, 0x06, 0x69, 0x73, 0x5f, 0x65, 0x6e, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x05, 0x69, 0x73, 0x45, 0x6e, 0x64, 0x22, 0x78, 0x0a, 0x15, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x48, 0x61, 0x73, 0x68,
	0x12, 0x19, 0x0a, 0x08, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x49, 0x64, 0x12, 0x15, 0x0a, 0x06, 0x65, 0x6e, 0x64, 0x5f, 0x69, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x6e, 0x64, 0x49, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73,
	0x68, 0x22, 0x75, 0x0a, 0x14, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x49, 0x44, 0x73, 0x12, 0x19, 0x0a, 0x08, 0x67, 0x72, 0x6f,
	0x75, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x67, 0x72, 0x6f,
	0x75, 0x70, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74, 0x49, 0x64, 0x12,
	0x15, 0x0a, 0x06, 0x65, 0x6e, 0x64, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x65, 0x6e, 0x64, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x64, 0x73, 0x18, 0x04, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x03, 0x69, 0x64, 0x73, 0x22, 0x27, 0x0a, 0x13, 0x47, 0x72, 0x6f, 0x75,
	0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x50, 0x75, 0x6c, 0x6c, 0x4d, 0x73, 0x67, 0x12,
	0x10, 0x0a, 0x03, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x03, 0x69, 0x64,
	0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_group_msg_sync_proto_rawDescOnce sync.Once
	file_proto_group_msg_sync_proto_rawDescData = file_proto_group_msg_sync_proto_rawDesc
)

func file_proto_group_msg_sync_proto_rawDescGZIP() []byte {
	file_proto_group_msg_sync_proto_rawDescOnce.Do(func() {
		file_proto_group_msg_sync_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_group_msg_sync_proto_rawDescData)
	})
	return file_proto_group_msg_sync_proto_rawDescData
}

var file_proto_group_msg_sync_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_group_msg_sync_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_proto_group_msg_sync_proto_goTypes = []interface{}{
	(GroupSyncMessage_Type)(0),    // 0: sync.pb.GroupSyncMessage.Type
	(*GroupSyncMessage)(nil),      // 1: sync.pb.GroupSyncMessage
	(*GroupMessageSummary)(nil),   // 2: sync.pb.GroupMessageSummary
	(*GroupMessageRangeHash)(nil), // 3: sync.pb.GroupMessageRangeHash
	(*GroupMessageRangeIDs)(nil),  // 4: sync.pb.GroupMessageRangeIDs
	(*GroupMessagePullMsg)(nil),   // 5: sync.pb.GroupMessagePullMsg
}
var file_proto_group_msg_sync_proto_depIdxs = []int32{
	0, // 0: sync.pb.GroupSyncMessage.type:type_name -> sync.pb.GroupSyncMessage.Type
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proto_group_msg_sync_proto_init() }
func file_proto_group_msg_sync_proto_init() {
	if File_proto_group_msg_sync_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_group_msg_sync_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupSyncMessage); i {
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
		file_proto_group_msg_sync_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupMessageSummary); i {
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
		file_proto_group_msg_sync_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupMessageRangeHash); i {
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
		file_proto_group_msg_sync_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupMessageRangeIDs); i {
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
		file_proto_group_msg_sync_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupMessagePullMsg); i {
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
			RawDescriptor: file_proto_group_msg_sync_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_group_msg_sync_proto_goTypes,
		DependencyIndexes: file_proto_group_msg_sync_proto_depIdxs,
		EnumInfos:         file_proto_group_msg_sync_proto_enumTypes,
		MessageInfos:      file_proto_group_msg_sync_proto_msgTypes,
	}.Build()
	File_proto_group_msg_sync_proto = out.File
	file_proto_group_msg_sync_proto_rawDesc = nil
	file_proto_group_msg_sync_proto_goTypes = nil
	file_proto_group_msg_sync_proto_depIdxs = nil
}
