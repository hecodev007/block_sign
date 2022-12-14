// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go impl.26.0
// 	protoc        v3.14.0
// source: xk_grpc.proto

package grpcs

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type ParamRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Method    string            `protobuf:"bytes,1,opt,name=method,proto3" json:"method,omitempty"`
	Params    map[string][]byte `protobuf:"bytes,2,rep,name=params,proto3" json:"params,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	TimeStamp int64             `protobuf:"varint,3,opt,name=time_stamp,json=timeStamp,proto3" json:"time_stamp,omitempty"`
}

func (x *ParamRequest) Reset() {
	*x = ParamRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xk_grpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ParamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParamRequest) ProtoMessage() {}

func (x *ParamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xk_grpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParamRequest.ProtoReflect.Descriptor instead.
func (*ParamRequest) Descriptor() ([]byte, []int) {
	return file_xk_grpc_proto_rawDescGZIP(), []int{0}
}

func (x *ParamRequest) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *ParamRequest) GetParams() map[string][]byte {
	if x != nil {
		return x.Params
	}
	return nil
}

func (x *ParamRequest) GetTimeStamp() int64 {
	if x != nil {
		return x.TimeStamp
	}
	return 0
}

type ParamReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code     int64             `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Msg      string            `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
	RpcReply map[string][]byte `protobuf:"bytes,3,rep,name=rpc_reply,json=rpcReply,proto3" json:"rpc_reply,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *ParamReply) Reset() {
	*x = ParamReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xk_grpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ParamReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParamReply) ProtoMessage() {}

func (x *ParamReply) ProtoReflect() protoreflect.Message {
	mi := &file_xk_grpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParamReply.ProtoReflect.Descriptor instead.
func (*ParamReply) Descriptor() ([]byte, []int) {
	return file_xk_grpc_proto_rawDescGZIP(), []int{1}
}

func (x *ParamReply) GetCode() int64 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *ParamReply) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

func (x *ParamReply) GetRpcReply() map[string][]byte {
	if x != nil {
		return x.RpcReply
	}
	return nil
}

var File_xk_grpc_proto protoreflect.FileDescriptor

var file_xk_grpc_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x78, 0x6b, 0x5f, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x05, 0x67, 0x72, 0x70, 0x63, 0x73, 0x22, 0xb9, 0x01, 0x0a, 0x0c, 0x50, 0x61, 0x72, 0x61, 0x6d,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12,
	0x37, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1f, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x73, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x74, 0x69, 0x6d, 0x65,
	0x5f, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69,
	0x6d, 0x65, 0x53, 0x74, 0x61, 0x6d, 0x70, 0x1a, 0x39, 0x0a, 0x0b, 0x50, 0x61, 0x72, 0x61, 0x6d,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x22, 0xad, 0x01, 0x0a, 0x0a, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x12, 0x3c, 0x0a, 0x09, 0x72, 0x70, 0x63, 0x5f, 0x72,
	0x65, 0x70, 0x6c, 0x79, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x67, 0x72, 0x70,
	0x63, 0x73, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x2e, 0x52, 0x70,
	0x63, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x72, 0x70, 0x63,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x1a, 0x3b, 0x0a, 0x0d, 0x52, 0x70, 0x63, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x32, 0x42, 0x0a, 0x07, 0x47, 0x72, 0x65, 0x65, 0x74, 0x65, 0x72, 0x12, 0x37, 0x0a,
	0x0b, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x13, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x73, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x11, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x73, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x09, 0x5a, 0x07, 0x2f, 0x3b, 0x67, 0x72, 0x70, 0x63,
	0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_xk_grpc_proto_rawDescOnce sync.Once
	file_xk_grpc_proto_rawDescData = file_xk_grpc_proto_rawDesc
)

func file_xk_grpc_proto_rawDescGZIP() []byte {
	file_xk_grpc_proto_rawDescOnce.Do(func() {
		file_xk_grpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_xk_grpc_proto_rawDescData)
	})
	return file_xk_grpc_proto_rawDescData
}

var file_xk_grpc_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_xk_grpc_proto_goTypes = []interface{}{
	(*ParamRequest)(nil), // 0: grpcs.ParamRequest
	(*ParamReply)(nil),   // 1: grpcs.ParamReply
	nil,                  // 2: grpcs.ParamRequest.ParamsEntry
	nil,                  // 3: grpcs.ParamReply.RpcReplyEntry
}
var file_xk_grpc_proto_depIdxs = []int32{
	2, // 0: grpcs.ParamRequest.params:type_name -> grpcs.ParamRequest.ParamsEntry
	3, // 1: grpcs.ParamReply.rpc_reply:type_name -> grpcs.ParamReply.RpcReplyEntry
	0, // 2: grpcs.Greeter.SendRequest:input_type -> grpcs.ParamRequest
	1, // 3: grpcs.Greeter.SendRequest:output_type -> grpcs.ParamReply
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_xk_grpc_proto_init() }
func file_xk_grpc_proto_init() {
	if File_xk_grpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_xk_grpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ParamRequest); i {
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
		file_xk_grpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ParamReply); i {
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
			RawDescriptor: file_xk_grpc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xk_grpc_proto_goTypes,
		DependencyIndexes: file_xk_grpc_proto_depIdxs,
		MessageInfos:      file_xk_grpc_proto_msgTypes,
	}.Build()
	File_xk_grpc_proto = out.File
	file_xk_grpc_proto_rawDesc = nil
	file_xk_grpc_proto_goTypes = nil
	file_xk_grpc_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// GreeterClient is the client API for Greeter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GreeterClient interface {
	SendRequest(ctx context.Context, in *ParamRequest, opts ...grpc.CallOption) (*ParamReply, error)
}

type greeterClient struct {
	cc grpc.ClientConnInterface
}

func NewGreeterClient(cc grpc.ClientConnInterface) GreeterClient {
	return &greeterClient{cc}
}

func (c *greeterClient) SendRequest(ctx context.Context, in *ParamRequest, opts ...grpc.CallOption) (*ParamReply, error) {
	out := new(ParamReply)
	err := c.cc.Invoke(ctx, "/grpcs.Greeter/SendRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GreeterServer is the server API for Greeter service.
type GreeterServer interface {
	SendRequest(context.Context, *ParamRequest) (*ParamReply, error)
}

// UnimplementedGreeterServer can be embedded to have forward compatible implementations.
type UnimplementedGreeterServer struct {
}

func (*UnimplementedGreeterServer) SendRequest(context.Context, *ParamRequest) (*ParamReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendRequest not implemented")
}

func RegisterGreeterServer(s *grpc.Server, srv GreeterServer) {
	s.RegisterService(&_Greeter_serviceDesc, srv)
}

func _Greeter_SendRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ParamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterServer).SendRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcs.Greeter/SendRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).SendRequest(ctx, req.(*ParamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Greeter_serviceDesc = grpc.ServiceDesc{
	ServiceName: "grpcs.Greeter",
	HandlerType: (*GreeterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendRequest",
			Handler:    _Greeter_SendRequest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "xk_grpc.proto",
}
