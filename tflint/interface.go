package tflint

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/schema"
	"github.com/zclconf/go-cty/cty"
)

// RuleSet is a list of rules that a plugin should provide.
// Normally, plugins can use BuiltinRuleSet directly,
// but you can also use custom rulesets that satisfy this interface.
type RuleSet interface {
	// RuleSetName is the name of the ruleset. This method is not expected to be overridden.
	RuleSetName() string

	// RuleSetVersion is the version of the plugin. This method is not expected to be overridden.
	RuleSetVersion() string

	// RuleNames is a list of rule names provided by the plugin. This method is not expected to be overridden.
	RuleNames() []string

	ConfigSchema() *schema.BodySchema

	// ApplyConfig reflects the configuration to the ruleset.
	// Custom rulesets can override this method to reflect the plugin's own configuration.
	// In that case, don't forget to call ApplyCommonConfig.
	ApplyConfig(*Config) error

	// TODO: Do not pass raw body content?
	NewApplyConfig(*schema.BodyContent) error

	// Check runs inspection for each rule by applying Runner.
	// This is a entrypoint for all inspections and can be used as a hook to inject a custom runner.
	Check(Runner) error
}

// Runner acts as a client for each plugin to query the host process about the Terraform configurations.
type Runner interface {
	ResourceContent(string, *schema.BodySchema) (*schema.BodyContent, hcl.Diagnostics)

	// File returns the hcl.File object.
	// This is low level API for accessing information such as comments and syntax.
	// When accessing resources, expressions, etc, it is recommended to use high-level APIs.
	File(string) (*hcl.File, error)

	// Files returns a map[string]hcl.File object, where the key is the file name.
	// This is low level API for accessing information such as comments and syntax.
	Files() (map[string]*hcl.File, error)

	// EvaluateExpr evaluates the passed expression and reflects the result in ret.
	// If you want to ensure the type of ret, you can pass the type as the 3rd argument.
	// If you pass nil as the type, it will be inferred from the type of ret.
	// Since this function returns an application error, it is expected to use the EnsureNoError
	// to determine whether to continue processing.
	EvaluateExpr(expr hcl.Expression, ret interface{}, wantType *cty.Type) error

	// EmitIssue sends an issue to TFLint. You need to pass the message of the issue and the range.
	// You should use EmitIssueOnExpr if you want to emit an issue for an expression.
	// This API provides a lower level interface.
	EmitIssue(rule Rule, message string, location hcl.Range) error
}

// Rule is the interface that the plugin's rules should satisfy.
type Rule interface {
	// Name will be displayed with a message of an issue and will be the identifier used to control
	// the behavior of this rule in the configuration file etc.
	// Therefore, it is expected that this will not duplicate the rule names provided by other plugins.
	Name() string

	// Enabled indicates whether the rule is enabled by default.
	Enabled() bool

	// Severity indicates the severity of the rule.
	Severity() string

	// Link allows you to add a reference link to the rule.
	Link() string

	// Check is the entrypoint of the rule. You can fetch Terraform configurations and send issues via Runner.
	Check(Runner) error
}
