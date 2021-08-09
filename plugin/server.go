package plugin

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/fromproto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/runner"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/toproto"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	tfclient "github.com/terraform-linters/tflint-plugin-sdk/tflint/client"
)

// Server is an RPC server acting as a plugin.
type Server struct {
	impl   tflint.RuleSet
	broker *plugin.MuxBroker
}

type GRPCServer struct {
	proto.UnimplementedRuleSetServer

	impl   tflint.RuleSet
	broker *plugin.GRPCBroker
}

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
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// RuleSetName returns the name of the plugin.
func (s *Server) RuleSetName(args interface{}, resp *string) error {
	*resp = s.impl.RuleSetName()
	return nil
}

func (s *GRPCServer) RuleSetName(ctx context.Context, req *proto.RuleSetName_Request) (*proto.RuleSetName_Response, error) {
	return &proto.RuleSetName_Response{Name: s.impl.RuleSetName()}, nil
}

// RuleSetVersion returns the version of the plugin.
func (s *Server) RuleSetVersion(args interface{}, resp *string) error {
	*resp = s.impl.RuleSetVersion()
	return nil
}

func (s *GRPCServer) RuleSetVersion(ctx context.Context, req *proto.RuleSetVersion_Request) (*proto.RuleSetVersion_Response, error) {
	return &proto.RuleSetVersion_Response{Version: s.impl.RuleSetVersion()}, nil
}

// RuleNames returns the list of rule names provided by the plugin.
func (s *Server) RuleNames(args interface{}, resp *[]string) error {
	*resp = s.impl.RuleNames()
	return nil
}

func (s *GRPCServer) RuleNames(ctx context.Context, req *proto.RuleNames_Request) (*proto.RuleNames_Response, error) {
	return &proto.RuleNames_Response{Names: s.impl.RuleNames()}, nil
}

func (s *GRPCServer) ConfigSchema(ctx context.Context, req *proto.ConfigSchema_Request) (*proto.ConfigSchema_Response, error) {
	return toproto.ConfigSchema_Response(s.impl.ConfigSchema()), nil
}

// ApplyConfig applies the passed config to its own plugin implementation.
func (s *Server) ApplyConfig(config *tflint.MarshalledConfig, resp *interface{}) error {
	cfg, err := config.Unmarshal()
	if err != nil {
		return err
	}
	return s.impl.ApplyConfig(cfg)
}

func (s *GRPCServer) ApplyConfig(ctx context.Context, req *proto.ApplyConfig_Request) (*proto.ApplyConfig_Response, error) {
	body, diags := fromproto.BodyContent(req.Body)
	if diags.HasErrors() {
		return nil, diags
	}

	return &proto.ApplyConfig_Response{}, s.impl.NewApplyConfig(body)
}

// Check calls its own plugin implementation with an RPC client that can send
// requests to the host process.
func (s *Server) Check(brokerID uint32, resp *interface{}) error {
	conn, err := s.broker.Dial(brokerID)
	if err != nil {
		return err
	}

	return s.impl.Check(tfclient.NewClient(conn))
}

func (s *GRPCServer) Check(ctx context.Context, req *proto.Check_Request) (*proto.Check_Response, error) {
	conn, err := s.broker.Dial(req.Runner)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = s.impl.NewCheck(&runner.GRPCClient{Client: proto.NewRunnerClient(conn)})

	return &proto.Check_Response{}, err
}
