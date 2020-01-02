package tflint

import hcl "github.com/hashicorp/hcl/v2"

const (
	// ERROR is possible errors
	ERROR = "Error"
	// WARNING doesn't cause problem immediately, but not good
	WARNING = "Warning"
	// NOTICE is not important, it's mentioned
	NOTICE = "Notice"
)

// Metadata is the additional data sent to the host process to build the issue.
type Metadata struct {
	Expr hcl.Expression
}

// RuleObject is an intermediate representation for communicating with RPC.
type RuleObject struct {
	Data *RuleObjectData
}

// RuleObjectData holds the data that RuleObject needs to satisfy the Rule interface.
type RuleObjectData struct {
	Name     string
	Enabled  bool
	Severity string
	Link     string
}

func newObjectFromRule(rule Rule) *RuleObject {
	return &RuleObject{
		Data: &RuleObjectData{
			Name:     rule.Name(),
			Enabled:  rule.Enabled(),
			Severity: rule.Severity(),
			Link:     rule.Link(),
		},
	}
}

// Name is a reference method to internal data
func (r *RuleObject) Name() string { return r.Data.Name }

// Enabled is a reference method to internal data
func (r *RuleObject) Enabled() bool { return r.Data.Enabled }

// Severity is a reference method to internal data
func (r *RuleObject) Severity() string { return r.Data.Severity }

// Link is a reference method to internal data
func (r *RuleObject) Link() string { return r.Data.Link }
