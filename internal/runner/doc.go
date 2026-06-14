// Package runner provides the rule-facing behavior shared by implementations
// of the tflint.Runner interface: callback dispatch and sentinel error
// filtering for EvaluateExpr, block filtering for the GetResourceContent and
// GetProviderContent shorthands, and rule config decoding.
//
// Each implementation supplies only what genuinely varies behind the seam:
// how module content is fetched and how an expression is evaluated into a
// value. The gRPC client delegates both to the host process; the helper
// runner serves them from local files.
package runner
