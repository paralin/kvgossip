// Code generated by protoc-gen-go.
// source: github.com/fuserobotics/kvgossip/tx/transaction.proto
// DO NOT EDIT!

/*
Package tx is a generated protocol buffer package.

It is generated from these files:
	github.com/fuserobotics/kvgossip/tx/transaction.proto

It has these top-level messages:
	Transaction
	TransactionValue
	TransactionVerification
*/
package tx

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import grant "github.com/fuserobotics/kvgossip/grant"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Transaction_TransactionType int32

const (
	Transaction_TRANSACTION_SET Transaction_TransactionType = 0
)

var Transaction_TransactionType_name = map[int32]string{
	0: "TRANSACTION_SET",
}
var Transaction_TransactionType_value = map[string]int32{
	"TRANSACTION_SET": 0,
}

func (x Transaction_TransactionType) String() string {
	return proto.EnumName(Transaction_TransactionType_name, int32(x))
}
func (Transaction_TransactionType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{0, 0}
}

// A transaction.
type Transaction struct {
	Key             string                      `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Value           []byte                      `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Verification    *TransactionVerification    `protobuf:"bytes,3,opt,name=verification" json:"verification,omitempty"`
	TransactionType Transaction_TransactionType `protobuf:"varint,4,opt,name=transaction_type,json=transactionType,enum=tx.Transaction_TransactionType" json:"transaction_type,omitempty"`
}

func (m *Transaction) Reset()                    { *m = Transaction{} }
func (m *Transaction) String() string            { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()               {}
func (*Transaction) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Transaction) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *Transaction) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *Transaction) GetVerification() *TransactionVerification {
	if m != nil {
		return m.Verification
	}
	return nil
}

func (m *Transaction) GetTransactionType() Transaction_TransactionType {
	if m != nil {
		return m.TransactionType
	}
	return Transaction_TRANSACTION_SET
}

type TransactionValue struct {
}

func (m *TransactionValue) Reset()                    { *m = TransactionValue{} }
func (m *TransactionValue) String() string            { return proto.CompactTextString(m) }
func (*TransactionValue) ProtoMessage()               {}
func (*TransactionValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type TransactionVerification struct {
	// Signature of the value object + timestamp (64 bit network order)
	ValueSignature []byte `protobuf:"bytes,1,opt,name=value_signature,json=valueSignature,proto3" json:"value_signature,omitempty"`
	// Public key of the signer.
	SignerPublicKey []byte `protobuf:"bytes,2,opt,name=signer_public_key,json=signerPublicKey,proto3" json:"signer_public_key,omitempty"`
	// Grant authorization
	Grant *grant.GrantAuthorizationPool `protobuf:"bytes,3,opt,name=grant" json:"grant,omitempty"`
	// Timestamp
	Timestamp uint64 `protobuf:"varint,4,opt,name=timestamp" json:"timestamp,omitempty"`
}

func (m *TransactionVerification) Reset()                    { *m = TransactionVerification{} }
func (m *TransactionVerification) String() string            { return proto.CompactTextString(m) }
func (*TransactionVerification) ProtoMessage()               {}
func (*TransactionVerification) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *TransactionVerification) GetValueSignature() []byte {
	if m != nil {
		return m.ValueSignature
	}
	return nil
}

func (m *TransactionVerification) GetSignerPublicKey() []byte {
	if m != nil {
		return m.SignerPublicKey
	}
	return nil
}

func (m *TransactionVerification) GetGrant() *grant.GrantAuthorizationPool {
	if m != nil {
		return m.Grant
	}
	return nil
}

func (m *TransactionVerification) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func init() {
	proto.RegisterType((*Transaction)(nil), "tx.Transaction")
	proto.RegisterType((*TransactionValue)(nil), "tx.TransactionValue")
	proto.RegisterType((*TransactionVerification)(nil), "tx.TransactionVerification")
	proto.RegisterEnum("tx.Transaction_TransactionType", Transaction_TransactionType_name, Transaction_TransactionType_value)
}

func init() {
	proto.RegisterFile("github.com/fuserobotics/kvgossip/tx/transaction.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 347 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x84, 0x51, 0x4d, 0x4b, 0xc3, 0x40,
	0x10, 0x35, 0xfd, 0x10, 0xba, 0x2d, 0x4d, 0x5c, 0x05, 0x83, 0x1f, 0x18, 0x72, 0xd0, 0xe0, 0x21,
	0x81, 0x16, 0xcf, 0x52, 0x44, 0x44, 0x85, 0x5a, 0xb6, 0xc1, 0x6b, 0x48, 0xc2, 0x36, 0x5d, 0xda,
	0x66, 0xc3, 0x66, 0x52, 0x1a, 0xff, 0x9d, 0x7f, 0xc7, 0x5f, 0x21, 0xd9, 0x55, 0x9a, 0x16, 0xc4,
	0xcb, 0x30, 0x79, 0x79, 0x6f, 0x76, 0xe6, 0x3d, 0x74, 0x97, 0x30, 0x98, 0x17, 0x91, 0x1b, 0xf3,
	0x95, 0x37, 0x2b, 0x72, 0x2a, 0x78, 0xc4, 0x81, 0xc5, 0xb9, 0xb7, 0x58, 0x27, 0x3c, 0xcf, 0x59,
	0xe6, 0xc1, 0xc6, 0x03, 0x11, 0xa6, 0x79, 0x18, 0x03, 0xe3, 0xa9, 0x9b, 0x09, 0x0e, 0x1c, 0x37,
	0x60, 0x73, 0x36, 0xf8, 0x57, 0x9a, 0x88, 0x30, 0x05, 0x55, 0x95, 0xce, 0xfe, 0xd2, 0x50, 0xd7,
	0xdf, 0x4e, 0xc3, 0x06, 0x6a, 0x2e, 0x68, 0x69, 0x6a, 0x96, 0xe6, 0x74, 0x48, 0xd5, 0xe2, 0x13,
	0xd4, 0x5e, 0x87, 0xcb, 0x82, 0x9a, 0x0d, 0x4b, 0x73, 0x7a, 0x44, 0x7d, 0xe0, 0x7b, 0xd4, 0x5b,
	0x53, 0xc1, 0x66, 0x2c, 0x0e, 0x2b, 0x9d, 0xd9, 0xb4, 0x34, 0xa7, 0x3b, 0x38, 0x77, 0x61, 0xe3,
	0xd6, 0xc6, 0xbd, 0xd7, 0x28, 0x64, 0x47, 0x80, 0x5f, 0x90, 0x51, 0xbb, 0x22, 0x80, 0x32, 0xa3,
	0x66, 0xcb, 0xd2, 0x9c, 0xfe, 0xe0, 0x6a, 0x6f, 0x48, 0xbd, 0xf7, 0xcb, 0x8c, 0x12, 0x1d, 0x76,
	0x01, 0xfb, 0x1a, 0xe9, 0x7b, 0x1c, 0x7c, 0x8c, 0x74, 0x9f, 0x8c, 0xc6, 0xd3, 0xd1, 0x83, 0xff,
	0xfc, 0x36, 0x0e, 0xa6, 0x8f, 0xbe, 0x71, 0x60, 0x63, 0x64, 0xd4, 0x97, 0xab, 0x0e, 0xb1, 0x3f,
	0x35, 0x74, 0xfa, 0xc7, 0xc6, 0xf8, 0x06, 0xe9, 0xf2, 0xda, 0x20, 0x67, 0x49, 0x1a, 0x42, 0x21,
	0xa8, 0x34, 0xa6, 0x47, 0xfa, 0x12, 0x9e, 0xfe, 0xa2, 0xf8, 0x16, 0x1d, 0x55, 0x14, 0x2a, 0x82,
	0xac, 0x88, 0x96, 0x2c, 0x0e, 0x2a, 0x0f, 0x95, 0x5f, 0xba, 0xfa, 0x31, 0x91, 0xf8, 0x2b, 0x2d,
	0xf1, 0x10, 0xb5, 0x65, 0x00, 0x3f, 0x96, 0x5d, 0xba, 0x2a, 0x8e, 0xa7, 0xaa, 0x8e, 0x0a, 0x98,
	0x73, 0xc1, 0x3e, 0xe4, 0xf3, 0x13, 0xce, 0x97, 0x44, 0x71, 0xf1, 0x05, 0xea, 0x00, 0x5b, 0xd1,
	0x1c, 0xc2, 0x55, 0x26, 0x6d, 0x6a, 0x91, 0x2d, 0x10, 0x1d, 0xca, 0x2c, 0x87, 0xdf, 0x01, 0x00,
	0x00, 0xff, 0xff, 0x57, 0xb4, 0xce, 0xac, 0x3c, 0x02, 0x00, 0x00,
}
