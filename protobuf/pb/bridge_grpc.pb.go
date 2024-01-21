// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.2
// source: protobuf/bridge.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	BridgeService_CrossChainTransfer_FullMethodName   = "/selaginella.proto_rpc.BridgeService/crossChainTransfer"
	BridgeService_ChangeTransferStatus_FullMethodName = "/selaginella.proto_rpc.BridgeService/changeTransferStatus"
)

// BridgeServiceClient is the client API for BridgeService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BridgeServiceClient interface {
	CrossChainTransfer(ctx context.Context, in *CrossChainTransferRequest, opts ...grpc.CallOption) (*CrossChainTransferResponse, error)
	ChangeTransferStatus(ctx context.Context, in *CrossChainTransferStatusRequest, opts ...grpc.CallOption) (*CrossChainTransferStatusResponse, error)
}

type bridgeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBridgeServiceClient(cc grpc.ClientConnInterface) BridgeServiceClient {
	return &bridgeServiceClient{cc}
}

func (c *bridgeServiceClient) CrossChainTransfer(ctx context.Context, in *CrossChainTransferRequest, opts ...grpc.CallOption) (*CrossChainTransferResponse, error) {
	out := new(CrossChainTransferResponse)
	err := c.cc.Invoke(ctx, BridgeService_CrossChainTransfer_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bridgeServiceClient) ChangeTransferStatus(ctx context.Context, in *CrossChainTransferStatusRequest, opts ...grpc.CallOption) (*CrossChainTransferStatusResponse, error) {
	out := new(CrossChainTransferStatusResponse)
	err := c.cc.Invoke(ctx, BridgeService_ChangeTransferStatus_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BridgeServiceServer is the server API for BridgeService service.
// All implementations must embed UnimplementedBridgeServiceServer
// for forward compatibility
type BridgeServiceServer interface {
	CrossChainTransfer(context.Context, *CrossChainTransferRequest) (*CrossChainTransferResponse, error)
	ChangeTransferStatus(context.Context, *CrossChainTransferStatusRequest) (*CrossChainTransferStatusResponse, error)
}

// UnimplementedBridgeServiceServer must be embedded to have forward compatible implementations.
type UnimplementedBridgeServiceServer struct {
}

func (UnimplementedBridgeServiceServer) CrossChainTransfer(context.Context, *CrossChainTransferRequest) (*CrossChainTransferResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CrossChainTransfer not implemented")
}
func (UnimplementedBridgeServiceServer) ChangeTransferStatus(context.Context, *CrossChainTransferStatusRequest) (*CrossChainTransferStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeTransferStatus not implemented")
}

func RegisterBridgeServiceServer(s grpc.ServiceRegistrar, srv BridgeServiceServer) {
	s.RegisterService(&BridgeService_ServiceDesc, srv)
}

func _BridgeService_CrossChainTransfer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CrossChainTransferRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BridgeServiceServer).CrossChainTransfer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BridgeService_CrossChainTransfer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BridgeServiceServer).CrossChainTransfer(ctx, req.(*CrossChainTransferRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BridgeService_ChangeTransferStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CrossChainTransferStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BridgeServiceServer).ChangeTransferStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BridgeService_ChangeTransferStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BridgeServiceServer).ChangeTransferStatus(ctx, req.(*CrossChainTransferStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BridgeService_ServiceDesc is the grpc.ServiceDesc for BridgeService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BridgeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "selaginella.proto_rpc.BridgeService",
	HandlerType: (*BridgeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "crossChainTransfer",
			Handler:    _BridgeService_CrossChainTransfer_Handler,
		},
		{
			MethodName: "changeTransferStatus",
			Handler:    _BridgeService_ChangeTransferStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protobuf/bridge.proto",
}
