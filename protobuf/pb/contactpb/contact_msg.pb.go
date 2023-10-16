// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: proto/contact_msg.proto

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

type ContactMessage_SendState int32

const (
	ContactMessage_Sending  ContactMessage_SendState = 0
	ContactMessage_SendSucc ContactMessage_SendState = 1
	ContactMessage_SendFail ContactMessage_SendState = 2
)

// Enum value maps for ContactMessage_SendState.
var (
	ContactMessage_SendState_name = map[int32]string{
		0: "Sending",
		1: "SendSucc",
		2: "SendFail",
	}
	ContactMessage_SendState_value = map[string]int32{
		"Sending":  0,
		"SendSucc": 1,
		"SendFail": 2,
	}
)

func (x ContactMessage_SendState) Enum() *ContactMessage_SendState {
	p := new(ContactMessage_SendState)
	*p = x
	return p
}

func (x ContactMessage_SendState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ContactMessage_SendState) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_contact_msg_proto_enumTypes[0].Descriptor()
}

func (ContactMessage_SendState) Type() protoreflect.EnumType {
	return &file_proto_contact_msg_proto_enumTypes[0]
}

func (x ContactMessage_SendState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ContactMessage_SendState.Descriptor instead.
func (ContactMessage_SendState) EnumDescriptor() ([]byte, []int) {
	return file_proto_contact_msg_proto_rawDescGZIP(), []int{1, 0}
}

type CoreMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	FromPeerId   []byte `protobuf:"bytes,2,opt,name=fromPeerId,proto3" json:"fromPeerId,omitempty"`
	ToPeerId     []byte `protobuf:"bytes,3,opt,name=toPeerId,proto3" json:"toPeerId,omitempty"`
	MsgType      string `protobuf:"bytes,4,opt,name=msgType,proto3" json:"msgType,omitempty"`
	MimeType     string `protobuf:"bytes,5,opt,name=mimeType,proto3" json:"mimeType,omitempty"`
	Payload      []byte `protobuf:"bytes,6,opt,name=payload,proto3" json:"payload,omitempty"`
	AttachmentId string `protobuf:"bytes,7,opt,name=attachmentId,proto3" json:"attachmentId,omitempty"`
	Lamportime   uint64 `protobuf:"varint,8,opt,name=lamportime,proto3" json:"lamportime,omitempty"`
	Timestamp    int64  `protobuf:"varint,9,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Signature    []byte `protobuf:"bytes,10,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *CoreMessage) Reset() {
	*x = CoreMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_msg_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CoreMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CoreMessage) ProtoMessage() {}

func (x *CoreMessage) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_msg_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CoreMessage.ProtoReflect.Descriptor instead.
func (*CoreMessage) Descriptor() ([]byte, []int) {
	return file_proto_contact_msg_proto_rawDescGZIP(), []int{0}
}

func (x *CoreMessage) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *CoreMessage) GetFromPeerId() []byte {
	if x != nil {
		return x.FromPeerId
	}
	return nil
}

func (x *CoreMessage) GetToPeerId() []byte {
	if x != nil {
		return x.ToPeerId
	}
	return nil
}

func (x *CoreMessage) GetMsgType() string {
	if x != nil {
		return x.MsgType
	}
	return ""
}

func (x *CoreMessage) GetMimeType() string {
	if x != nil {
		return x.MimeType
	}
	return ""
}

func (x *CoreMessage) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *CoreMessage) GetAttachmentId() string {
	if x != nil {
		return x.AttachmentId
	}
	return ""
}

func (x *CoreMessage) GetLamportime() uint64 {
	if x != nil {
		return x.Lamportime
	}
	return 0
}

func (x *CoreMessage) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *CoreMessage) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type ContactMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string                   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	CoreMessage *CoreMessage             `protobuf:"bytes,2,opt,name=coreMessage,proto3" json:"coreMessage,omitempty"`
	SendState   ContactMessage_SendState `protobuf:"varint,3,opt,name=sendState,proto3,enum=peermsg.pb.ContactMessage_SendState" json:"sendState,omitempty"`
	// directly or deposit
	IsDeposit  bool  `protobuf:"varint,4,opt,name=isDeposit,proto3" json:"isDeposit,omitempty"`
	CreateTime int64 `protobuf:"varint,5,opt,name=createTime,proto3" json:"createTime,omitempty"`
	UpdateTime int64 `protobuf:"varint,6,opt,name=updateTime,proto3" json:"updateTime,omitempty"`
}

func (x *ContactMessage) Reset() {
	*x = ContactMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_msg_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContactMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContactMessage) ProtoMessage() {}

func (x *ContactMessage) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_msg_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContactMessage.ProtoReflect.Descriptor instead.
func (*ContactMessage) Descriptor() ([]byte, []int) {
	return file_proto_contact_msg_proto_rawDescGZIP(), []int{1}
}

func (x *ContactMessage) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ContactMessage) GetCoreMessage() *CoreMessage {
	if x != nil {
		return x.CoreMessage
	}
	return nil
}

func (x *ContactMessage) GetSendState() ContactMessage_SendState {
	if x != nil {
		return x.SendState
	}
	return ContactMessage_Sending
}

func (x *ContactMessage) GetIsDeposit() bool {
	if x != nil {
		return x.IsDeposit
	}
	return false
}

func (x *ContactMessage) GetCreateTime() int64 {
	if x != nil {
		return x.CreateTime
	}
	return 0
}

func (x *ContactMessage) GetUpdateTime() int64 {
	if x != nil {
		return x.UpdateTime
	}
	return 0
}

type MessageEnvelope struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string       `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	CoreMessage *CoreMessage `protobuf:"bytes,2,opt,name=coreMessage,proto3" json:"coreMessage,omitempty"`
	Attachment  []byte       `protobuf:"bytes,3,opt,name=attachment,proto3" json:"attachment,omitempty"`
}

func (x *MessageEnvelope) Reset() {
	*x = MessageEnvelope{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_msg_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageEnvelope) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageEnvelope) ProtoMessage() {}

func (x *MessageEnvelope) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_msg_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageEnvelope.ProtoReflect.Descriptor instead.
func (*MessageEnvelope) Descriptor() ([]byte, []int) {
	return file_proto_contact_msg_proto_rawDescGZIP(), []int{2}
}

func (x *MessageEnvelope) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *MessageEnvelope) GetCoreMessage() *CoreMessage {
	if x != nil {
		return x.CoreMessage
	}
	return nil
}

func (x *MessageEnvelope) GetAttachment() []byte {
	if x != nil {
		return x.Attachment
	}
	return nil
}

type ContactMessageAck struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *ContactMessageAck) Reset() {
	*x = ContactMessageAck{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_contact_msg_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContactMessageAck) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContactMessageAck) ProtoMessage() {}

func (x *ContactMessageAck) ProtoReflect() protoreflect.Message {
	mi := &file_proto_contact_msg_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContactMessageAck.ProtoReflect.Descriptor instead.
func (*ContactMessageAck) Descriptor() ([]byte, []int) {
	return file_proto_contact_msg_proto_rawDescGZIP(), []int{3}
}

func (x *ContactMessageAck) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

var File_proto_contact_msg_proto protoreflect.FileDescriptor

var file_proto_contact_msg_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x5f,
	0x6d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x70, 0x65, 0x65, 0x72, 0x6d,
	0x73, 0x67, 0x2e, 0x70, 0x62, 0x22, 0xa9, 0x02, 0x0a, 0x0b, 0x43, 0x6f, 0x72, 0x65, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x66, 0x72, 0x6f, 0x6d, 0x50, 0x65, 0x65,
	0x72, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x66, 0x72, 0x6f, 0x6d, 0x50,
	0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x74, 0x6f, 0x50, 0x65, 0x65, 0x72, 0x49,
	0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x74, 0x6f, 0x50, 0x65, 0x65, 0x72, 0x49,
	0x64, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6d,
	0x69, 0x6d, 0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6d,
	0x69, 0x6d, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x12, 0x22, 0x0a, 0x0c, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x49,
	0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d,
	0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x6c, 0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x6c, 0x61, 0x6d, 0x70, 0x6f,
	0x72, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x18, 0x09, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x22, 0xb3, 0x02, 0x0a, 0x0e, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x39, 0x0a, 0x0b, 0x63, 0x6f, 0x72, 0x65, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x70, 0x65, 0x65, 0x72,
	0x6d, 0x73, 0x67, 0x2e, 0x70, 0x62, 0x2e, 0x43, 0x6f, 0x72, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x52, 0x0b, 0x63, 0x6f, 0x72, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x42, 0x0a, 0x09, 0x73, 0x65, 0x6e, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x24, 0x2e, 0x70, 0x65, 0x65, 0x72, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x62, 0x2e,
	0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x53,
	0x65, 0x6e, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x09, 0x73, 0x65, 0x6e, 0x64, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x73, 0x44, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x69, 0x73, 0x44, 0x65, 0x70, 0x6f, 0x73, 0x69,
	0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d,
	0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d,
	0x65, 0x22, 0x34, 0x0a, 0x09, 0x53, 0x65, 0x6e, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0b,
	0x0a, 0x07, 0x53, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x53,
	0x65, 0x6e, 0x64, 0x53, 0x75, 0x63, 0x63, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x53, 0x65, 0x6e,
	0x64, 0x46, 0x61, 0x69, 0x6c, 0x10, 0x02, 0x22, 0x7c, 0x0a, 0x0f, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x45, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x39, 0x0a, 0x0b, 0x63, 0x6f,
	0x72, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x17, 0x2e, 0x70, 0x65, 0x65, 0x72, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x62, 0x2e, 0x43, 0x6f, 0x72,
	0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x0b, 0x63, 0x6f, 0x72, 0x65, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d,
	0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x61, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x6d, 0x65, 0x6e, 0x74, 0x22, 0x23, 0x0a, 0x11, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x41, 0x63, 0x6b, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_proto_contact_msg_proto_rawDescOnce sync.Once
	file_proto_contact_msg_proto_rawDescData = file_proto_contact_msg_proto_rawDesc
)

func file_proto_contact_msg_proto_rawDescGZIP() []byte {
	file_proto_contact_msg_proto_rawDescOnce.Do(func() {
		file_proto_contact_msg_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_contact_msg_proto_rawDescData)
	})
	return file_proto_contact_msg_proto_rawDescData
}

var file_proto_contact_msg_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_contact_msg_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proto_contact_msg_proto_goTypes = []interface{}{
	(ContactMessage_SendState)(0), // 0: peermsg.pb.ContactMessage.SendState
	(*CoreMessage)(nil),           // 1: peermsg.pb.CoreMessage
	(*ContactMessage)(nil),        // 2: peermsg.pb.ContactMessage
	(*MessageEnvelope)(nil),       // 3: peermsg.pb.MessageEnvelope
	(*ContactMessageAck)(nil),     // 4: peermsg.pb.ContactMessageAck
}
var file_proto_contact_msg_proto_depIdxs = []int32{
	1, // 0: peermsg.pb.ContactMessage.coreMessage:type_name -> peermsg.pb.CoreMessage
	0, // 1: peermsg.pb.ContactMessage.sendState:type_name -> peermsg.pb.ContactMessage.SendState
	1, // 2: peermsg.pb.MessageEnvelope.coreMessage:type_name -> peermsg.pb.CoreMessage
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_proto_contact_msg_proto_init() }
func file_proto_contact_msg_proto_init() {
	if File_proto_contact_msg_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_contact_msg_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CoreMessage); i {
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
		file_proto_contact_msg_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContactMessage); i {
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
		file_proto_contact_msg_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageEnvelope); i {
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
		file_proto_contact_msg_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContactMessageAck); i {
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
			RawDescriptor: file_proto_contact_msg_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_contact_msg_proto_goTypes,
		DependencyIndexes: file_proto_contact_msg_proto_depIdxs,
		EnumInfos:         file_proto_contact_msg_proto_enumTypes,
		MessageInfos:      file_proto_contact_msg_proto_msgTypes,
	}.Build()
	File_proto_contact_msg_proto = out.File
	file_proto_contact_msg_proto_rawDesc = nil
	file_proto_contact_msg_proto_goTypes = nil
	file_proto_contact_msg_proto_depIdxs = nil
}
