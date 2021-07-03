package plugin

import (
	"encoding/gob"
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// handShakeConfig is used for UX. ProcotolVersion will be updated by incompatible changes.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  9,
	MagicCookieKey:   "TFLINT_RULESET_PLUGIN",
	MagicCookieValue: "5adSn1bX8nrDfgBqiAqqEkC6OE1h3iD8SqbMc5UUONx8x3xCF0KlPDsBRNDjoYDP",
}

// RuleSetPlugin is a wrapper to satisfy the interface of go-plugin.
type RuleSetPlugin struct {
	impl tflint.RuleSet
}

// Server returns an RPC server acting as a plugin.
func (p *RuleSetPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &Server{impl: p.impl, broker: b}, nil
}

// Client returns an RPC client for the host.
func (RuleSetPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &Client{rpcClient: c, broker: b}, nil
}

func init() {
	gob.Register(tflint.Error{})
}
