package tflint

import hcl "github.com/hashicorp/hcl/v2"

// List of issue severity levels. The rules implemented by a plugin can be set to any severity.
const (
	// ERROR is possible errors
	ERROR = "Error"
	// WARNING doesn't cause problem immediately, but not good
	WARNING = "Warning"
	// NOTICE is not important, it's mentioned
	NOTICE = "Notice"
)

// Metadata is an additional data sent to the host process to build the issue.
type Metadata struct {
	Expr hcl.Expression
}
