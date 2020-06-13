package tflint

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Client is an RPC client for plugins to query the host process for Terraform configurations
// Actually, it is an RPC client, but its details are hidden on the plugin side because it satisfies the Runner interface
type Client struct {
	rpcClient *rpc.Client
}

// NewClient returns a new Client
func NewClient(conn net.Conn) *Client {
	return &Client{rpcClient: rpc.NewClient(conn)}
}

// AttributesRequest is the interface used to communicate via RPC.
type AttributesRequest struct {
	Resource      string
	AttributeName string
}

// AttributesResponse is the interface used to communicate via RPC.
type AttributesResponse struct {
	Attributes []*Attribute
	Err        error
}

// Attribute is an intermediate representation of hcl.Attribute.
// It has an expression as a string of bytes so that hcl.Expression is not transferred via RPC.
type Attribute struct {
	Name      string
	Expr      []byte
	ExprRange hcl.Range
	Range     hcl.Range
	NameRange hcl.Range
}

// WalkResourceAttributes queries the host process, receives a list of attributes that match the conditions,
// and passes each to the walker function.
func (c *Client) WalkResourceAttributes(resource, attributeName string, walker func(*hcl.Attribute) error) error {
	log.Printf("[DEBUG] Walk `%s.*.%s` attribute", resource, attributeName)

	var response AttributesResponse
	if err := c.rpcClient.Call("Plugin.Attributes", AttributesRequest{Resource: resource, AttributeName: attributeName}, &response); err != nil {
		return err
	}
	if response.Err != nil {
		return response.Err
	}

	for _, attribute := range response.Attributes {
		expr, diags := parseExpression(attribute.Expr, attribute.ExprRange.Filename, attribute.ExprRange.Start)
		if diags.HasErrors() {
			return diags
		}
		attr := &hcl.Attribute{
			Name:      attribute.Name,
			Expr:      expr,
			Range:     attribute.Range,
			NameRange: attribute.NameRange,
		}

		if err := walker(attr); err != nil {
			return err
		}
	}

	return nil
}

// BlocksRequest is the interface used to communicate via RPC.
type BlocksRequest struct {
	Resource  string
	BlockName string
}

// BlocksResponse is the interface used to communicate via RPC.
type BlocksResponse struct {
	Blocks []*Block
	Err    error
}

// Block is an intermediate representation of hcl.Block.
// It has an body as a string of bytes so that hcl.Body is not transferred via RPC.
type Block struct {
	Type      string
	Labels    []string
	Body      []byte
	BodyRange hcl.Range

	DefRange    hcl.Range
	TypeRange   hcl.Range
	LabelRanges []hcl.Range
}

// WalkResourceBlocks queries the host process, receives a list of blocks that match the conditions,
// and passes each to the walker function.
func (c *Client) WalkResourceBlocks(resource, blockName string, walker func(*hcl.Block) error) error {
	log.Printf("[DEBUG] Walk `%s.*.%s` block", resource, blockName)

	var response BlocksResponse
	if err := c.rpcClient.Call("Plugin.Blocks", BlocksRequest{Resource: resource, BlockName: blockName}, &response); err != nil {
		return err
	}
	if response.Err != nil {
		return response.Err
	}

	for _, block := range response.Blocks {
		file, diags := parseConfig(block.Body, block.BodyRange.Filename, block.BodyRange.Start)
		if diags.HasErrors() {
			return diags
		}
		b := &hcl.Block{
			Type:        block.Type,
			Labels:      block.Labels,
			Body:        file.Body,
			DefRange:    block.DefRange,
			TypeRange:   block.TypeRange,
			LabelRanges: block.LabelRanges,
		}

		if err := walker(b); err != nil {
			return err
		}
	}

	return nil
}

// ResourcesRequest is the interface used to communicate via RPC.
type ResourcesRequest struct {
	Name string
}

// ResourcesResponse is the interface used to communicate via RPC.
type ResourcesResponse struct {
	Resources []*Resource
	Err       error
}

// Resource is an intermediate representation of configs.Resource.
type Resource struct {
	Name string
	Type string

	DeclRange hcl.Range
	TypeRange hcl.Range
}

// WalkResources queries the host process, receives a list of blocks that match the conditions,
// and passes each to the walker function.
func (c *Client) WalkResources(resource string, walker func(*Resource) error) error {
	log.Printf("[DEBUG] Walk `%s` resource", resource)

	var response ResourcesResponse
	if err := c.rpcClient.Call("Plugin.Resources", ResourcesRequest{Name: resource}, &response); err != nil {
		return err
	}
	if response.Err != nil {
		return response.Err
	}

	for _, block := range response.Resources {
		if err := walker(block); err != nil {
			return err
		}
	}
	return nil
}

// EvalExprRequest is the interface used to communicate via RPC.
type EvalExprRequest struct {
	Expr      []byte
	ExprRange hcl.Range
	Ret       interface{}
}

// EvalExprResponse is the interface used to communicate with RPC.
type EvalExprResponse struct {
	Val cty.Value
	Err error
}

// EvaluateExpr queries the host process for the result of evaluating the value of the passed expression
// and reflects it as the value of the second argument based on that.
func (c *Client) EvaluateExpr(expr hcl.Expression, ret interface{}) error {
	var response EvalExprResponse
	var err error

	src, err := ioutil.ReadFile(expr.Range().Filename)
	if err != nil {
		return err
	}
	req := EvalExprRequest{
		Expr:      expr.Range().SliceBytes(src),
		ExprRange: expr.Range(),
		Ret:       ret,
	}
	if err := c.rpcClient.Call("Plugin.EvalExpr", req, &response); err != nil {
		return err
	}
	if response.Err != nil {
		return response.Err
	}

	err = gocty.FromCtyValue(response.Val, ret)
	if err != nil {
		err := &Error{
			Code:  TypeMismatchError,
			Level: ErrorLevel,
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

// EmitIssueRequest is the interface used to communicate via RPC.
type EmitIssueRequest struct {
	Rule      *RuleObject
	Message   string
	Location  hcl.Range
	Expr      []byte
	ExprRange hcl.Range
}

// EmitIssue emits attributes to build the issue to the host process
// Note that the passed rule need to be converted to generic objects
// because the custom structure defined in the plugin cannot be sent via RPC.
func (c *Client) EmitIssue(rule Rule, message string, location hcl.Range, meta Metadata) error {
	src, err := ioutil.ReadFile(meta.Expr.Range().Filename)
	if err != nil {
		return err
	}

	req := &EmitIssueRequest{
		Rule:      newObjectFromRule(rule),
		Message:   message,
		Location:  location,
		Expr:      meta.Expr.Range().SliceBytes(src),
		ExprRange: meta.Expr.Range(),
	}
	if err := c.rpcClient.Call("Plugin.EmitIssue", &req, new(interface{})); err != nil {
		return err
	}
	return nil
}

// EnsureNoError is a helper for processing when no error occurs
// This function skips processing without returning an error to the caller when the error is warning
func (*Client) EnsureNoError(err error, proc func() error) error {
	if err == nil {
		return proc()
	}

	if appErr, ok := err.(Error); ok {
		switch appErr.Level {
		case WarningLevel:
			return nil
		case ErrorLevel:
			return appErr
		default:
			panic(appErr)
		}
	} else {
		return err
	}
}

func parseExpression(src []byte, filename string, start hcl.Pos) (hcl.Expression, hcl.Diagnostics) {
	if strings.HasSuffix(filename, ".tf") {
		return hclsyntax.ParseExpression(src, filename, start)
	}

	if strings.HasSuffix(filename, ".tf.json") {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "JSON configuration syntax is not supported",
				Subject:  &hcl.Range{Filename: filename, Start: start, End: start},
			},
		}
	}

	panic(fmt.Sprintf("Unexpected file: %s", filename))
}

func parseConfig(src []byte, filename string, start hcl.Pos) (*hcl.File, hcl.Diagnostics) {
	if strings.HasSuffix(filename, ".tf") {
		return hclsyntax.ParseConfig(src, filename, start)
	}

	if strings.HasSuffix(filename, ".tf.json") {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "JSON configuration syntax is not supported",
				Subject:  &hcl.Range{Filename: filename, Start: start, End: start},
			},
		}
	}

	panic(fmt.Sprintf("Unexpected file: %s", filename))
}
