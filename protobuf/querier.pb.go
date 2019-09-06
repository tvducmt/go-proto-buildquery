// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: querier.proto

package querier

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type FieldQuery struct {
	Query                *string  `protobuf:"bytes,1,opt,name=query" json:"query,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FieldQuery) Reset()         { *m = FieldQuery{} }
func (m *FieldQuery) String() string { return proto.CompactTextString(m) }
func (*FieldQuery) ProtoMessage()    {}
func (*FieldQuery) Descriptor() ([]byte, []int) {
	return fileDescriptor_7edfe438abd6b96f, []int{0}
}
func (m *FieldQuery) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FieldQuery.Unmarshal(m, b)
}
func (m *FieldQuery) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FieldQuery.Marshal(b, m, deterministic)
}
func (m *FieldQuery) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FieldQuery.Merge(m, src)
}
func (m *FieldQuery) XXX_Size() int {
	return xxx_messageInfo_FieldQuery.Size(m)
}
func (m *FieldQuery) XXX_DiscardUnknown() {
	xxx_messageInfo_FieldQuery.DiscardUnknown(m)
}

var xxx_messageInfo_FieldQuery proto.InternalMessageInfo

func (m *FieldQuery) GetQuery() string {
	if m != nil && m.Query != nil {
		return *m.Query
	}
	return ""
}

var E_Field = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FieldOptions)(nil),
	ExtensionType: (*FieldQuery)(nil),
	Field:         65020,
	Name:          "querier.field",
	Tag:           "bytes,65020,opt,name=field",
	Filename:      "querier.proto",
}

func init() {
	proto.RegisterType((*FieldQuery)(nil), "querier.FieldQuery")
	proto.RegisterExtension(E_Field)
}

func init() { proto.RegisterFile("querier.proto", fileDescriptor_7edfe438abd6b96f) }

var fileDescriptor_7edfe438abd6b96f = []byte{
	// 148 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x2c, 0x4d, 0x2d,
	0xca, 0x4c, 0x2d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x87, 0x72, 0xa5, 0x14, 0xd2,
	0xf3, 0xf3, 0xd3, 0x73, 0x52, 0xf5, 0xc1, 0xc2, 0x49, 0xa5, 0x69, 0xfa, 0x29, 0xa9, 0xc5, 0xc9,
	0x45, 0x99, 0x05, 0x25, 0xf9, 0x50, 0xa5, 0x4a, 0x4a, 0x5c, 0x5c, 0x6e, 0x99, 0xa9, 0x39, 0x29,
	0x81, 0xa5, 0xa9, 0x45, 0x95, 0x42, 0x22, 0x5c, 0xac, 0x20, 0xad, 0x95, 0x12, 0x8c, 0x0a, 0x8c,
	0x1a, 0x9c, 0x41, 0x10, 0x8e, 0x95, 0x17, 0x17, 0x6b, 0x1a, 0x48, 0x8d, 0x90, 0xac, 0x1e, 0xc4,
	0x3c, 0x3d, 0x98, 0x79, 0x7a, 0x60, 0xbd, 0xfe, 0x05, 0x25, 0x99, 0xf9, 0x79, 0xc5, 0x12, 0x7f,
	0x7e, 0x33, 0x2b, 0x30, 0x6a, 0x70, 0x1b, 0x09, 0xeb, 0xc1, 0x9c, 0x83, 0x30, 0x3a, 0x08, 0x62,
	0x84, 0x13, 0x67, 0x14, 0xcc, 0x71, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x26, 0xe4, 0x2f, 0x64,
	0xb5, 0x00, 0x00, 0x00,
}
