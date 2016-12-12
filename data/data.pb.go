// Code generated by protoc-gen-go.
// source: github.com/fuserobotics/kvgossip/data/data.proto
// DO NOT EDIT!

/*
Package data is a generated protocol buffer package.

It is generated from these files:
	github.com/fuserobotics/kvgossip/data/data.proto

It has these top-level messages:
	SignedData
*/
package data

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

type SignedData_SignedDataType int32

const (
	SignedData_SIGNED_GRANT            SignedData_SignedDataType = 0
	SignedData_SIGNED_GRANT_REVOCATION SignedData_SignedDataType = 2
)

var SignedData_SignedDataType_name = map[int32]string{
	0: "SIGNED_GRANT",
	2: "SIGNED_GRANT_REVOCATION",
}
var SignedData_SignedDataType_value = map[string]int32{
	"SIGNED_GRANT":            0,
	"SIGNED_GRANT_REVOCATION": 2,
}

func (x SignedData_SignedDataType) String() string {
	return proto.EnumName(SignedData_SignedDataType_name, int32(x))
}
func (SignedData_SignedDataType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type SignedData struct {
	BodyType  SignedData_SignedDataType `protobuf:"varint,1,opt,name=body_type,json=bodyType,enum=data.SignedData_SignedDataType" json:"body_type,omitempty"`
	Body      []byte                    `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
	Signature []byte                    `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (m *SignedData) Reset()                    { *m = SignedData{} }
func (m *SignedData) String() string            { return proto.CompactTextString(m) }
func (*SignedData) ProtoMessage()               {}
func (*SignedData) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *SignedData) GetBodyType() SignedData_SignedDataType {
	if m != nil {
		return m.BodyType
	}
	return SignedData_SIGNED_GRANT
}

func (m *SignedData) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

func (m *SignedData) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func init() {
	proto.RegisterType((*SignedData)(nil), "data.SignedData")
	proto.RegisterEnum("data.SignedData_SignedDataType", SignedData_SignedDataType_name, SignedData_SignedDataType_value)
}

func init() { proto.RegisterFile("github.com/fuserobotics/kvgossip/data/data.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 209 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x32, 0x48, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0x2b, 0x2d, 0x4e, 0x2d, 0xca, 0x4f, 0xca, 0x2f,
	0xc9, 0x4c, 0x2e, 0xd6, 0xcf, 0x2e, 0x4b, 0xcf, 0x2f, 0x2e, 0xce, 0x2c, 0xd0, 0x4f, 0x49, 0x2c,
	0x49, 0x04, 0x13, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42, 0x2c, 0x20, 0xb6, 0xd2, 0x5e, 0x46,
	0x2e, 0xae, 0xe0, 0xcc, 0xf4, 0xbc, 0xd4, 0x14, 0x97, 0xc4, 0x92, 0x44, 0x21, 0x1b, 0x2e, 0xce,
	0xa4, 0xfc, 0x94, 0xca, 0xf8, 0x92, 0xca, 0x82, 0x54, 0x09, 0x46, 0x05, 0x46, 0x0d, 0x3e, 0x23,
	0x79, 0x3d, 0xb0, 0x26, 0x84, 0x22, 0x24, 0x66, 0x48, 0x65, 0x41, 0x6a, 0x10, 0x07, 0x48, 0x07,
	0x88, 0x25, 0x24, 0xc4, 0xc5, 0x02, 0x62, 0x4b, 0x30, 0x29, 0x30, 0x6a, 0xf0, 0x04, 0x81, 0xd9,
	0x42, 0x32, 0x5c, 0x9c, 0xc5, 0x99, 0xe9, 0x79, 0x89, 0x25, 0xa5, 0x45, 0xa9, 0x12, 0xcc, 0x60,
	0x09, 0x84, 0x80, 0x92, 0x3d, 0x17, 0x1f, 0xaa, 0x69, 0x42, 0x02, 0x5c, 0x3c, 0xc1, 0x9e, 0xee,
	0x7e, 0xae, 0x2e, 0xf1, 0xee, 0x41, 0x8e, 0x7e, 0x21, 0x02, 0x0c, 0x42, 0xd2, 0x5c, 0xe2, 0xc8,
	0x22, 0xf1, 0x41, 0xae, 0x61, 0xfe, 0xce, 0x8e, 0x21, 0x9e, 0xfe, 0x7e, 0x02, 0x4c, 0x49, 0x6c,
	0x60, 0xcf, 0x18, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x30, 0x21, 0x25, 0x9d, 0x00, 0x01, 0x00,
	0x00,
}
