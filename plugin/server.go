package plugin

import (
	plugin "github.com/hashicorp/go-plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/host2plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	tfclient "github.com/terraform-linters/tflint-plugin-sdk/tflint/client"
)

// ServeOpts is an option for serving a plugin.
// Each plugin can pass a RuleSet that represents its own functionality.
type ServeOpts = host2plugin.ServeOpts

// Serve is a wrapper of plugin.Serve. This is entrypoint of all plugins.
var Serve = host2plugin.Serve

// Server is an RPC server acting as a plugin.
type Server struct {
	impl   tflint.RPCRuleSet
	broker *plugin.MuxBroker
}

// RuleSetName returns the name of the plugin.
func (s *Server) RuleSetName(args interface{}, resp *string) error {
	*resp = s.impl.RuleSetName()
	return nil
}

// RuleSetVersion returns the version of the plugin.
func (s *Server) RuleSetVersion(args interface{}, resp *string) error {
	*resp = s.impl.RuleSetVersion()
	return nil
}

// RuleNames returns the list of rule names provided by the plugin.
func (s *Server) RuleNames(args interface{}, resp *[]string) error {
	*resp = s.impl.RuleNames()
	return nil
}

// ApplyConfig applies the passed config to its own plugin implementation.
func (s *Server) ApplyConfig(config *tflint.MarshalledConfig, resp *interface{}) error {
	cfg, err := config.Unmarshal()
	if err != nil {
		return err
	}
	return s.impl.ApplyConfig(cfg)
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
