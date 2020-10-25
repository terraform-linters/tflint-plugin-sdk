package plugin

import (
	"net/rpc"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	tfserver "github.com/terraform-linters/tflint-plugin-sdk/tflint/server"
)

// Client is an RPC client for the host.
type Client struct {
	rpcClient *rpc.Client
	broker    *plugin.MuxBroker
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
		Cmd: opts.Cmd,
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

// RuleSetVersion calls the server-side RuleSetVersion method and returns its version.
func (c *Client) RuleSetVersion() (string, error) {
	var resp string
	err := c.rpcClient.Call("Plugin.RuleSetVersion", new(interface{}), &resp)
	return resp, err
}

// RuleNames calls the server-side RuleNames method and returns the list of names.
func (c *Client) RuleNames() ([]string, error) {
	var resp []string
	err := c.rpcClient.Call("Plugin.RuleNames", new(interface{}), &resp)
	return resp, err
}

// ApplyConfig calls the server-side ApplyConfig method.
func (c *Client) ApplyConfig(config *tflint.MarshalledConfig) error {
	return c.rpcClient.Call("Plugin.ApplyConfig", config, new(interface{}))
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
