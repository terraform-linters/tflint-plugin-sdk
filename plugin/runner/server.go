package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/fromproto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/toproto"
	"github.com/terraform-linters/tflint-plugin-sdk/schema"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type GRPCServer struct {
	proto.UnimplementedRunnerServer

	Impl Host
}

type Host interface {
	ResourceContent(string, *schema.BodySchema) (*schema.BodyContent, hcl.Diagnostics)
	File(string) (*hcl.File, error)
	Files() (map[string]*hcl.File, error)
	EvaluateExpr(expr hcl.Expression, ty cty.Type) (cty.Value, error)
	EmitIssue(rule tflint.Rule, message string, location hcl.Range) error
}

func (s *GRPCServer) ResourceContent(ctx context.Context, req *proto.ResourceContent_Request) (*proto.ResourceContent_Response, error) {
	body, diags := s.Impl.ResourceContent(req.Resource, fromproto.BodySchema(req.Schema))

	sources := map[string][]byte{}
	files, err := s.Impl.Files()
	if err != nil {
		return nil, err
	}
	for name, file := range files {
		sources[name] = file.Bytes
	}

	content := toproto.BodyContent(body, sources)

	// TODO: Should return diags as response?
	if diags.HasErrors() {
		err = diags
	} else {
		err = nil
	}

	return &proto.ResourceContent_Response{Content: content}, err
}

func (s *GRPCServer) File(ctx context.Context, req *proto.File_Request) (*proto.File_Response, error) {
	file, err := s.Impl.File(req.Name)
	if err != nil {
		return nil, err
	}
	return &proto.File_Response{File: file.Bytes}, nil
}

func (s *GRPCServer) EvaluateExpr(ctx context.Context, req *proto.EvaluateExpr_Request) (*proto.EvaluateExpr_Response, error) {
	expr, diags := parseExpression(req.Expr, req.ExprRange.Filename, fromproto.Pos(req.ExprRange.Start))
	if diags.HasErrors() {
		return nil, diags
	}
	ty, err := ctyjson.UnmarshalType(req.Type)
	if err != nil {
		return nil, err
	}

	value, err := s.Impl.EvaluateExpr(expr, ty)
	if err != nil {
		return nil, err
	}
	val, err := ctyjson.Marshal(value, ty)
	if err != nil {
		return nil, err
	}

	return &proto.EvaluateExpr_Response{Value: val}, nil
}

func (s *GRPCServer) EmitIssue(ctx context.Context, req *proto.EmitIssue_Request) (*proto.EmitIssue_Response, error) {
	return &proto.EmitIssue_Response{}, s.Impl.EmitIssue(fromproto.EmitIssue_Rule(req.Rule), req.Message, fromproto.Range(req.Range))
}

// TODO: Move to another package
func parseExpression(src []byte, filename string, start hcl.Pos) (hcl.Expression, hcl.Diagnostics) {
	if strings.HasSuffix(filename, ".tf") {
		// HACK: Always add a newline to avoid heredoc parse errors.
		// @see https://github.com/hashicorp/hcl/issues/441
		src = []byte(string(src) + "\n")
		return hclsyntax.ParseExpression(src, filename, start)
	}

	if strings.HasSuffix(filename, ".tf.json") {
		return json.ParseExpressionWithStartPos(src, filename, start)
	}

	panic(fmt.Sprintf("Unexpected file: %s", filename))
}
