package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Client is an RPC client for plugins.
type Client struct {
	rpcClient *rpc.Client
}

// NewClient returns a new Client.
func NewClient(conn net.Conn) *Client {
	return &Client{rpcClient: rpc.NewClient(conn)}
}

// WalkResourceAttributes calls the server-side Attributes method and passes the decoded response
// to the passed function.
func (c *Client) WalkResourceAttributes(resource, attributeName string, walker func(*hcl.Attribute) error) error {
	log.Printf("[DEBUG] Walk `%s.*.%s` attribute", resource, attributeName)

	var response AttributesResponse
	if err := c.rpcClient.Call("Plugin.Attributes", AttributesRequest{Resource: resource, AttributeName: attributeName}, &response); err != nil {
		return err
	}
	if response.Err != nil {
		return response.Err
	}

	for _, attr := range response.Attributes {
		attribute, diags := decodeAttribute(attr)
		if diags.HasErrors() {
			return diags
		}

		if err := walker(attribute); err != nil {
			return err
		}
	}

	return nil
}

// WalkResourceBlocks calls the server-side Blocks method and passes the decoded response
// to the passed function.
func (c *Client) WalkResourceBlocks(resource, blockType string, walker func(*hcl.Block) error) error {
	log.Printf("[DEBUG] Walk `%s.*.%s` block", resource, blockType)

	var response BlocksResponse
	if err := c.rpcClient.Call("Plugin.Blocks", BlocksRequest{Resource: resource, BlockType: blockType}, &response); err != nil {
		return err
	}
	if response.Err != nil {
		return response.Err
	}

	for _, b := range response.Blocks {
		block, diags := decodeBlock(b)
		if diags.HasErrors() {
			return diags
		}

		if err := walker(block); err != nil {
			return err
		}
	}

	return nil
}

// WalkResources calls the server-side Resources method and passes the decoded response
// to the passed function.
func (c *Client) WalkResources(resource string, walker func(*terraform.Resource) error) error {
	log.Printf("[DEBUG] Walk `%s` resource", resource)

	var response ResourcesResponse
	if err := c.rpcClient.Call("Plugin.Resources", ResourcesRequest{Name: resource}, &response); err != nil {
		return err
	}
	if response.Err != nil {
		return response.Err
	}

	for _, r := range response.Resources {
		resource, diags := decodeResource(r)
		if diags.HasErrors() {
			return diags
		}

		if err := walker(resource); err != nil {
			return err
		}
	}
	return nil
}

// WalkModuleCalls calls the server-side ModuleCalls method and passed the decode response
// to the passed function.
func (c *Client) WalkModuleCalls(walker func(*terraform.ModuleCall) error) error {
	log.Printf("[DEBUG] WalkModuleCalls")

	var response ModuleCallsResponse
	if err := c.rpcClient.Call("Plugin.ModuleCalls", ModuleCallsRequest{}, &response); err != nil {
		return err
	}
	if response.Err != nil {
		return response.Err
	}

	for _, c := range response.ModuleCalls {
		call, diags := decodeModuleCall(c)
		if diags.HasErrors() {
			return diags
		}

		if err := walker(call); err != nil {
			return err
		}
	}

	return nil
}

// Backend calls the server-side Backend method and returns the backend configuration.
func (c *Client) Backend() (*terraform.Backend, error) {
	log.Printf("[DEBUG] Backend")

	var response BackendResponse
	if err := c.rpcClient.Call("Plugin.Backend", BackendRequest{}, &response); err != nil {
		return nil, err
	}
	if response.Err != nil {
		return nil, response.Err
	}

	backend, diags := decodeBackend(response.Backend)
	if diags.HasErrors() {
		return nil, diags
	}

	return backend, nil
}

// EvaluateExpr calls the server-side EvalExpr method and reflects the response
// in the passed argument.
func (c *Client) EvaluateExpr(expr hcl.Expression, ret interface{}) error {
	return c.evaluateExpr(expr, ret, cty.Type{})
}

// EvaluateExprType calls the server-side EvalExpr method with a specific cty.Type
// and reflects the response in the passed argument.
func (c *Client) EvaluateExprType(expr hcl.Expression, ret interface{}, wantType cty.Type) error {
	return c.evaluateExpr(expr, ret, wantType)
}

// EvaluateExprType calls the server-side EvalExpr method and reflects the response
// in the passed argument.
func (c *Client) evaluateExpr(expr hcl.Expression, ret interface{}, wantType cty.Type) error {
	var response EvalExprResponse
	var err error

	src, err := ioutil.ReadFile(expr.Range().Filename)
	if err != nil {
		return err
	}
	req := EvalExprRequest{Ret: ret, WantType: wantType}
	req.Expr, req.ExprRange = encodeExpr(src, expr)
	if err := c.rpcClient.Call("Plugin.EvalExpr", req, &response); err != nil {
		return err
	}
	if response.Err != nil {
		return response.Err
	}

	err = gocty.FromCtyValue(response.Val, ret)
	if err != nil {
		err := &tflint.Error{
			Code:  tflint.TypeMismatchError,
			Level: tflint.ErrorLevel,
			Message: fmt.Sprintf(
				"Invalid type expression in %s:%d",
				expr.Range().Filename,
				expr.Range().Start.Line,
			),
			Cause: err,
		}
		log.Printf("[ERROR] %s", err)
		return err
	}
	return nil
}

// EmitIssueOnExpr calls the server-side EmitIssue method with the passed expression.
func (c *Client) EmitIssueOnExpr(rule tflint.Rule, message string, expr hcl.Expression) error {
	req := &EmitIssueRequest{
		Rule:     encodeRule(rule),
		Message:  message,
		Location: expr.Range(),
	}

	src, err := ioutil.ReadFile(expr.Range().Filename)
	if err != nil {
		return err
	}
	req.Expr, req.ExprRange = encodeExpr(src, expr)

	if err := c.rpcClient.Call("Plugin.EmitIssue", &req, new(interface{})); err != nil {
		return err
	}
	return nil
}

// EmitIssue calls the server-side EmitIssue method with the passed rule and range.
// You should use EmitIssueOnExpr if you want to emit an issue for an expression.
// This API provides a lower level interface.
func (c *Client) EmitIssue(rule tflint.Rule, message string, location hcl.Range) error {
	req := &EmitIssueRequest{
		Rule:     encodeRule(rule),
		Message:  message,
		Location: location,
	}

	if err := c.rpcClient.Call("Plugin.EmitIssue", &req, new(interface{})); err != nil {
		return err
	}
	return nil
}

// EnsureNoError is a helper for error handling. Depending on the type of error generated by EvaluateExpr,
// determine whether to exit, skip, or continue. If it is continued, the passed function will be executed.
func (*Client) EnsureNoError(err error, proc func() error) error {
	if err == nil {
		return proc()
	}

	if appErr, ok := err.(tflint.Error); ok {
		switch appErr.Level {
		case tflint.WarningLevel:
			return nil
		case tflint.ErrorLevel:
			return appErr
		default:
			panic(appErr)
		}
	} else {
		return err
	}
}
