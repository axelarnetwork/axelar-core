// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: axelar/permission/v1beta1/service.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	golang_proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = golang_proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

func init() {
	proto.RegisterFile("axelar/permission/v1beta1/service.proto", fileDescriptor_5d763a569c6664cc)
}
func init() {
	golang_proto.RegisterFile("axelar/permission/v1beta1/service.proto", fileDescriptor_5d763a569c6664cc)
}

var fileDescriptor_5d763a569c6664cc = []byte{
	// 402 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0x3f, 0xcb, 0xda, 0x40,
	0x1c, 0xc7, 0xbd, 0xfe, 0x1b, 0x02, 0x5d, 0xae, 0x2e, 0x95, 0x92, 0x21, 0xd0, 0x7f, 0x16, 0x73,
	0xa8, 0x55, 0xc1, 0xb1, 0x2d, 0x74, 0x28, 0x1d, 0x2a, 0xb8, 0x74, 0x91, 0x33, 0xfe, 0xb8, 0x06,
	0xe3, 0xfd, 0xe2, 0xdd, 0xc5, 0x9a, 0xb5, 0xaf, 0xa0, 0xd0, 0x17, 0xd2, 0xa5, 0x50, 0x84, 0x0e,
	0x1d, 0x3b, 0x0a, 0x5d, 0x3a, 0x16, 0xf3, 0xbc, 0x90, 0x07, 0x63, 0xf4, 0x79, 0xc4, 0x44, 0x9e,
	0x6c, 0x09, 0xf9, 0x7c, 0x72, 0x1f, 0xc2, 0x37, 0xd6, 0x53, 0xbe, 0x84, 0x80, 0x2b, 0x16, 0x82,
	0x9a, 0xf9, 0x5a, 0xfb, 0x28, 0xd9, 0xa2, 0x39, 0x06, 0xc3, 0x9b, 0x4c, 0x83, 0x5a, 0xf8, 0x1e,
	0xb8, 0xa1, 0x42, 0x83, 0xf4, 0xe1, 0x0e, 0x74, 0xaf, 0x40, 0x37, 0x03, 0x6b, 0x55, 0x81, 0x02,
	0x53, 0x8a, 0x6d, 0xaf, 0x76, 0x42, 0xed, 0x91, 0x40, 0x14, 0x01, 0x30, 0x1e, 0xfa, 0x8c, 0x4b,
	0x89, 0x86, 0x1b, 0x1f, 0xa5, 0xce, 0x9e, 0x3a, 0xc5, 0xe7, 0x9a, 0x65, 0xc6, 0x3c, 0x2e, 0x66,
	0xe6, 0x11, 0xa8, 0x78, 0x87, 0xb5, 0x7e, 0xde, 0xb1, 0x6e, 0xbf, 0xd7, 0x82, 0xfe, 0x20, 0x16,
	0x1d, 0x80, 0xf0, 0xb5, 0x01, 0xf5, 0x1a, 0xa5, 0x51, 0x18, 0x04, 0xa0, 0xe8, 0x4b, 0xb7, 0xb0,
	0xdc, 0x3d, 0xc5, 0x07, 0x30, 0x8f, 0x40, 0x9b, 0x5a, 0xa7, 0xa4, 0xa5, 0x43, 0x94, 0x1a, 0x9c,
	0xe6, 0x97, 0xbf, 0x17, 0xdf, 0x6e, 0xbd, 0x70, 0x9e, 0xb0, 0xd3, 0x76, 0x95, 0x69, 0x23, 0xef,
	0xe0, 0xf5, 0x49, 0x9d, 0xfe, 0x22, 0x56, 0xf5, 0x0d, 0xa8, 0xd3, 0xf0, 0xee, 0x99, 0x84, 0x3c,
	0x61, 0x9f, 0xde, 0x2b, 0xed, 0x65, 0xf1, 0xed, 0x34, 0xbe, 0xe1, 0x3c, 0xcb, 0x89, 0x9f, 0x40,
	0x41, 0xfe, 0x8a, 0x58, 0x0f, 0x86, 0xe1, 0x84, 0x1b, 0x78, 0x8b, 0x0b, 0x50, 0x92, 0x4b, 0x0f,
	0xde, 0x41, 0x4c, 0xcf, 0x7d, 0xc0, 0x1c, 0x7e, 0x1f, 0xdf, 0x2d, 0xab, 0xdd, 0xa0, 0x3d, 0x4a,
	0xbd, 0x91, 0x38, 0x88, 0xa3, 0x29, 0xc4, 0x7d, 0x52, 0x6f, 0xad, 0x88, 0x75, 0xf7, 0xc3, 0x76,
	0x49, 0xf4, 0x3b, 0xb1, 0xee, 0x1f, 0xf7, 0x9f, 0x9b, 0x4d, 0xea, 0xe4, 0xe6, 0x77, 0x4a, 0x5a,
	0xc7, 0xb3, 0xa1, 0xcf, 0x59, 0xf1, 0xe4, 0x8f, 0xf3, 0x5f, 0x0d, 0xff, 0x6c, 0x6c, 0xb2, 0xde,
	0xd8, 0xe4, 0xff, 0xc6, 0x26, 0x5f, 0x13, 0xbb, 0xf2, 0x3b, 0xb1, 0xc9, 0x3a, 0xb1, 0x2b, 0xff,
	0x12, 0xbb, 0xf2, 0xb1, 0x27, 0x7c, 0xf3, 0x29, 0x1a, 0xbb, 0x1e, 0xce, 0xb2, 0x57, 0x4a, 0x30,
	0x9f, 0x51, 0x4d, 0xb3, 0xbb, 0x86, 0x87, 0x0a, 0xd8, 0xf2, 0xfa, 0x39, 0x26, 0x0e, 0x41, 0x8f,
	0xef, 0xa5, 0xff, 0x54, 0xfb, 0x32, 0x00, 0x00, 0xff, 0xff, 0x74, 0xbc, 0xa5, 0x04, 0x18, 0x04,
	0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	RegisterController(ctx context.Context, in *RegisterControllerRequest, opts ...grpc.CallOption) (*RegisterControllerResponse, error)
	DeregisterController(ctx context.Context, in *DeregisterControllerRequest, opts ...grpc.CallOption) (*DeregisterControllerResponse, error)
	UpdateGovernanceKey(ctx context.Context, in *UpdateGovernanceKeyRequest, opts ...grpc.CallOption) (*UpdateGovernanceKeyResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) RegisterController(ctx context.Context, in *RegisterControllerRequest, opts ...grpc.CallOption) (*RegisterControllerResponse, error) {
	out := new(RegisterControllerResponse)
	err := c.cc.Invoke(ctx, "/axelar.permission.v1beta1.Msg/RegisterController", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) DeregisterController(ctx context.Context, in *DeregisterControllerRequest, opts ...grpc.CallOption) (*DeregisterControllerResponse, error) {
	out := new(DeregisterControllerResponse)
	err := c.cc.Invoke(ctx, "/axelar.permission.v1beta1.Msg/DeregisterController", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateGovernanceKey(ctx context.Context, in *UpdateGovernanceKeyRequest, opts ...grpc.CallOption) (*UpdateGovernanceKeyResponse, error) {
	out := new(UpdateGovernanceKeyResponse)
	err := c.cc.Invoke(ctx, "/axelar.permission.v1beta1.Msg/UpdateGovernanceKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	RegisterController(context.Context, *RegisterControllerRequest) (*RegisterControllerResponse, error)
	DeregisterController(context.Context, *DeregisterControllerRequest) (*DeregisterControllerResponse, error)
	UpdateGovernanceKey(context.Context, *UpdateGovernanceKeyRequest) (*UpdateGovernanceKeyResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) RegisterController(ctx context.Context, req *RegisterControllerRequest) (*RegisterControllerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterController not implemented")
}
func (*UnimplementedMsgServer) DeregisterController(ctx context.Context, req *DeregisterControllerRequest) (*DeregisterControllerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeregisterController not implemented")
}
func (*UnimplementedMsgServer) UpdateGovernanceKey(ctx context.Context, req *UpdateGovernanceKeyRequest) (*UpdateGovernanceKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGovernanceKey not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_RegisterController_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterControllerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RegisterController(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/axelar.permission.v1beta1.Msg/RegisterController",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RegisterController(ctx, req.(*RegisterControllerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_DeregisterController_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeregisterControllerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).DeregisterController(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/axelar.permission.v1beta1.Msg/DeregisterController",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).DeregisterController(ctx, req.(*DeregisterControllerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateGovernanceKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateGovernanceKeyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateGovernanceKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/axelar.permission.v1beta1.Msg/UpdateGovernanceKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateGovernanceKey(ctx, req.(*UpdateGovernanceKeyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "axelar.permission.v1beta1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterController",
			Handler:    _Msg_RegisterController_Handler,
		},
		{
			MethodName: "DeregisterController",
			Handler:    _Msg_DeregisterController_Handler,
		},
		{
			MethodName: "UpdateGovernanceKey",
			Handler:    _Msg_UpdateGovernanceKey_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "axelar/permission/v1beta1/service.proto",
}

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	// GovernanceKey returns the multisig governance key
	GovernanceKey(ctx context.Context, in *QueryGovernanceKeyRequest, opts ...grpc.CallOption) (*QueryGovernanceKeyResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) GovernanceKey(ctx context.Context, in *QueryGovernanceKeyRequest, opts ...grpc.CallOption) (*QueryGovernanceKeyResponse, error) {
	out := new(QueryGovernanceKeyResponse)
	err := c.cc.Invoke(ctx, "/axelar.permission.v1beta1.Query/GovernanceKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// GovernanceKey returns the multisig governance key
	GovernanceKey(context.Context, *QueryGovernanceKeyRequest) (*QueryGovernanceKeyResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) GovernanceKey(ctx context.Context, req *QueryGovernanceKeyRequest) (*QueryGovernanceKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GovernanceKey not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_GovernanceKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryGovernanceKeyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GovernanceKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/axelar.permission.v1beta1.Query/GovernanceKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GovernanceKey(ctx, req.(*QueryGovernanceKeyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "axelar.permission.v1beta1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GovernanceKey",
			Handler:    _Query_GovernanceKey_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "axelar/permission/v1beta1/service.proto",
}