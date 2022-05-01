package plugin2host

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/fromproto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/toproto"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
	"github.com/zclconf/go-cty/cty/msgpack"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer is a host-side implementation. Host must implement a server that returns a response for a request from plugin.
// The behavior as gRPC server is implemented in the SDK, and the actual behavior is delegated to impl.
type GRPCServer struct {
	proto.UnimplementedRunnerServer

	Impl Server
}

var _ proto.RunnerServer = &GRPCServer{}

// Server is the interface that the host should implement when a plugin communicates with the host.
type Server interface {
	GetModuleContent(*hclext.BodySchema, tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics)
	GetFile(string) (*hcl.File, error)
	// For performance, GetFiles returns map[string][]bytes instead of map[string]*hcl.File.
	GetFiles(tflint.ModuleCtxType) map[string][]byte
	GetRuleConfigContent(string, *hclext.BodySchema) (*hclext.BodyContent, map[string][]byte, error)
	EvaluateExpr(hcl.Expression, tflint.EvaluateExprOption) (cty.Value, error)
	EmitIssue(rule tflint.Rule, message string, location hcl.Range) error
}

// GetModuleContent gets the contents of the module based on the schema.
func (s *GRPCServer) GetModuleContent(ctx context.Context, req *proto.GetModuleContent_Request) (*proto.GetModuleContent_Response, error) {
	if req.Schema == nil {
		return nil, status.Error(codes.InvalidArgument, "schema should not be null")
	}
	if req.Option == nil {
		return nil, status.Error(codes.InvalidArgument, "option should not be null")
	}

	opts := fromproto.GetModuleContentOption(req.Option)
	body, diags := s.Impl.GetModuleContent(fromproto.BodySchema(req.Schema), opts)
	if diags.HasErrors() {
		return nil, toproto.Error(codes.FailedPrecondition, diags)
	}
	if body == nil {
		return nil, status.Error(codes.FailedPrecondition, "response body is empty")
	}

	content := toproto.BodyContent(body, s.Impl.GetFiles(opts.ModuleCtx))

	return &proto.GetModuleContent_Response{Content: content}, nil
}

// GetFile returns bytes of hcl.File based on the passed file name.
func (s *GRPCServer) GetFile(ctx context.Context, req *proto.GetFile_Request) (*proto.GetFile_Response, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name should not be empty")
	}
	file, err := s.Impl.GetFile(req.Name)
	if err != nil {
		return nil, toproto.Error(codes.FailedPrecondition, err)
	}
	if file == nil {
		return nil, status.Error(codes.NotFound, "file not found")
	}
	return &proto.GetFile_Response{File: file.Bytes}, nil
}

// GetFiles returns bytes of hcl.File in the self module context.
func (s *GRPCServer) GetFiles(ctx context.Context, req *proto.GetFiles_Request) (*proto.GetFiles_Response, error) {
	return &proto.GetFiles_Response{Files: s.Impl.GetFiles(tflint.SelfModuleCtxType)}, nil
}

// GetRuleConfigContent returns BodyContent based on the rule name and config schema.
func (s *GRPCServer) GetRuleConfigContent(ctx context.Context, req *proto.GetRuleConfigContent_Request) (*proto.GetRuleConfigContent_Response, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name should not be empty")
	}
	if req.Schema == nil {
		return nil, status.Error(codes.InvalidArgument, "schema should not be null")
	}

	body, sources, err := s.Impl.GetRuleConfigContent(req.Name, fromproto.BodySchema(req.Schema))
	if err != nil {
		return nil, toproto.Error(codes.FailedPrecondition, err)
	}
	if body == nil {
		return nil, status.Error(codes.FailedPrecondition, "response body is empty")
	}
	if len(sources) == 0 {
		return nil, status.Error(codes.NotFound, "config file not found")
	}

	content := toproto.BodyContent(body, sources)
	return &proto.GetRuleConfigContent_Response{Content: content}, nil
}

// EvaluateExpr evals the passed expression based on the type.
func (s *GRPCServer) EvaluateExpr(ctx context.Context, req *proto.EvaluateExpr_Request) (*proto.EvaluateExpr_Response, error) {
	if req.Expr == nil {
		return nil, status.Error(codes.InvalidArgument, "expr should not be null")
	}
	if req.ExprRange == nil {
		return nil, status.Error(codes.InvalidArgument, "expr_range should not be null")
	}
	if req.Option == nil {
		return nil, status.Error(codes.InvalidArgument, "option should not be null")
	}

	expr, diags := hclext.ParseExpression(req.Expr, req.ExprRange.Filename, fromproto.Pos(req.ExprRange.Start))
	if diags.HasErrors() {
		return nil, toproto.Error(codes.InvalidArgument, diags)
	}
	ty, err := json.UnmarshalType(req.Option.Type)
	if err != nil {
		return nil, toproto.Error(codes.InvalidArgument, err)
	}

	value, err := s.Impl.EvaluateExpr(expr, tflint.EvaluateExprOption{WantType: &ty, ModuleCtx: fromproto.ModuleCtxType(req.Option.ModuleCtx)})
	if err != nil {
		return nil, toproto.Error(codes.FailedPrecondition, err)
	}
	val, err := msgpack.Marshal(value, ty)
	if err != nil {
		return nil, toproto.Error(codes.FailedPrecondition, err)
	}

	return &proto.EvaluateExpr_Response{Value: val}, nil
}

// EmitIssue emits the issue with the passed rule, message, location
func (s *GRPCServer) EmitIssue(ctx context.Context, req *proto.EmitIssue_Request) (*proto.EmitIssue_Response, error) {
	if req.Rule == nil {
		return nil, status.Error(codes.InvalidArgument, "rule should not be null")
	}
	if req.Range == nil {
		return nil, status.Error(codes.InvalidArgument, "range should not be null")
	}

	err := s.Impl.EmitIssue(fromproto.Rule(req.Rule), req.Message, fromproto.Range(req.Range))
	if err != nil {
		return nil, toproto.Error(codes.FailedPrecondition, err)
	}
	return &proto.EmitIssue_Response{}, nil
}
