package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	hcljson "github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/fromproto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/toproto"
	"github.com/terraform-linters/tflint-plugin-sdk/schema"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
)

type GRPCClient struct {
	Client proto.RunnerClient
}

func (c *GRPCClient) ResourceContent(resource string, schema *schema.BodySchema) (*schema.BodyContent, hcl.Diagnostics) {
	// TODO: handle errors
	resp, _ := c.Client.ResourceContent(context.Background(), &proto.ResourceContent_Request{Resource: resource, Schema: toproto.BodySchema(schema)})
	body, diags := fromproto.BodyContent(resp.Content)
	return body, diags
}

func (c *GRPCClient) File(file string) (*hcl.File, error) {
	resp, err := c.Client.File(context.Background(), &proto.File_Request{Name: file})
	if err != nil {
		return nil, err
	}

	var f *hcl.File
	var diags hcl.Diagnostics
	if strings.HasSuffix(file, ".tf") {
		f, diags = hclsyntax.ParseConfig(resp.File, file, hcl.InitialPos)
	} else {
		f, diags = hcljson.Parse(resp.File, file)
	}

	if diags.HasErrors() {
		err = errors.New(diags.Error())
	}
	return f, err
}

func (c *GRPCClient) Files() (map[string]*hcl.File, error) {
	return map[string]*hcl.File{}, nil
}

func (c *GRPCClient) EvaluateExpr(expr hcl.Expression, ret interface{}, wantType *cty.Type) error {
	file, err := c.File(expr.Range().Filename)
	if err != nil {
		return err
	}

	var ty cty.Type
	if wantType != nil {
		ty = *wantType
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
		default:
			panic(fmt.Errorf("Unexpected result type: %T", ret))
		}
	}
	tyby, err := json.MarshalType(ty)
	if err != nil {
		return err
	}

	resp, err := c.Client.EvaluateExpr(
		context.Background(),
		&proto.EvaluateExpr_Request{
			Expr:      expr.Range().SliceBytes(file.Bytes),
			ExprRange: toproto.Range(expr.Range()),
			Type:      tyby,
		},
	)
	if err != nil {
		return err
	}

	val, err := json.Unmarshal(resp.Value, ty)
	if err != nil {
		return err
	}

	return gocty.FromCtyValue(val, ret)
}

func (c *GRPCClient) EmitIssue(rule tflint.Rule, message string, location hcl.Range) error {
	_, err := c.Client.EmitIssue(context.Background(), &proto.EmitIssue_Request{Rule: toproto.EmitIssue_Rule(rule), Message: message, Range: toproto.Range(location)})
	return err
}
