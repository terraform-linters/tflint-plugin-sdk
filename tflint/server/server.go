package server

import "github.com/terraform-linters/tflint-plugin-sdk/tflint/client"

// Server is the interface that hosts that provide the plugin mechanism must meet in order to respond to queries from the plugin.
type Server interface {
	Attributes(*client.AttributesRequest, *client.AttributesResponse) error
	Blocks(*client.BlocksRequest, *client.BlocksResponse) error
	Resources(*client.ResourcesRequest, *client.ResourcesResponse) error
	ModuleCalls(*client.ModuleCallsRequest, *client.ModuleCallsResponse) error
	Backend(*client.BackendRequest, *client.BackendResponse) error
	EvalExpr(*client.EvalExprRequest, *client.EvalExprResponse) error
	EmitIssue(*client.EmitIssueRequest, *interface{}) error
}
