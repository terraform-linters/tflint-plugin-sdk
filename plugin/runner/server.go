package runner

import (
	"context"

	"github.com/terraform-linters/tflint-plugin-sdk/plugin/fromproto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/toproto"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type GRPCServer struct {
	proto.UnimplementedRunnerServer

	Impl tflint.Runner
}

func (s *GRPCServer) ResourceContent(ctx context.Context, req *proto.ResourceContent_Request) (*proto.ResourceContent_Response, error) {
	body, diags := s.Impl.ResourceContent(req.Resource, fromproto.BodySchema(req.Schema))

	return &proto.ResourceContent_Response{Content: toproto.BodyContent(body)}, diags
}
