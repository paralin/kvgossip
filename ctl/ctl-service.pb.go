// Code generated by protoc-gen-gogo.
// source: github.com/fuserobotics/kvgossip/ctl/ctl-service.proto
// DO NOT EDIT!

/*
Package ctl is a generated protocol buffer package.

It is generated from these files:
	github.com/fuserobotics/kvgossip/ctl/ctl-service.proto

It has these top-level messages:
	PutGrantRequest
	PutGrantResponse
	PutRevocationRequest
	PutRevocationResponse
	BuildTransactionRequest
	BuildTransactionResponse
	PutTransactionRequest
	PutTransactionResponse
	GetGrantsRequest
	GetGrantsResponse
	GetKeyRequest
	GetKeyResponse
*/
package ctl

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import data "github.com/fuserobotics/kvgossip/data"
import tx "github.com/fuserobotics/kvgossip/tx"
import grant "github.com/fuserobotics/kvgossip/grant"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.GoGoProtoPackageIsVersion1

// Store a new grant in the DB
type PutGrantRequest struct {
	Pool *grant.GrantAuthorizationPool `protobuf:"bytes,1,opt,name=pool" json:"pool,omitempty"`
}

func (m *PutGrantRequest) Reset()                    { *m = PutGrantRequest{} }
func (m *PutGrantRequest) String() string            { return proto.CompactTextString(m) }
func (*PutGrantRequest) ProtoMessage()               {}
func (*PutGrantRequest) Descriptor() ([]byte, []int) { return fileDescriptorCtlService, []int{0} }

func (m *PutGrantRequest) GetPool() *grant.GrantAuthorizationPool {
	if m != nil {
		return m.Pool
	}
	return nil
}

type PutGrantResponse struct {
	Revocations []*data.SignedData `protobuf:"bytes,1,rep,name=revocations" json:"revocations,omitempty"`
}

func (m *PutGrantResponse) Reset()                    { *m = PutGrantResponse{} }
func (m *PutGrantResponse) String() string            { return proto.CompactTextString(m) }
func (*PutGrantResponse) ProtoMessage()               {}
func (*PutGrantResponse) Descriptor() ([]byte, []int) { return fileDescriptorCtlService, []int{1} }

func (m *PutGrantResponse) GetRevocations() []*data.SignedData {
	if m != nil {
		return m.Revocations
	}
	return nil
}

type PutRevocationRequest struct {
	Revocation *data.SignedData `protobuf:"bytes,1,opt,name=revocation" json:"revocation,omitempty"`
}

func (m *PutRevocationRequest) Reset()                    { *m = PutRevocationRequest{} }
func (m *PutRevocationRequest) String() string            { return proto.CompactTextString(m) }
func (*PutRevocationRequest) ProtoMessage()               {}
func (*PutRevocationRequest) Descriptor() ([]byte, []int) { return fileDescriptorCtlService, []int{2} }

func (m *PutRevocationRequest) GetRevocation() *data.SignedData {
	if m != nil {
		return m.Revocation
	}
	return nil
}

type PutRevocationResponse struct {
}

func (m *PutRevocationResponse) Reset()                    { *m = PutRevocationResponse{} }
func (m *PutRevocationResponse) String() string            { return proto.CompactTextString(m) }
func (*PutRevocationResponse) ProtoMessage()               {}
func (*PutRevocationResponse) Descriptor() ([]byte, []int) { return fileDescriptorCtlService, []int{3} }

// Request a pool of grants that would satisfy a request.
type BuildTransactionRequest struct {
	EntityPublicKey []byte `protobuf:"bytes,1,opt,name=entity_public_key,json=entityPublicKey,proto3" json:"entity_public_key,omitempty"`
	Key             string `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
}

func (m *BuildTransactionRequest) Reset()         { *m = BuildTransactionRequest{} }
func (m *BuildTransactionRequest) String() string { return proto.CompactTextString(m) }
func (*BuildTransactionRequest) ProtoMessage()    {}
func (*BuildTransactionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorCtlService, []int{4}
}

type BuildTransactionResponse struct {
	Transaction *tx.Transaction    `protobuf:"bytes,1,opt,name=transaction" json:"transaction,omitempty"`
	Revocations []*data.SignedData `protobuf:"bytes,2,rep,name=revocations" json:"revocations,omitempty"`
	Invalid     []*data.SignedData `protobuf:"bytes,3,rep,name=invalid" json:"invalid,omitempty"`
}

func (m *BuildTransactionResponse) Reset()         { *m = BuildTransactionResponse{} }
func (m *BuildTransactionResponse) String() string { return proto.CompactTextString(m) }
func (*BuildTransactionResponse) ProtoMessage()    {}
func (*BuildTransactionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorCtlService, []int{5}
}

func (m *BuildTransactionResponse) GetTransaction() *tx.Transaction {
	if m != nil {
		return m.Transaction
	}
	return nil
}

func (m *BuildTransactionResponse) GetRevocations() []*data.SignedData {
	if m != nil {
		return m.Revocations
	}
	return nil
}

func (m *BuildTransactionResponse) GetInvalid() []*data.SignedData {
	if m != nil {
		return m.Invalid
	}
	return nil
}

type PutTransactionRequest struct {
	Transaction *tx.Transaction `protobuf:"bytes,1,opt,name=transaction" json:"transaction,omitempty"`
}

func (m *PutTransactionRequest) Reset()                    { *m = PutTransactionRequest{} }
func (m *PutTransactionRequest) String() string            { return proto.CompactTextString(m) }
func (*PutTransactionRequest) ProtoMessage()               {}
func (*PutTransactionRequest) Descriptor() ([]byte, []int) { return fileDescriptorCtlService, []int{6} }

func (m *PutTransactionRequest) GetTransaction() *tx.Transaction {
	if m != nil {
		return m.Transaction
	}
	return nil
}

type PutTransactionResponse struct {
}

func (m *PutTransactionResponse) Reset()                    { *m = PutTransactionResponse{} }
func (m *PutTransactionResponse) String() string            { return proto.CompactTextString(m) }
func (*PutTransactionResponse) ProtoMessage()               {}
func (*PutTransactionResponse) Descriptor() ([]byte, []int) { return fileDescriptorCtlService, []int{7} }

type GetGrantsRequest struct {
}

func (m *GetGrantsRequest) Reset()                    { *m = GetGrantsRequest{} }
func (m *GetGrantsRequest) String() string            { return proto.CompactTextString(m) }
func (*GetGrantsRequest) ProtoMessage()               {}
func (*GetGrantsRequest) Descriptor() ([]byte, []int) { return fileDescriptorCtlService, []int{8} }

type GetGrantsResponse struct {
	Grants []*data.SignedData `protobuf:"bytes,1,rep,name=grants" json:"grants,omitempty"`
}

func (m *GetGrantsResponse) Reset()                    { *m = GetGrantsResponse{} }
func (m *GetGrantsResponse) String() string            { return proto.CompactTextString(m) }
func (*GetGrantsResponse) ProtoMessage()               {}
func (*GetGrantsResponse) Descriptor() ([]byte, []int) { return fileDescriptorCtlService, []int{9} }

func (m *GetGrantsResponse) GetGrants() []*data.SignedData {
	if m != nil {
		return m.Grants
	}
	return nil
}

type GetKeyRequest struct {
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (m *GetKeyRequest) Reset()                    { *m = GetKeyRequest{} }
func (m *GetKeyRequest) String() string            { return proto.CompactTextString(m) }
func (*GetKeyRequest) ProtoMessage()               {}
func (*GetKeyRequest) Descriptor() ([]byte, []int) { return fileDescriptorCtlService, []int{10} }

type GetKeyResponse struct {
	Transaction *tx.Transaction `protobuf:"bytes,1,opt,name=transaction" json:"transaction,omitempty"`
}

func (m *GetKeyResponse) Reset()                    { *m = GetKeyResponse{} }
func (m *GetKeyResponse) String() string            { return proto.CompactTextString(m) }
func (*GetKeyResponse) ProtoMessage()               {}
func (*GetKeyResponse) Descriptor() ([]byte, []int) { return fileDescriptorCtlService, []int{11} }

func (m *GetKeyResponse) GetTransaction() *tx.Transaction {
	if m != nil {
		return m.Transaction
	}
	return nil
}

func init() {
	proto.RegisterType((*PutGrantRequest)(nil), "ctl.PutGrantRequest")
	proto.RegisterType((*PutGrantResponse)(nil), "ctl.PutGrantResponse")
	proto.RegisterType((*PutRevocationRequest)(nil), "ctl.PutRevocationRequest")
	proto.RegisterType((*PutRevocationResponse)(nil), "ctl.PutRevocationResponse")
	proto.RegisterType((*BuildTransactionRequest)(nil), "ctl.BuildTransactionRequest")
	proto.RegisterType((*BuildTransactionResponse)(nil), "ctl.BuildTransactionResponse")
	proto.RegisterType((*PutTransactionRequest)(nil), "ctl.PutTransactionRequest")
	proto.RegisterType((*PutTransactionResponse)(nil), "ctl.PutTransactionResponse")
	proto.RegisterType((*GetGrantsRequest)(nil), "ctl.GetGrantsRequest")
	proto.RegisterType((*GetGrantsResponse)(nil), "ctl.GetGrantsResponse")
	proto.RegisterType((*GetKeyRequest)(nil), "ctl.GetKeyRequest")
	proto.RegisterType((*GetKeyResponse)(nil), "ctl.GetKeyResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for ControlService service

type ControlServiceClient interface {
	PutGrant(ctx context.Context, in *PutGrantRequest, opts ...grpc.CallOption) (*PutGrantResponse, error)
	PutRevocation(ctx context.Context, in *PutRevocationRequest, opts ...grpc.CallOption) (*PutRevocationResponse, error)
	BuildTransaction(ctx context.Context, in *BuildTransactionRequest, opts ...grpc.CallOption) (*BuildTransactionResponse, error)
	PutTransaction(ctx context.Context, in *PutTransactionRequest, opts ...grpc.CallOption) (*PutTransactionResponse, error)
	GetGrants(ctx context.Context, in *GetGrantsRequest, opts ...grpc.CallOption) (*GetGrantsResponse, error)
	GetKey(ctx context.Context, in *GetKeyRequest, opts ...grpc.CallOption) (*GetKeyResponse, error)
}

type controlServiceClient struct {
	cc *grpc.ClientConn
}

func NewControlServiceClient(cc *grpc.ClientConn) ControlServiceClient {
	return &controlServiceClient{cc}
}

func (c *controlServiceClient) PutGrant(ctx context.Context, in *PutGrantRequest, opts ...grpc.CallOption) (*PutGrantResponse, error) {
	out := new(PutGrantResponse)
	err := grpc.Invoke(ctx, "/ctl.ControlService/PutGrant", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) PutRevocation(ctx context.Context, in *PutRevocationRequest, opts ...grpc.CallOption) (*PutRevocationResponse, error) {
	out := new(PutRevocationResponse)
	err := grpc.Invoke(ctx, "/ctl.ControlService/PutRevocation", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) BuildTransaction(ctx context.Context, in *BuildTransactionRequest, opts ...grpc.CallOption) (*BuildTransactionResponse, error) {
	out := new(BuildTransactionResponse)
	err := grpc.Invoke(ctx, "/ctl.ControlService/BuildTransaction", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) PutTransaction(ctx context.Context, in *PutTransactionRequest, opts ...grpc.CallOption) (*PutTransactionResponse, error) {
	out := new(PutTransactionResponse)
	err := grpc.Invoke(ctx, "/ctl.ControlService/PutTransaction", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) GetGrants(ctx context.Context, in *GetGrantsRequest, opts ...grpc.CallOption) (*GetGrantsResponse, error) {
	out := new(GetGrantsResponse)
	err := grpc.Invoke(ctx, "/ctl.ControlService/GetGrants", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) GetKey(ctx context.Context, in *GetKeyRequest, opts ...grpc.CallOption) (*GetKeyResponse, error) {
	out := new(GetKeyResponse)
	err := grpc.Invoke(ctx, "/ctl.ControlService/GetKey", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ControlService service

type ControlServiceServer interface {
	PutGrant(context.Context, *PutGrantRequest) (*PutGrantResponse, error)
	PutRevocation(context.Context, *PutRevocationRequest) (*PutRevocationResponse, error)
	BuildTransaction(context.Context, *BuildTransactionRequest) (*BuildTransactionResponse, error)
	PutTransaction(context.Context, *PutTransactionRequest) (*PutTransactionResponse, error)
	GetGrants(context.Context, *GetGrantsRequest) (*GetGrantsResponse, error)
	GetKey(context.Context, *GetKeyRequest) (*GetKeyResponse, error)
}

func RegisterControlServiceServer(s *grpc.Server, srv ControlServiceServer) {
	s.RegisterService(&_ControlService_serviceDesc, srv)
}

func _ControlService_PutGrant_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutGrantRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).PutGrant(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ctl.ControlService/PutGrant",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).PutGrant(ctx, req.(*PutGrantRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_PutRevocation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRevocationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).PutRevocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ctl.ControlService/PutRevocation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).PutRevocation(ctx, req.(*PutRevocationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_BuildTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuildTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).BuildTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ctl.ControlService/BuildTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).BuildTransaction(ctx, req.(*BuildTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_PutTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).PutTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ctl.ControlService/PutTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).PutTransaction(ctx, req.(*PutTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_GetGrants_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGrantsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).GetGrants(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ctl.ControlService/GetGrants",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).GetGrants(ctx, req.(*GetGrantsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_GetKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetKeyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).GetKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ctl.ControlService/GetKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).GetKey(ctx, req.(*GetKeyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ControlService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ctl.ControlService",
	HandlerType: (*ControlServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PutGrant",
			Handler:    _ControlService_PutGrant_Handler,
		},
		{
			MethodName: "PutRevocation",
			Handler:    _ControlService_PutRevocation_Handler,
		},
		{
			MethodName: "BuildTransaction",
			Handler:    _ControlService_BuildTransaction_Handler,
		},
		{
			MethodName: "PutTransaction",
			Handler:    _ControlService_PutTransaction_Handler,
		},
		{
			MethodName: "GetGrants",
			Handler:    _ControlService_GetGrants_Handler,
		},
		{
			MethodName: "GetKey",
			Handler:    _ControlService_GetKey_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptorCtlService,
}

var fileDescriptorCtlService = []byte{
	// 538 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x54, 0x5d, 0x6f, 0xd3, 0x30,
	0x14, 0xa5, 0x0b, 0x2a, 0xec, 0x96, 0xb5, 0x99, 0xd9, 0x47, 0x09, 0x4c, 0x02, 0x3f, 0x4d, 0x48,
	0xa4, 0x5b, 0x27, 0x90, 0x90, 0xe0, 0x01, 0x36, 0xb1, 0x89, 0xbd, 0x94, 0x0c, 0x89, 0xc7, 0x29,
	0x4d, 0x4d, 0x67, 0x2d, 0xc4, 0x21, 0x76, 0xaa, 0x96, 0x5f, 0xc4, 0x0b, 0xff, 0x11, 0xc7, 0x76,
	0x3e, 0x9a, 0xb6, 0x2a, 0xec, 0xa1, 0x51, 0x74, 0xef, 0x39, 0xc7, 0x37, 0xe7, 0x1e, 0x17, 0xde,
	0x8c, 0xa9, 0xb8, 0x49, 0x87, 0x6e, 0xc0, 0x7e, 0xf4, 0xbe, 0xa7, 0x9c, 0x24, 0x6c, 0xc8, 0x04,
	0x0d, 0x78, 0xef, 0x76, 0x32, 0x66, 0x9c, 0xd3, 0xb8, 0x17, 0x88, 0x30, 0xfb, 0xbd, 0x92, 0xbd,
	0x09, 0x0d, 0x88, 0x1b, 0x27, 0x4c, 0x30, 0x64, 0xc9, 0x92, 0x73, 0xb4, 0x96, 0x3c, 0xf2, 0x85,
	0xaf, 0x1e, 0x9a, 0xe6, 0xbc, 0x5e, 0xcb, 0x10, 0xd3, 0x9e, 0x48, 0xfc, 0x88, 0xfb, 0x81, 0xa0,
	0x2c, 0x32, 0xb4, 0xfe, 0x5a, 0xda, 0x58, 0x72, 0x84, 0x7e, 0x6a, 0x0e, 0x3e, 0x83, 0xce, 0x20,
	0x15, 0xe7, 0x59, 0xc5, 0x23, 0x3f, 0x53, 0xc2, 0x05, 0x3a, 0x86, 0xfb, 0x31, 0x63, 0x61, 0xb7,
	0xf1, 0xbc, 0x71, 0xd8, 0xea, 0x1f, 0xb8, 0x1a, 0xae, 0x20, 0x1f, 0x52, 0x71, 0xc3, 0x12, 0xfa,
	0xcb, 0xcf, 0x4e, 0x1d, 0x48, 0x90, 0xa7, 0xa0, 0xf8, 0x13, 0xd8, 0xa5, 0x0a, 0x8f, 0x59, 0xc4,
	0x09, 0xea, 0x43, 0x2b, 0x21, 0x13, 0x16, 0x28, 0x2c, 0x97, 0x6a, 0x96, 0x54, 0xb3, 0x5d, 0xf5,
	0x99, 0x57, 0x74, 0x1c, 0x91, 0xd1, 0x99, 0x7c, 0xf5, 0xaa, 0x20, 0x7c, 0x01, 0x3b, 0x52, 0xc7,
	0x2b, 0x2a, 0xf9, 0x48, 0x47, 0x00, 0x25, 0xcc, 0x0c, 0xb6, 0x28, 0x55, 0xc1, 0xe0, 0x7d, 0xd8,
	0xad, 0x29, 0xe9, 0xb1, 0xf0, 0x37, 0xd8, 0xff, 0x98, 0xd2, 0x70, 0xf4, 0xb5, 0xb4, 0x2f, 0x3f,
	0xe5, 0x25, 0x6c, 0x93, 0x48, 0x50, 0x31, 0xbb, 0x8e, 0xd3, 0x61, 0x48, 0x83, 0xeb, 0x5b, 0x32,
	0x53, 0x87, 0x3d, 0xf2, 0x3a, 0xba, 0x31, 0x50, 0xf5, 0x4b, 0x32, 0x43, 0x36, 0x58, 0x59, 0x77,
	0x43, 0x76, 0x37, 0xbd, 0xec, 0x15, 0xff, 0x69, 0x40, 0x77, 0x51, 0xd9, 0x98, 0x71, 0x0c, 0xad,
	0xca, 0xbe, 0xcc, 0x17, 0x74, 0x5c, 0x31, 0x75, 0xab, 0xe8, 0x2a, 0xa6, 0xee, 0xdf, 0xc6, 0x3f,
	0xf8, 0x27, 0xbf, 0xe0, 0x01, 0x8d, 0x26, 0x7e, 0x48, 0x47, 0x5d, 0x6b, 0x05, 0x3e, 0x07, 0xe0,
	0xcf, 0xca, 0xa1, 0x25, 0x36, 0xfc, 0xff, 0xac, 0xb8, 0x0b, 0x7b, 0x75, 0x2d, 0x63, 0x37, 0x02,
	0xfb, 0x9c, 0xe8, 0x64, 0x70, 0x73, 0x00, 0x7e, 0x0f, 0xdb, 0x95, 0x9a, 0x71, 0xe8, 0x10, 0x9a,
	0x2a, 0x68, 0xab, 0x93, 0x62, 0xfa, 0xf8, 0x05, 0x6c, 0x49, 0xba, 0x5c, 0x42, 0x3e, 0xb0, 0xd9,
	0x45, 0xa3, 0xdc, 0xc5, 0x29, 0xb4, 0x73, 0xc8, 0x9d, 0x17, 0xd0, 0xff, 0x6d, 0x41, 0xfb, 0x94,
	0x45, 0x22, 0x61, 0xe1, 0x95, 0xbe, 0xd5, 0xe8, 0x2d, 0x3c, 0xcc, 0x73, 0x8e, 0x76, 0x5c, 0x79,
	0xb9, 0xdd, 0xda, 0xe5, 0x71, 0x76, 0x6b, 0x55, 0x63, 0xc3, 0x3d, 0x74, 0x01, 0x5b, 0x73, 0x81,
	0x44, 0x4f, 0x72, 0xe4, 0x42, 0xdc, 0x1d, 0x67, 0x59, 0xab, 0x50, 0xfa, 0x02, 0x76, 0x3d, 0x67,
	0xe8, 0x99, 0x62, 0xac, 0x08, 0xb6, 0x73, 0xb0, 0xa2, 0x5b, 0x48, 0x5e, 0x42, 0x7b, 0x7e, 0x7f,
	0xa8, 0x18, 0x61, 0x89, 0xdc, 0xd3, 0xa5, 0xbd, 0x42, 0xec, 0x1d, 0x6c, 0x16, 0xeb, 0x45, 0xda,
	0x8f, 0x7a, 0x04, 0x9c, 0xbd, 0x7a, 0xb9, 0x60, 0x9f, 0x40, 0x53, 0xaf, 0x0e, 0xa1, 0x1c, 0x53,
	0xae, 0xda, 0x79, 0x3c, 0x57, 0xcb, 0x49, 0xc3, 0xa6, 0xfa, 0x33, 0x3b, 0xf9, 0x1b, 0x00, 0x00,
	0xff, 0xff, 0x9b, 0xe8, 0x65, 0x9f, 0xa8, 0x05, 0x00, 0x00,
}
