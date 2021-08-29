package tflint

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/schema"
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
