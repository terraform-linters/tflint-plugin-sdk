package host2plugin

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/logger"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/fromproto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/interceptor"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/plugin2host"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/toproto"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer is a plugin-side implementation. Plugin must implement a server that returns a response for a request from host.
// The behavior as gRPC server is implemented in the SDK, and the actual behavior is delegated to impl.
type GRPCServer struct {
	proto.UnimplementedRuleSetServer

	impl   tflint.RuleSet
	broker *plugin.GRPCBroker
}

var _ proto.RuleSetServer = &GRPCServer{}

// ServeOpts is an option for serving a plugin.
// Each plugin can pass a RuleSet that represents its own functionality.
type ServeOpts struct {
	RuleSet tflint.RuleSet
}

// Serve is a wrapper of plugin.Serve. This is entrypoint of all plugins.
func Serve(opts *ServeOpts) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"ruleset": &RuleSetPlugin{impl: opts.RuleSet},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			opts = append(opts, grpc.UnaryInterceptor(interceptor.RequestLogging("host2plugin")))
			return grpc.NewServer(opts...)
		},
		Logger: logger.Logger(),
	})
}

// GetName returns the name of the plugin.
func (s *GRPCServer) GetName(ctx context.Context, req *proto.GetName_Request) (*proto.GetName_Response, error) {
	return &proto.GetName_Response{Name: s.impl.RuleSetName()}, nil
}

// GetVersion returns the version of the plugin.
func (s *GRPCServer) GetVersion(ctx context.Context, req *proto.GetVersion_Request) (*proto.GetVersion_Response, error) {
	return &proto.GetVersion_Response{Version: s.impl.RuleSetVersion()}, nil
}

// GetRuleNames returns the list of rule names provided by the plugin.
func (s *GRPCServer) GetRuleNames(ctx context.Context, req *proto.GetRuleNames_Request) (*proto.GetRuleNames_Response, error) {
	return &proto.GetRuleNames_Response{Names: s.impl.RuleNames()}, nil
}

// GetVersionConstraint returns a constraint of TFLint versions.
func (s *GRPCServer) GetVersionConstraint(ctx context.Context, req *proto.GetVersionConstraint_Request) (*proto.GetVersionConstraint_Response, error) {
	return &proto.GetVersionConstraint_Response{Constraint: s.impl.VersionConstraint()}, nil
}

// GetSDKVersion returns the SDK version.
func (s *GRPCServer) GetSDKVersion(ctx context.Context, req *proto.GetSDKVersion_Request) (*proto.GetSDKVersion_Response, error) {
	return &proto.GetSDKVersion_Response{Version: SDKVersion}, nil
}

// GetConfigSchema returns the config schema of the plugin.
func (s *GRPCServer) GetConfigSchema(ctx context.Context, req *proto.GetConfigSchema_Request) (*proto.GetConfigSchema_Response, error) {
	return &proto.GetConfigSchema_Response{Schema: toproto.BodySchema(s.impl.ConfigSchema())}, nil
}

// ApplyGlobalConfig applies a common config to the plugin.
func (s *GRPCServer) ApplyGlobalConfig(ctx context.Context, req *proto.ApplyGlobalConfig_Request) (*proto.ApplyGlobalConfig_Response, error) {
	if req.Config == nil {
		return nil, status.Error(codes.InvalidArgument, "config should not be null")
	}

	config := fromproto.Config(req.Config)
	if err := s.impl.ApplyGlobalConfig(config); err != nil {
		return nil, toproto.Error(codes.FailedPrecondition, err)
	}
	return &proto.ApplyGlobalConfig_Response{}, nil
}

// ApplyConfig applies the plugin config retrieved from the host to the plugin.
func (s *GRPCServer) ApplyConfig(ctx context.Context, req *proto.ApplyConfig_Request) (*proto.ApplyConfig_Response, error) {
	if req.Content == nil {
		return nil, status.Error(codes.InvalidArgument, "content should not be null")
	}

	content, diags := fromproto.BodyContent(req.Content)
	if diags.HasErrors() {
		return nil, toproto.Error(codes.InvalidArgument, diags)
	}
	if err := s.impl.ApplyConfig(content); err != nil {
		return nil, toproto.Error(codes.FailedPrecondition, err)
	}
	return &proto.ApplyConfig_Response{}, nil
}

// Check calls plugin rules with a gRPC client that can send requests
// to the host process.
func (s *GRPCServer) Check(ctx context.Context, req *proto.Check_Request) (*proto.Check_Response, error) {
	conn, err := s.broker.Dial(req.Runner)
	if err != nil {
		return nil, toproto.Error(codes.InvalidArgument, err)
	}
	defer conn.Close()

	runner, err := s.impl.NewRunner(&plugin2host.GRPCClient{Client: proto.NewRunnerClient(conn)})
	if err != nil {
		return nil, toproto.Error(codes.FailedPrecondition, err)
	}

	for _, rule := range s.impl.BuiltinImpl().EnabledRules {
		if err := rule.Check(runner); err != nil {
			return nil, toproto.Error(codes.Aborted, fmt.Errorf(`failed to check "%s" rule: %s`, rule.Name(), err))
		}
	}
	return &proto.Check_Response{}, nil
}
