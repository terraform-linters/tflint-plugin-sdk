package plugin

import (
	"net/rpc"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// Client is an RPC client for use by the host
type Client struct {
	rpcClient *rpc.Client
	broker    *plugin.MuxBroker
}

// ClientOpts is an option for initializing the RPC client
type ClientOpts struct {
	Cmd *exec.Cmd
}

// NewClient is a wrapper of plugin.NewClient
func NewClient(opts *ClientOpts) *plugin.Client {
	return plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"ruleset": &RuleSetPlugin{},
		},
		Cmd: opts.Cmd,
		Logger: hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Output: os.Stderr,
			Level:  hclog.LevelFromString(os.Getenv("TFLINT_LOG")),
		}),
	})
}

// RuleSetName queries the RPC server for RuleSetName
func (c *Client) RuleSetName() (string, error) {
	var resp string
	err := c.rpcClient.Call("Plugin.RuleSetName", new(interface{}), &resp)
	return resp, err
}

// RuleSetVersion queries the RPC server for RuleSetVersion
func (c *Client) RuleSetVersion() (string, error) {
	var resp string
	err := c.rpcClient.Call("Plugin.RuleSetVersion", new(interface{}), &resp)
	return resp, err
}

// RuleNames queries the RPC server for RuleNames
func (c *Client) RuleNames() ([]string, error) {
	var resp []string
	err := c.rpcClient.Call("Plugin.RuleNames", new(interface{}), &resp)
	return resp, err
}

// ApplyConfig queries the RPC server for ApplyConfig
func (c *Client) ApplyConfig(config *tflint.Config) error {
	return c.rpcClient.Call("Plugin.ApplyConfig", config, new(interface{}))
}

// Check queries the RPC server for Check
// For bi-directional communication, you can pass a server that accepts Runner's queries
func (c *Client) Check(server tflint.Server) error {
	brokerID := c.broker.NextId()
	go c.broker.AcceptAndServe(brokerID, server)

	return c.rpcClient.Call("Plugin.Check", brokerID, new(interface{}))
}
