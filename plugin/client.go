package plugin

import (
	"context"
	"net/rpc"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/fromproto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/runner"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/toproto"
	"github.com/terraform-linters/tflint-plugin-sdk/schema"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	tfserver "github.com/terraform-linters/tflint-plugin-sdk/tflint/server"
	"google.golang.org/grpc"
)

// Client is an RPC client for the host.
type Client struct {
	rpcClient *rpc.Client
	broker    *plugin.MuxBroker
}

type GRPCClient struct {
	broker *plugin.GRPCBroker
	client proto.RuleSetClient
}

// ClientOpts is an option for initializing a Client.
type ClientOpts struct {
	Cmd *exec.Cmd
}

// NewClient is a wrapper of plugin.NewClient.
func NewClient(opts *ClientOpts) *plugin.Client {
	return plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"ruleset": &RuleSetPlugin{},
		},
		Cmd:              opts.Cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger: hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Output: os.Stderr,
			Level:  hclog.LevelFromString(os.Getenv("TFLINT_LOG")),
		}),
	})
}

// RuleSetName calls the server-side RuleSetName method and returns its name.
func (c *Client) RuleSetName() (string, error) {
	var resp string
	err := c.rpcClient.Call("Plugin.RuleSetName", new(interface{}), &resp)
	return resp, err
}

func (c *GRPCClient) RuleSetName() (string, error) {
	resp, err := c.client.RuleSetName(context.Background(), &proto.RuleSetName_Request{})
	if err != nil {
		return "", err
	}
	return resp.Name, nil
}

// RuleSetVersion calls the server-side RuleSetVersion method and returns its version.
func (c *Client) RuleSetVersion() (string, error) {
	var resp string
	err := c.rpcClient.Call("Plugin.RuleSetVersion", new(interface{}), &resp)
	return resp, err
}

func (c *GRPCClient) RuleSetVersion() (string, error) {
	resp, err := c.client.RuleSetVersion(context.Background(), &proto.RuleSetVersion_Request{})
	if err != nil {
		return "", err
	}
	return resp.Version, nil
}

// RuleNames calls the server-side RuleNames method and returns the list of names.
func (c *Client) RuleNames() ([]string, error) {
	var resp []string
	err := c.rpcClient.Call("Plugin.RuleNames", new(interface{}), &resp)
	return resp, err
}

func (c *GRPCClient) RuleNames() ([]string, error) {
	resp, err := c.client.RuleNames(context.Background(), &proto.RuleNames_Request{})
	if err != nil {
		return []string{}, err
	}
	return resp.Names, nil
}

func (c *GRPCClient) ConfigSchema() (*schema.BodySchema, error) {
	resp, err := c.client.ConfigSchema(context.Background(), &proto.ConfigSchema_Request{})
	if err != nil {
		return nil, err
	}
	return fromproto.BodySchema(resp.Body), nil
}

// ApplyConfig calls the server-side ApplyConfig method.
func (c *Client) ApplyConfig(config *tflint.MarshalledConfig) error {
	return c.rpcClient.Call("Plugin.ApplyConfig", config, new(interface{}))
}

func (c *GRPCClient) ApplyConfig(config *schema.BodyContent, sources map[string][]byte) error {
	_, err := c.client.ApplyConfig(context.Background(), toproto.ApplyConfig_Request(config, sources))
	return err
}

// Check calls the server-side Check method.
// At the same time, it starts the server to respond to requests from the plugin side.
// Note that this server (tfserver.Server) serves clients that satisfy the Runner interface
// and is different from the server (plugin.Server) that provides the plugin system.
func (c *Client) Check(server tfserver.Server) error {
	brokerID := c.broker.NextId()
	go c.broker.AcceptAndServe(brokerID, server)

	return c.rpcClient.Call("Plugin.Check", brokerID, new(interface{}))
}

func (c *GRPCClient) Check(r runner.Host) error {
	brokerID := c.broker.NextId()
	go c.broker.AcceptAndServe(brokerID, func(opts []grpc.ServerOption) *grpc.Server {
		server := grpc.NewServer(opts...)
		proto.RegisterRunnerServer(server, &runner.GRPCServer{Impl: r})
		return server
	})

	_, err := c.client.Check(context.Background(), &proto.Check_Request{Runner: brokerID})
	return err
}
