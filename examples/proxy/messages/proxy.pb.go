// Code generated by protoc-gen-go. DO NOT EDIT.
// source: messages/proxy.proto

package messages

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ID struct {
	PublicKey            []byte   `protobuf:"bytes,1,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	Address              string   `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ID) Reset()         { *m = ID{} }
func (m *ID) String() string { return proto.CompactTextString(m) }
func (*ID) ProtoMessage()    {}
func (*ID) Descriptor() ([]byte, []int) {
	return fileDescriptor_proxy_269bdaf01f61e907, []int{0}
}
func (m *ID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ID.Unmarshal(m, b)
}
func (m *ID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ID.Marshal(b, m, deterministic)
}
func (dst *ID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ID.Merge(dst, src)
}
func (m *ID) XXX_Size() int {
	return xxx_messageInfo_ID.Size(m)
}
func (m *ID) XXX_DiscardUnknown() {
	xxx_messageInfo_ID.DiscardUnknown(m)
}

var xxx_messageInfo_ID proto.InternalMessageInfo

func (m *ID) GetPublicKey() []byte {
	if m != nil {
		return m.PublicKey
	}
	return nil
}

func (m *ID) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type ProxyMessage struct {
	Message              string   `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Destination          *ID      `protobuf:"bytes,2,opt,name=destination,proto3" json:"destination,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProxyMessage) Reset()         { *m = ProxyMessage{} }
func (m *ProxyMessage) String() string { return proto.CompactTextString(m) }
func (*ProxyMessage) ProtoMessage()    {}
func (*ProxyMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_proxy_269bdaf01f61e907, []int{1}
}
func (m *ProxyMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProxyMessage.Unmarshal(m, b)
}
func (m *ProxyMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProxyMessage.Marshal(b, m, deterministic)
}
func (dst *ProxyMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProxyMessage.Merge(dst, src)
}
func (m *ProxyMessage) XXX_Size() int {
	return xxx_messageInfo_ProxyMessage.Size(m)
}
func (m *ProxyMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_ProxyMessage.DiscardUnknown(m)
}

var xxx_messageInfo_ProxyMessage proto.InternalMessageInfo

func (m *ProxyMessage) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *ProxyMessage) GetDestination() *ID {
	if m != nil {
		return m.Destination
	}
	return nil
}

func init() {
	proto.RegisterType((*ID)(nil), "messages.ID")
	proto.RegisterType((*ProxyMessage)(nil), "messages.ProxyMessage")
}

func init() { proto.RegisterFile("messages/proxy.proto", fileDescriptor_proxy_269bdaf01f61e907) }

var fileDescriptor_proxy_269bdaf01f61e907 = []byte{
	// 158 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xc9, 0x4d, 0x2d, 0x2e,
	0x4e, 0x4c, 0x4f, 0x2d, 0xd6, 0x2f, 0x28, 0xca, 0xaf, 0xa8, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9,
	0x17, 0xe2, 0x80, 0x89, 0x2a, 0xd9, 0x72, 0x31, 0x79, 0xba, 0x08, 0xc9, 0x72, 0x71, 0x15, 0x94,
	0x26, 0xe5, 0x64, 0x26, 0xc7, 0x67, 0xa7, 0x56, 0x4a, 0x30, 0x2a, 0x30, 0x6a, 0xf0, 0x04, 0x71,
	0x42, 0x44, 0xbc, 0x53, 0x2b, 0x85, 0x24, 0xb8, 0xd8, 0x13, 0x53, 0x52, 0x8a, 0x52, 0x8b, 0x8b,
	0x25, 0x98, 0x14, 0x18, 0x35, 0x38, 0x83, 0x60, 0x5c, 0xa5, 0x08, 0x2e, 0x9e, 0x00, 0x90, 0xb9,
	0xbe, 0x10, 0xf3, 0x40, 0x2a, 0xa1, 0x46, 0x83, 0x4d, 0xe1, 0x0c, 0x82, 0x71, 0x85, 0xf4, 0xb8,
	0xb8, 0x53, 0x52, 0x8b, 0x4b, 0x32, 0xf3, 0x12, 0x4b, 0x32, 0xf3, 0xf3, 0xc0, 0xe6, 0x70, 0x1b,
	0xf1, 0xe8, 0xc1, 0x1c, 0xa2, 0xe7, 0xe9, 0x12, 0x84, 0xac, 0x20, 0x89, 0x0d, 0xec, 0x52, 0x63,
	0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x14, 0x8d, 0x1e, 0xff, 0xc1, 0x00, 0x00, 0x00,
}