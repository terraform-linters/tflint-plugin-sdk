// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: tflint.proto

package proto

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

// RuleSetClient is the client API for RuleSet service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RuleSetClient interface {
	GetName(ctx context.Context, in *GetName_Request, opts ...grpc.CallOption) (*GetName_Response, error)
	GetVersion(ctx context.Context, in *GetVersion_Request, opts ...grpc.CallOption) (*GetVersion_Response, error)
	GetVersionConstraint(ctx context.Context, in *GetVersionConstraint_Request, opts ...grpc.CallOption) (*GetVersionConstraint_Response, error)
	GetSDKVersion(ctx context.Context, in *GetSDKVersion_Request, opts ...grpc.CallOption) (*GetSDKVersion_Response, error)
	GetRuleNames(ctx context.Context, in *GetRuleNames_Request, opts ...grpc.CallOption) (*GetRuleNames_Response, error)
	GetConfigSchema(ctx context.Context, in *GetConfigSchema_Request, opts ...grpc.CallOption) (*GetConfigSchema_Response, error)
	ApplyGlobalConfig(ctx context.Context, in *ApplyGlobalConfig_Request, opts ...grpc.CallOption) (*ApplyGlobalConfig_Response, error)
	ApplyConfig(ctx context.Context, in *ApplyConfig_Request, opts ...grpc.CallOption) (*ApplyConfig_Response, error)
	Check(ctx context.Context, in *Check_Request, opts ...grpc.CallOption) (*Check_Response, error)
}

type ruleSetClient struct {
	cc grpc.ClientConnInterface
}

func NewRuleSetClient(cc grpc.ClientConnInterface) RuleSetClient {
	return &ruleSetClient{cc}
}

func (c *ruleSetClient) GetName(ctx context.Context, in *GetName_Request, opts ...grpc.CallOption) (*GetName_Response, error) {
	out := new(GetName_Response)
	err := c.cc.Invoke(ctx, "/proto.RuleSet/GetName", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ruleSetClient) GetVersion(ctx context.Context, in *GetVersion_Request, opts ...grpc.CallOption) (*GetVersion_Response, error) {
	out := new(GetVersion_Response)
	err := c.cc.Invoke(ctx, "/proto.RuleSet/GetVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ruleSetClient) GetVersionConstraint(ctx context.Context, in *GetVersionConstraint_Request, opts ...grpc.CallOption) (*GetVersionConstraint_Response, error) {
	out := new(GetVersionConstraint_Response)
	err := c.cc.Invoke(ctx, "/proto.RuleSet/GetVersionConstraint", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ruleSetClient) GetSDKVersion(ctx context.Context, in *GetSDKVersion_Request, opts ...grpc.CallOption) (*GetSDKVersion_Response, error) {
	out := new(GetSDKVersion_Response)
	err := c.cc.Invoke(ctx, "/proto.RuleSet/GetSDKVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ruleSetClient) GetRuleNames(ctx context.Context, in *GetRuleNames_Request, opts ...grpc.CallOption) (*GetRuleNames_Response, error) {
	out := new(GetRuleNames_Response)
	err := c.cc.Invoke(ctx, "/proto.RuleSet/GetRuleNames", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ruleSetClient) GetConfigSchema(ctx context.Context, in *GetConfigSchema_Request, opts ...grpc.CallOption) (*GetConfigSchema_Response, error) {
	out := new(GetConfigSchema_Response)
	err := c.cc.Invoke(ctx, "/proto.RuleSet/GetConfigSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ruleSetClient) ApplyGlobalConfig(ctx context.Context, in *ApplyGlobalConfig_Request, opts ...grpc.CallOption) (*ApplyGlobalConfig_Response, error) {
	out := new(ApplyGlobalConfig_Response)
	err := c.cc.Invoke(ctx, "/proto.RuleSet/ApplyGlobalConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ruleSetClient) ApplyConfig(ctx context.Context, in *ApplyConfig_Request, opts ...grpc.CallOption) (*ApplyConfig_Response, error) {
	out := new(ApplyConfig_Response)
	err := c.cc.Invoke(ctx, "/proto.RuleSet/ApplyConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ruleSetClient) Check(ctx context.Context, in *Check_Request, opts ...grpc.CallOption) (*Check_Response, error) {
	out := new(Check_Response)
	err := c.cc.Invoke(ctx, "/proto.RuleSet/Check", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RuleSetServer is the server API for RuleSet service.
// All implementations must embed UnimplementedRuleSetServer
// for forward compatibility
type RuleSetServer interface {
	GetName(context.Context, *GetName_Request) (*GetName_Response, error)
	GetVersion(context.Context, *GetVersion_Request) (*GetVersion_Response, error)
	GetVersionConstraint(context.Context, *GetVersionConstraint_Request) (*GetVersionConstraint_Response, error)
	GetSDKVersion(context.Context, *GetSDKVersion_Request) (*GetSDKVersion_Response, error)
	GetRuleNames(context.Context, *GetRuleNames_Request) (*GetRuleNames_Response, error)
	GetConfigSchema(context.Context, *GetConfigSchema_Request) (*GetConfigSchema_Response, error)
	ApplyGlobalConfig(context.Context, *ApplyGlobalConfig_Request) (*ApplyGlobalConfig_Response, error)
	ApplyConfig(context.Context, *ApplyConfig_Request) (*ApplyConfig_Response, error)
	Check(context.Context, *Check_Request) (*Check_Response, error)
	mustEmbedUnimplementedRuleSetServer()
}

// UnimplementedRuleSetServer must be embedded to have forward compatible implementations.
type UnimplementedRuleSetServer struct {
}

func (UnimplementedRuleSetServer) GetName(context.Context, *GetName_Request) (*GetName_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetName not implemented")
}
func (UnimplementedRuleSetServer) GetVersion(context.Context, *GetVersion_Request) (*GetVersion_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVersion not implemented")
}
func (UnimplementedRuleSetServer) GetVersionConstraint(context.Context, *GetVersionConstraint_Request) (*GetVersionConstraint_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVersionConstraint not implemented")
}
func (UnimplementedRuleSetServer) GetSDKVersion(context.Context, *GetSDKVersion_Request) (*GetSDKVersion_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSDKVersion not implemented")
}
func (UnimplementedRuleSetServer) GetRuleNames(context.Context, *GetRuleNames_Request) (*GetRuleNames_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRuleNames not implemented")
}
func (UnimplementedRuleSetServer) GetConfigSchema(context.Context, *GetConfigSchema_Request) (*GetConfigSchema_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConfigSchema not implemented")
}
func (UnimplementedRuleSetServer) ApplyGlobalConfig(context.Context, *ApplyGlobalConfig_Request) (*ApplyGlobalConfig_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplyGlobalConfig not implemented")
}
func (UnimplementedRuleSetServer) ApplyConfig(context.Context, *ApplyConfig_Request) (*ApplyConfig_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplyConfig not implemented")
}
func (UnimplementedRuleSetServer) Check(context.Context, *Check_Request) (*Check_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Check not implemented")
}
func (UnimplementedRuleSetServer) mustEmbedUnimplementedRuleSetServer() {}

// UnsafeRuleSetServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RuleSetServer will
// result in compilation errors.
type UnsafeRuleSetServer interface {
	mustEmbedUnimplementedRuleSetServer()
}

func RegisterRuleSetServer(s grpc.ServiceRegistrar, srv RuleSetServer) {
	s.RegisterService(&RuleSet_ServiceDesc, srv)
}

func _RuleSet_GetName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetName_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuleSetServer).GetName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RuleSet/GetName",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuleSetServer).GetName(ctx, req.(*GetName_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuleSet_GetVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVersion_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuleSetServer).GetVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RuleSet/GetVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuleSetServer).GetVersion(ctx, req.(*GetVersion_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuleSet_GetVersionConstraint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVersionConstraint_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuleSetServer).GetVersionConstraint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RuleSet/GetVersionConstraint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuleSetServer).GetVersionConstraint(ctx, req.(*GetVersionConstraint_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuleSet_GetSDKVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSDKVersion_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuleSetServer).GetSDKVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RuleSet/GetSDKVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuleSetServer).GetSDKVersion(ctx, req.(*GetSDKVersion_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuleSet_GetRuleNames_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRuleNames_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuleSetServer).GetRuleNames(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RuleSet/GetRuleNames",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuleSetServer).GetRuleNames(ctx, req.(*GetRuleNames_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuleSet_GetConfigSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConfigSchema_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuleSetServer).GetConfigSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RuleSet/GetConfigSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuleSetServer).GetConfigSchema(ctx, req.(*GetConfigSchema_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuleSet_ApplyGlobalConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApplyGlobalConfig_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuleSetServer).ApplyGlobalConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RuleSet/ApplyGlobalConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuleSetServer).ApplyGlobalConfig(ctx, req.(*ApplyGlobalConfig_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuleSet_ApplyConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApplyConfig_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuleSetServer).ApplyConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RuleSet/ApplyConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuleSetServer).ApplyConfig(ctx, req.(*ApplyConfig_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _RuleSet_Check_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Check_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RuleSetServer).Check(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RuleSet/Check",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RuleSetServer).Check(ctx, req.(*Check_Request))
	}
	return interceptor(ctx, in, info, handler)
}

// RuleSet_ServiceDesc is the grpc.ServiceDesc for RuleSet service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RuleSet_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.RuleSet",
	HandlerType: (*RuleSetServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetName",
			Handler:    _RuleSet_GetName_Handler,
		},
		{
			MethodName: "GetVersion",
			Handler:    _RuleSet_GetVersion_Handler,
		},
		{
			MethodName: "GetVersionConstraint",
			Handler:    _RuleSet_GetVersionConstraint_Handler,
		},
		{
			MethodName: "GetSDKVersion",
			Handler:    _RuleSet_GetSDKVersion_Handler,
		},
		{
			MethodName: "GetRuleNames",
			Handler:    _RuleSet_GetRuleNames_Handler,
		},
		{
			MethodName: "GetConfigSchema",
			Handler:    _RuleSet_GetConfigSchema_Handler,
		},
		{
			MethodName: "ApplyGlobalConfig",
			Handler:    _RuleSet_ApplyGlobalConfig_Handler,
		},
		{
			MethodName: "ApplyConfig",
			Handler:    _RuleSet_ApplyConfig_Handler,
		},
		{
			MethodName: "Check",
			Handler:    _RuleSet_Check_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tflint.proto",
}

// RunnerClient is the client API for Runner service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RunnerClient interface {
	GetModulePath(ctx context.Context, in *GetModulePath_Request, opts ...grpc.CallOption) (*GetModulePath_Response, error)
	GetModuleContent(ctx context.Context, in *GetModuleContent_Request, opts ...grpc.CallOption) (*GetModuleContent_Response, error)
	GetFile(ctx context.Context, in *GetFile_Request, opts ...grpc.CallOption) (*GetFile_Response, error)
	GetFiles(ctx context.Context, in *GetFiles_Request, opts ...grpc.CallOption) (*GetFiles_Response, error)
	GetRuleConfigContent(ctx context.Context, in *GetRuleConfigContent_Request, opts ...grpc.CallOption) (*GetRuleConfigContent_Response, error)
	EvaluateExpr(ctx context.Context, in *EvaluateExpr_Request, opts ...grpc.CallOption) (*EvaluateExpr_Response, error)
	EmitIssue(ctx context.Context, in *EmitIssue_Request, opts ...grpc.CallOption) (*EmitIssue_Response, error)
}

type runnerClient struct {
	cc grpc.ClientConnInterface
}

func NewRunnerClient(cc grpc.ClientConnInterface) RunnerClient {
	return &runnerClient{cc}
}

func (c *runnerClient) GetModulePath(ctx context.Context, in *GetModulePath_Request, opts ...grpc.CallOption) (*GetModulePath_Response, error) {
	out := new(GetModulePath_Response)
	err := c.cc.Invoke(ctx, "/proto.Runner/GetModulePath", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runnerClient) GetModuleContent(ctx context.Context, in *GetModuleContent_Request, opts ...grpc.CallOption) (*GetModuleContent_Response, error) {
	out := new(GetModuleContent_Response)
	err := c.cc.Invoke(ctx, "/proto.Runner/GetModuleContent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runnerClient) GetFile(ctx context.Context, in *GetFile_Request, opts ...grpc.CallOption) (*GetFile_Response, error) {
	out := new(GetFile_Response)
	err := c.cc.Invoke(ctx, "/proto.Runner/GetFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runnerClient) GetFiles(ctx context.Context, in *GetFiles_Request, opts ...grpc.CallOption) (*GetFiles_Response, error) {
	out := new(GetFiles_Response)
	err := c.cc.Invoke(ctx, "/proto.Runner/GetFiles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runnerClient) GetRuleConfigContent(ctx context.Context, in *GetRuleConfigContent_Request, opts ...grpc.CallOption) (*GetRuleConfigContent_Response, error) {
	out := new(GetRuleConfigContent_Response)
	err := c.cc.Invoke(ctx, "/proto.Runner/GetRuleConfigContent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runnerClient) EvaluateExpr(ctx context.Context, in *EvaluateExpr_Request, opts ...grpc.CallOption) (*EvaluateExpr_Response, error) {
	out := new(EvaluateExpr_Response)
	err := c.cc.Invoke(ctx, "/proto.Runner/EvaluateExpr", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runnerClient) EmitIssue(ctx context.Context, in *EmitIssue_Request, opts ...grpc.CallOption) (*EmitIssue_Response, error) {
	out := new(EmitIssue_Response)
	err := c.cc.Invoke(ctx, "/proto.Runner/EmitIssue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RunnerServer is the server API for Runner service.
// All implementations must embed UnimplementedRunnerServer
// for forward compatibility
type RunnerServer interface {
	GetModulePath(context.Context, *GetModulePath_Request) (*GetModulePath_Response, error)
	GetModuleContent(context.Context, *GetModuleContent_Request) (*GetModuleContent_Response, error)
	GetFile(context.Context, *GetFile_Request) (*GetFile_Response, error)
	GetFiles(context.Context, *GetFiles_Request) (*GetFiles_Response, error)
	GetRuleConfigContent(context.Context, *GetRuleConfigContent_Request) (*GetRuleConfigContent_Response, error)
	EvaluateExpr(context.Context, *EvaluateExpr_Request) (*EvaluateExpr_Response, error)
	EmitIssue(context.Context, *EmitIssue_Request) (*EmitIssue_Response, error)
	mustEmbedUnimplementedRunnerServer()
}

// UnimplementedRunnerServer must be embedded to have forward compatible implementations.
type UnimplementedRunnerServer struct {
}

func (UnimplementedRunnerServer) GetModulePath(context.Context, *GetModulePath_Request) (*GetModulePath_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModulePath not implemented")
}
func (UnimplementedRunnerServer) GetModuleContent(context.Context, *GetModuleContent_Request) (*GetModuleContent_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModuleContent not implemented")
}
func (UnimplementedRunnerServer) GetFile(context.Context, *GetFile_Request) (*GetFile_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFile not implemented")
}
func (UnimplementedRunnerServer) GetFiles(context.Context, *GetFiles_Request) (*GetFiles_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFiles not implemented")
}
func (UnimplementedRunnerServer) GetRuleConfigContent(context.Context, *GetRuleConfigContent_Request) (*GetRuleConfigContent_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRuleConfigContent not implemented")
}
func (UnimplementedRunnerServer) EvaluateExpr(context.Context, *EvaluateExpr_Request) (*EvaluateExpr_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EvaluateExpr not implemented")
}
func (UnimplementedRunnerServer) EmitIssue(context.Context, *EmitIssue_Request) (*EmitIssue_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EmitIssue not implemented")
}
func (UnimplementedRunnerServer) mustEmbedUnimplementedRunnerServer() {}

// UnsafeRunnerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RunnerServer will
// result in compilation errors.
type UnsafeRunnerServer interface {
	mustEmbedUnimplementedRunnerServer()
}

func RegisterRunnerServer(s grpc.ServiceRegistrar, srv RunnerServer) {
	s.RegisterService(&Runner_ServiceDesc, srv)
}

func _Runner_GetModulePath_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetModulePath_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerServer).GetModulePath(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Runner/GetModulePath",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerServer).GetModulePath(ctx, req.(*GetModulePath_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Runner_GetModuleContent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetModuleContent_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerServer).GetModuleContent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Runner/GetModuleContent",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerServer).GetModuleContent(ctx, req.(*GetModuleContent_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Runner_GetFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFile_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerServer).GetFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Runner/GetFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerServer).GetFile(ctx, req.(*GetFile_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Runner_GetFiles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFiles_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerServer).GetFiles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Runner/GetFiles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerServer).GetFiles(ctx, req.(*GetFiles_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Runner_GetRuleConfigContent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRuleConfigContent_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerServer).GetRuleConfigContent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Runner/GetRuleConfigContent",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerServer).GetRuleConfigContent(ctx, req.(*GetRuleConfigContent_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Runner_EvaluateExpr_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EvaluateExpr_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerServer).EvaluateExpr(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Runner/EvaluateExpr",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerServer).EvaluateExpr(ctx, req.(*EvaluateExpr_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Runner_EmitIssue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmitIssue_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerServer).EmitIssue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Runner/EmitIssue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerServer).EmitIssue(ctx, req.(*EmitIssue_Request))
	}
	return interceptor(ctx, in, info, handler)
}

// Runner_ServiceDesc is the grpc.ServiceDesc for Runner service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Runner_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Runner",
	HandlerType: (*RunnerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetModulePath",
			Handler:    _Runner_GetModulePath_Handler,
		},
		{
			MethodName: "GetModuleContent",
			Handler:    _Runner_GetModuleContent_Handler,
		},
		{
			MethodName: "GetFile",
			Handler:    _Runner_GetFile_Handler,
		},
		{
			MethodName: "GetFiles",
			Handler:    _Runner_GetFiles_Handler,
		},
		{
			MethodName: "GetRuleConfigContent",
			Handler:    _Runner_GetRuleConfigContent_Handler,
		},
		{
			MethodName: "EvaluateExpr",
			Handler:    _Runner_EvaluateExpr_Handler,
		},
		{
			MethodName: "EmitIssue",
			Handler:    _Runner_EmitIssue_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tflint.proto",
}
