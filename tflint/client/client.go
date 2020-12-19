package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
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
func (c *Client) WalkResources(resource string, walker func(*configs.Resource) error) error {
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
func (c *Client) WalkModuleCalls(walker func(*configs.ModuleCall) error) error {
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
func (c *Client) Backend() (*configs.Backend, error) {
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

// Config calls the server-side Config method and returns the Terraform configuration.
func (c *Client) Config() (*configs.Config, error) {
	log.Print("[DEBUG] Accessing to Config")

	var response ConfigResponse
	if err := c.rpcClient.Call("Plugin.Config", ConfigRequest{}, &response); err != nil {
		return nil, err
	}
	if response.Err != nil {
		return nil, response.Err
	}

	config, diags := decodeConfig(response.Config)
	if diags.HasErrors() {
		return nil, diags
	}

	return config, nil
}

// RootProvider calls the server-side RootProvider method and returns the provider configuration.
func (c *Client) RootProvider(name string) (*configs.Provider, error) {
	log.Printf("[DEBUG] Accessing to the `%s` provider config in the root module", name)

	var response RootProviderResponse
	if err := c.rpcClient.Call("Plugin.RootProvider", RootProviderRequest{Name: name}, &response); err != nil {
		return nil, err
	}
	if response.Err != nil {
		return nil, response.Err
	}

	provider, diags := decodeProvider(response.Provider)
	if diags.HasErrors() {
		return nil, diags
	}

	return provider, nil
}

// DecodeRuleConfig calls the server-side RuleConfig method and reflects the response
// in the passed argument.
func (c *Client) DecodeRuleConfig(name string, ret interface{}) error {
	var response RuleConfigResponse

	req := RuleConfigRequest{Name: name}
	if err := c.rpcClient.Call("Plugin.RuleConfig", req, &response); err != nil {
		return err
	}
	if response.Err != nil {
		return response.Err
	}

	if !response.Exists {
		return nil
	}
	file, diags := hclsyntax.ParseConfig(response.Config, response.Range.Filename, response.Range.Start)
	if diags.HasErrors() {
		return diags
	}
	if diags = gohcl.DecodeBody(file.Body, nil, ret); diags.HasErrors() {
		return diags
	}

	return nil
}

// EvaluateExpr calls the server-side EvalExpr method and reflects the response
// in the passed argument.
func (c *Client) EvaluateExpr(expr hcl.Expression, ret interface{}, wantType *cty.Type) error {
	if wantType == nil {
		wantType = &cty.Type{}
	}

	var response EvalExprResponse
	var err error

	src, err := ioutil.ReadFile(expr.Range().Filename)
	if err != nil {
		return err
	}
	req := EvalExprRequest{Ret: ret, Type: *wantType}
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

// EvaluateExprOnRootCtx calls the server-side EvalExprOnRootCtx method and reflects the response
// in the passed argument.
func (c *Client) EvaluateExprOnRootCtx(expr hcl.Expression, ret interface{}, wantType *cty.Type) error {
	if wantType == nil {
		wantType = &cty.Type{}
	}

	var response EvalExprResponse
	var err error

	src, err := ioutil.ReadFile(expr.Range().Filename)
	if err != nil {
		return err
	}
	req := EvalExprRequest{Ret: ret, Type: *wantType}
	req.Expr, req.ExprRange = encodeExpr(src, expr)
	if err := c.rpcClient.Call("Plugin.EvalExprOnRootCtx", req, &response); err != nil {
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

// IsNullExpr calls the server-side IsNullExpr method with the passed expression.
func (c *Client) IsNullExpr(expr hcl.Expression) (bool, error) {
	var response IsNullExprResponse

	src, err := ioutil.ReadFile(expr.Range().Filename)
	if err != nil {
		return false, err
	}
	req := &IsNullExprRequest{}
	req.Expr, req.Range = encodeExpr(src, expr)
	if err := c.rpcClient.Call("Plugin.IsNullExpr", req, &response); err != nil {
		return false, err
	}

	return response.Ret, response.Err
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
