package client

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func encodeExpr(src []byte, expr hcl.Expression) ([]byte, hcl.Range) {
	return expr.Range().SliceBytes(src), expr.Range()
}

// Rule is an intermediate representation of tflint.Rule.
type Rule struct {
	Data *RuleObject
}

// RuleObject holds the data that Rule needs to satisfy the Rule interface.
type RuleObject struct {
	Name     string
	Enabled  bool
	Severity string
	Link     string
}

// Name is a reference method to internal data.
func (r *Rule) Name() string { return r.Data.Name }

// Enabled is a reference method to internal data.
func (r *Rule) Enabled() bool { return r.Data.Enabled }

// Severity is a reference method to internal data.
func (r *Rule) Severity() string { return r.Data.Severity }

// Link is a reference method to internal data.
func (r *Rule) Link() string { return r.Data.Link }

func encodeRule(rule tflint.RPCRule) *Rule {
	return &Rule{
		Data: &RuleObject{
			Name:     rule.Name(),
			Enabled:  rule.Enabled(),
			Severity: rule.Severity(),
			Link:     rule.Link(),
		},
	}
}
