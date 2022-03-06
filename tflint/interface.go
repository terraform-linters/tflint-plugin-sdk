package tflint

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
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

	// ConfigSchema returns the ruleset plugin config schema.
	// This schema should be a schema inside of "plugin" block.
	ConfigSchema() *hclext.BodySchema

	// ApplyGlobalConfig applies the common config to the ruleset.
	// This is not supposed to be overridden from custom rulesets.
	// Override the ApplyConfig if you want to apply the plugin's own configuration.
	ApplyGlobalConfig(*Config) error

	// ApplyConfig applies the configuration to the ruleset.
	// Custom rulesets can override this method to reflect the plugin's own configuration.
	ApplyConfig(*hclext.BodyContent) error

	// Check runs inspection for each rule by applying Runner.
	// This is a entrypoint for all inspections and can be used as a hook to inject a custom runner.
	Check(Runner) error

	// All Ruleset must embed the builtin ruleset.
	mustEmbedBuiltinRuleSet()
}

// Runner acts as a client for each plugin to query the host process about the Terraform configurations.
type Runner interface {
	GetResourceContent(string, *hclext.BodySchema, *GetModuleContentOption) (*hclext.BodyContent, error)
	GetModuleContent(*hclext.BodySchema, *GetModuleContentOption) (*hclext.BodyContent, error)
	GetFile(string) (*hcl.File, error)
	GetFiles() (map[string]*hcl.File, error)
	DecodeRuleConfig(name string, ret interface{}) error
	EvaluateExpr(hcl.Expression, interface{}, *EvaluateExprOption) error
	EmitIssue(Rule, string, hcl.Range) error
	EnsureNoError(error, func() error) error
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
	Severity() Severity

	// Link allows you to add a reference link to the rule.
	Link() string

	// Metadata allows you to set any metadata to the rule.
	// This value is never referenced by the SDK and can be used for your custom ruleset.
	Metadata() interface{}

	// Check is the entrypoint of the rule. You can fetch Terraform configurations and send issues via Runner.
	Check(Runner) error

	// All rules must embed the default rule.
	mustEmbedDefaultRule()
}

// RPCRuleSet is a list of rules that a plugin should provide.
// Normally, plugins can use BuiltinRuleSet directly,
// but you can also use custom rulesets that satisfy this interface.
type RPCRuleSet interface {
	// RuleSetName is the name of the ruleset. This method is not expected to be overridden.
	RuleSetName() string

	// RuleSetVersion is the version of the plugin. This method is not expected to be overridden.
	RuleSetVersion() string

	// RuleNames is a list of rule names provided by the plugin. This method is not expected to be overridden.
	RuleNames() []string

	// ApplyConfig reflects the configuration to the ruleset.
	// Custom rulesets can override this method to reflect the plugin's own configuration.
	// In that case, don't forget to call ApplyCommonConfig.
	ApplyConfig(*Config) error

	// Check runs inspection for each rule by applying Runner.
	// This is a entrypoint for all inspections and can be used as a hook to inject a custom runner.
	Check(RPCRunner) error
}

// RPCRunner acts as a client for each plugin to query the host process about the Terraform configurations.
type RPCRunner interface {
	// WalkResourceAttributes visits attributes with the passed function.
	// You must pass a resource type as the first argument and an attribute name as the second argument.
	WalkResourceAttributes(string, string, func(*hcl.Attribute) error) error

	// WalkResourceBlocks visits blocks with the passed function.
	// You must pass a resource type as the first argument and a block type as the second argument.
	// This API currently does not support labeled blocks.
	WalkResourceBlocks(string, string, func(*hcl.Block) error) error

	// WalkResources visits resources with the passed function.
	// You must pass a resource type as the first argument.
	WalkResources(string, func(*configs.Resource) error) error

	// WalkModuleCalls visits module calls with the passed function.
	WalkModuleCalls(func(*configs.ModuleCall) error) error

	// Backend returns the backend configuration, if any.
	Backend() (*configs.Backend, error)

	// Config returns the Terraform configuration.
	// This object contains almost all accessible data structures from plugins.
	Config() (*configs.Config, error)

	// File returns the hcl.File object.
	// This is low level API for accessing information such as comments and syntax.
	// When accessing resources, expressions, etc, it is recommended to use high-level APIs.
	File(string) (*hcl.File, error)

	// Files returns a map[string]hcl.File object, where the key is the file name.
	// This is low level API for accessing information such as comments and syntax.
	Files() (map[string]*hcl.File, error)

	// RootProvider returns the provider configuration in the root module.
	// It can be used by child modules to access the credentials defined in the root module.
	RootProvider(name string) (*configs.Provider, error)

	// DecodeRuleConfig fetches the rule's configuration and reflects the result in ret.
	DecodeRuleConfig(name string, ret interface{}) error

	// EvaluateExpr evaluates the passed expression and reflects the result in ret.
	// If you want to ensure the type of ret, you can pass the type as the 3rd argument.
	// If you pass nil as the type, it will be inferred from the type of ret.
	// Since this function returns an application error, it is expected to use the EnsureNoError
	// to determine whether to continue processing.
	EvaluateExpr(expr hcl.Expression, ret interface{}, wantType *cty.Type) error

	// EvaluateExprOnRootCtx is the equivalent of EvaluateExpr method in the context of the root module.
	// Its main use is to evaluate the provider block obtained by the RootProvider method.
	EvaluateExprOnRootCtx(expr hcl.Expression, ret interface{}, wantType *cty.Type) error

	// IsNullExpr checks whether the passed expression is null or not.
	// This returns an error when the passed expression is invalid, occurs evaluation errors, etc.
	IsNullExpr(expr hcl.Expression) (bool, error)

	// EmitIssue sends an issue with an expression to TFLint. You need to pass the message of the issue and the expression.
	EmitIssueOnExpr(rule RPCRule, message string, expr hcl.Expression) error

	// EmitIssue sends an issue to TFLint. You need to pass the message of the issue and the range.
	// You should use EmitIssueOnExpr if you want to emit an issue for an expression.
	// This API provides a lower level interface.
	EmitIssue(rule RPCRule, message string, location hcl.Range) error

	// EnsureNoError is a helper for error handling. Depending on the type of error generated by EvaluateExpr,
	// determine whether to exit, skip, or continue. If it is continued, the passed function will be executed.
	EnsureNoError(error, func() error) error
}

// RPCRule is the interface that the plugin's rules should satisfy.
type RPCRule interface {
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
	Check(RPCRunner) error
}
