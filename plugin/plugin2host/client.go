package plugin2host

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	hcljson "github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/fromproto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/toproto"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
	"github.com/zclconf/go-cty/cty/msgpack"
)

// GRPCClient is a plugin-side implementation. Plugin can send requests through the client to host's gRPC server.
type GRPCClient struct {
	Client proto.RunnerClient
}

var _ tflint.Runner = &GRPCClient{}

// GetResourceContent gets the contents of resources based on the schema.
// This is shorthand of GetModuleContent for resources
func (c *GRPCClient) GetResourceContent(name string, inner *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if opts == nil {
		opts = &tflint.GetModuleContentOption{}
	}
	opts.Hint.ResourceType = name

	body, err := c.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}, Body: inner},
		},
	}, opts)
	if err != nil {
		return nil, err
	}

	content := &hclext.BodyContent{Blocks: []*hclext.Block{}}
	for _, resource := range body.Blocks {
		if resource.Labels[0] != name {
			continue
		}

		content.Blocks = append(content.Blocks, resource)
	}

	return content, nil
}

// GetModuleContent gets the contents of the module based on the schema.
func (c *GRPCClient) GetModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if opts == nil {
		opts = &tflint.GetModuleContentOption{}
	}

	req := &proto.GetModuleContent_Request{
		Schema: toproto.BodySchema(schema),
		Option: toproto.GetModuleContentOption(opts),
	}
	resp, err := c.Client.GetModuleContent(context.Background(), req)
	if err != nil {
		return nil, fromproto.Error(err)
	}

	body, diags := fromproto.BodyContent(resp.Content)
	if diags.HasErrors() {
		err = diags
	}
	return body, err
}

// GetFile returns hcl.File based on the passed file name.
func (c *GRPCClient) GetFile(file string) (*hcl.File, error) {
	resp, err := c.Client.GetFile(context.Background(), &proto.GetFile_Request{Name: file})
	if err != nil {
		return nil, fromproto.Error(err)
	}

	var f *hcl.File
	var diags hcl.Diagnostics
	if strings.HasSuffix(file, ".tf") {
		f, diags = hclsyntax.ParseConfig(resp.File, file, hcl.InitialPos)
	} else {
		f, diags = hcljson.Parse(resp.File, file)
	}

	if diags.HasErrors() {
		err = diags
	}
	return f, err
}

// GetFiles returns bytes of hcl.File in the self module context.
func (c *GRPCClient) GetFiles() (map[string]*hcl.File, error) {
	resp, err := c.Client.GetFiles(context.Background(), &proto.GetFiles_Request{})
	if err != nil {
		return nil, fromproto.Error(err)
	}

	files := map[string]*hcl.File{}
	var f *hcl.File
	var diags hcl.Diagnostics
	for name, bytes := range resp.Files {
		var d hcl.Diagnostics
		if strings.HasSuffix(name, ".tf") {
			f, d = hclsyntax.ParseConfig(bytes, name, hcl.InitialPos)
		} else {
			f, d = hcljson.Parse(bytes, name)
		}
		diags = diags.Extend(d)

		files[name] = f
	}

	if diags.HasErrors() {
		return files, diags
	}
	return files, nil
}

// DecodeRuleConfig guesses the schema of the rule config from the passed interface and sends the schema to GRPC server.
// Content retrieved based on the schema is decoded into the passed interface.
func (c *GRPCClient) DecodeRuleConfig(name string, ret interface{}) error {
	resp, err := c.Client.GetRuleConfigContent(context.Background(), &proto.GetRuleConfigContent_Request{
		Name:   name,
		Schema: toproto.BodySchema(hclext.ImpliedBodySchema(ret)),
	})
	if err != nil {
		return fromproto.Error(err)
	}

	content, diags := fromproto.BodyContent(resp.Content)
	if diags.HasErrors() {
		return diags
	}
	diags = hclext.DecodeBody(content, nil, ret)
	if diags.HasErrors() {
		return diags
	}
	return nil
}

// EvaluateExpr evals the passed expression based on the type.
func (c *GRPCClient) EvaluateExpr(expr hcl.Expression, ret interface{}, opts *tflint.EvaluateExprOption) error {
	if opts == nil {
		opts = &tflint.EvaluateExprOption{}
	}

	var ty cty.Type
	if opts.WantType != nil {
		ty = *opts.WantType
	} else {
		switch ret.(type) {
		case *string, string:
			ty = cty.String
		case *int, int:
			ty = cty.Number
		case *[]string, []string:
			ty = cty.List(cty.String)
		case *[]int, []int:
			ty = cty.List(cty.Number)
		case *map[string]string, map[string]string:
			ty = cty.Map(cty.String)
		case *map[string]int, map[string]int:
			ty = cty.Map(cty.Number)
		case cty.Value, *cty.Value:
			ty = cty.DynamicPseudoType
		default:
			panic(fmt.Sprintf("unsupported result type: %T", ret))
		}
	}
	tyby, err := json.MarshalType(ty)
	if err != nil {
		return err
	}

	file, err := c.GetFile(expr.Range().Filename)
	if err != nil {
		return err
	}

	resp, err := c.Client.EvaluateExpr(
		context.Background(),
		&proto.EvaluateExpr_Request{
			Expr:      expr.Range().SliceBytes(file.Bytes),
			ExprRange: toproto.Range(expr.Range()),
			Option:    &proto.EvaluateExpr_Option{Type: tyby, ModuleCtx: toproto.ModuleCtxType(opts.ModuleCtx)},
		},
	)
	if err != nil {
		return fromproto.Error(err)
	}

	val, err := msgpack.Unmarshal(resp.Value, ty)
	if err != nil {
		return err
	}

	return gocty.FromCtyValue(val, ret)
}

// EmitIssue emits the issue with the passed rule, message, location
func (c *GRPCClient) EmitIssue(rule tflint.Rule, message string, location hcl.Range) error {
	_, err := c.Client.EmitIssue(context.Background(), &proto.EmitIssue_Request{Rule: toproto.Rule(rule), Message: message, Range: toproto.Range(location)})
	if err != nil {
		return fromproto.Error(err)
	}
	return nil
}

// EnsureNoError is a helper for error handling. Depending on the type of error generated by EvaluateExpr,
// determine whether to exit, skip, or continue. If it is continued, the passed function will be executed.
func (*GRPCClient) EnsureNoError(err error, proc func() error) error {
	if err == nil {
		return proc()
	}

	if errors.Is(err, tflint.ErrUnevaluable) || errors.Is(err, tflint.ErrNullValue) || errors.Is(err, tflint.ErrUnknownValue) {
		return nil
	}
	return err
}
