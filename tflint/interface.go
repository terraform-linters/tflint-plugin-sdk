package tflint

import (
	"github.com/hashicorp/hcl/v2"
)

// Runner acts as a client for each plugin to query the host process about the Terraform configurations.
type Runner interface {
	WalkResourceAttributes(string, string, func(*hcl.Attribute) error) error
	WalkResourceBlocks(string, string, func(*hcl.Block) error) error
	EvaluateExpr(expr hcl.Expression, ret interface{}) error
	EmitIssue(rule Rule, message string, location hcl.Range, meta Metadata) error
	EnsureNoError(error, func() error) error
}

// Rule is the interface that the plugin's rules should satisfy.
type Rule interface {
	Name() string
	Enabled() bool
	Severity() string
	Link() string
	Check(Runner) error
}

// Server is the interface that hosts that provide the plugin mechanism must meet in order to respond to queries from the plugin.
type Server interface {
	Attributes(*AttributesRequest, *AttributesResponse) error
	EvalExpr(*EvalExprRequest, *EvalExprResponse) error
	EmitIssue(*EmitIssueRequest, *interface{}) error
}
