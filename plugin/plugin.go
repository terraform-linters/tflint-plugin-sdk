package plugin

import (
	"encoding/gob"
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// handShakeConfig is used for UX. ProcotolVersion will be updated by incompatible changes
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TFLINT_RULESET_PLUGIN",
	MagicCookieValue: "5adSn1bX8nrDfgBqiAqqEkC6OE1h3iD8SqbMc5UUONx8x3xCF0KlPDsBRNDjoYDP",
}

// RuleSetPlugin is a wrapper to satisfy the interface of go-plugin
type RuleSetPlugin struct {
	impl tflint.RuleSet
}

// Server returns an RPC server acting as a plugin
func (p *RuleSetPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &Server{impl: p.impl, broker: b}, nil
}

// Client returns an RPC client for use by the host
func (RuleSetPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &Client{rpcClient: c, broker: b}, nil
}

// In order to communicate the interface correctly with RPC,
// the type of the related structure is registered in gob at the initial time.
func init() {
	gob.Register(&hclsyntax.TemplateExpr{})
	gob.Register(&hclsyntax.LiteralValueExpr{})
	gob.Register(&hclsyntax.ScopeTraversalExpr{})
	gob.Register(hcl.TraverseRoot{})
	gob.Register(hcl.TraverseAttr{})
}
