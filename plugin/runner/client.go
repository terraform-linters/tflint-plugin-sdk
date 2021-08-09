package runner

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/fromproto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/toproto"
	"github.com/terraform-linters/tflint-plugin-sdk/schema"
)

type GRPCClient struct {
	Client proto.RunnerClient
}

func (c *GRPCClient) ResourceContent(resource string, schema *schema.BodySchema) (*schema.BodyContent, hcl.Diagnostics) {
	// TODO
	resp, _ := c.Client.ResourceContent(context.Background(), &proto.ResourceContent_Request{Resource: resource, Schema: toproto.BodySchema(schema)})
	body, diags := fromproto.BodyContent(resp.Content)
	return body, diags
}
