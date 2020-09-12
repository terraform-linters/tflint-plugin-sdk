package client

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// AttributesRequest is a request to the server-side Attributes method.
type AttributesRequest struct {
	Resource      string
	AttributeName string
}

// AttributesResponse is a response to the server-side Attributes method.
type AttributesResponse struct {
	Attributes []*Attribute
	Err        error
}

// BackendRequest is a request to the server-side Backend method.
type BackendRequest struct{}

// BackendResponse is a response to the server-side Backend method.
type BackendResponse struct {
	Backend *Backend
	Err     error
}

// BlocksRequest is a request to the server-side Blocks method.
type BlocksRequest struct {
	Resource  string
	BlockType string
}

// BlocksResponse is a response to the server-side Blocks method.
type BlocksResponse struct {
	Blocks []*Block
	Err    error
}

// ModuleCallsRequest is a request to the server-side ModuleCalls method.
type ModuleCallsRequest struct{}

// ModuleCallsResponse is a response to the server-side ModuleCalls method.
type ModuleCallsResponse struct {
	ModuleCalls []*ModuleCall
	Err         error
}

// ResourcesRequest is a request to the server-side Resources method.
type ResourcesRequest struct {
	Name string
}

// ResourcesResponse is a response to the server-side Resources method.
type ResourcesResponse struct {
	Resources []*Resource
	Err       error
}

// ConfigRequest is a request to the server-side Config method.
type ConfigRequest struct{}

// ConfigResponse is a response to the server-side Config method.
type ConfigResponse struct {
	Config *Config
	Err    error
}

// EvalExprRequest is a request to the server-side EvalExpr method.
type EvalExprRequest struct {
	Expr      []byte
	ExprRange hcl.Range
	Ret       interface{}
}

// EvalExprResponse is a response to the server-side EvalExpr method.
type EvalExprResponse struct {
	Val cty.Value
	Err error
}

// EmitIssueRequest is a request to the server-side EmitIssue method.
type EmitIssueRequest struct {
	Rule      *Rule
	Message   string
	Location  hcl.Range
	Expr      []byte
	ExprRange hcl.Range
}
