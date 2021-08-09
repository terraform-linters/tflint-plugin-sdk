package plugin

import (
	"context"
	"encoding/gob"
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"google.golang.org/grpc"
)

// handShakeConfig is used for UX. ProcotolVersion will be updated by incompatible changes.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  10,
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

func (p *RuleSetPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterRuleSetServer(s, &GRPCServer{
		impl:   p.impl,
		broker: broker,
	})
	return nil
}

// Client returns an RPC client for the host.
func (RuleSetPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &Client{rpcClient: c, broker: b}, nil
}

func (*RuleSetPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{
		client: proto.NewRuleSetClient(c),
		broker: broker,
	}, nil
}

func init() {
	gob.Register(tflint.Error{})
}
