package plugin

import (
	plugin "github.com/hashicorp/go-plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	tfclient "github.com/terraform-linters/tflint-plugin-sdk/tflint/client"
)

// Server is an RPC server acting as a plugin.
type Server struct {
	impl   tflint.RuleSet
	broker *plugin.MuxBroker
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
	})
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
func (s *Server) ApplyConfig(config *tflint.Config, resp *interface{}) error {
	s.impl.ApplyConfig(config)
	return nil
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
