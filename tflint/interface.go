package tflint

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
)

// RuleSet is a list of rules that a plugin should provide.
// Normally, plugins can use BuiltinRuleSet directly,
// but you can also use custom rulesets that satisfy this interface.
// The actual implementation can be found in plugin/host2plugin.GRPCServer.
type RuleSet interface {
	// RuleSetName is the name of the ruleset. This method is not expected to be overridden.
	RuleSetName() string

	// RuleSetVersion is the version of the plugin. This method is not expected to be overridden.
	RuleSetVersion() string

	// RuleNames is a list of rule names provided by the plugin. This method is not expected to be overridden.
	RuleNames() []string

	// VersionConstraint declares the version of TFLint the plugin will work with. Default is no constraint.
	VersionConstraint() string

	// ConfigSchema returns the ruleset plugin config schema.
	// If you return a schema, TFLint will extract the config from .tflint.hcl based on that schema
	// and pass it to ApplyConfig. This schema should be a schema inside of "plugin" block.
	// If you don't need a config that controls the entire plugin, you don't need to override this method.
	//
	// It is recommended to use hclext.ImpliedBodySchema to generate the schema from the structure:
	//
	// ```
	// type myPluginConfig struct {
	//   Style       string `hclext:"style"`
	//   Description string `hclext:"description,optional"`
	//   Detail      Detail `hclext:"detail,block"`
	// }
	//
	// config := &myPluginConfig{}
	// hclext.ImpliedBodySchema(config)
	// ```
	ConfigSchema() *hclext.BodySchema

	// ApplyGlobalConfig applies the common config to the ruleset.
	// This is not supposed to be overridden from custom rulesets.
	// Override the ApplyConfig if you want to apply the plugin's custom configuration.
	ApplyGlobalConfig(*Config) error

	// ApplyConfig applies the configuration to the ruleset.
	// Custom rulesets can override this method to reflect the plugin's custom configuration.
	//
	// You can reflect the body in the structure by using hclext.DecodeBody:
	//
	// ```
	// type myPluginConfig struct {
	//   Style       string `hclext:"style"`
	//   Description string `hclext:"description,optional"`
	//   Detail      Detail `hclext:"detail,block"`
	// }
	//
	// config := &myPluginConfig{}
	// hclext.DecodeBody(body, nil, config)
	// ```
	ApplyConfig(*hclext.BodyContent) error

	// NewRunner returns a new runner based on the original runner.
	// Custom rulesets can override this method to inject a custom runner.
	NewRunner(Runner) (Runner, error)

	// Check is a entrypoint for all inspections.
	// This is not supposed to be overridden from custom rulesets.
	Check(Runner) error

	// All Ruleset must embed the builtin ruleset.
	mustEmbedBuiltinRuleSet()
}

// Runner acts as a client for each plugin to query the host process about the Terraform configurations.
// The actual implementation can be found in plugin/plugin2host.GRPCClient.
type Runner interface {
	// GetOriginalwd returns the original working directory.
	// Normally this is equal to os.Getwd(), but differs if --chdir or --recursive is used.
	// If you need the absolute path of the file, joining with the original working directory is appropriate.
	GetOriginalwd() (string, error)

	// GetModulePath returns the current module path address.
	GetModulePath() (addrs.Module, error)

	// GetResourceContent retrieves the content of resources based on the passed schema.
	// The schema allows you to specify attributes and blocks that describe the structure needed for the inspection:
	//
	// ```
	// runner.GetResourceContent("aws_instance", &hclext.BodySchema{
	//   Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
	//   Blocks: []hclext.BlockSchema{
	//     {
	//       Type: "ebs_block_device",
	//       Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "volume_size"}}},
	//     },
	//   },
	// }, nil)
	// ```
	GetResourceContent(resourceName string, schema *hclext.BodySchema, option *GetModuleContentOption) (*hclext.BodyContent, error)

	// GetProviderContent retrieves the content of providers based on the passed schema.
	// This method is GetResourceContent for providers.
	GetProviderContent(providerName string, schema *hclext.BodySchema, option *GetModuleContentOption) (*hclext.BodyContent, error)

	// GetModuleContent retrieves the content of the module based on the passed schema.
	// GetResourceContent/GetProviderContent are syntactic sugar for GetModuleContent, which you can use to access other structures.
	GetModuleContent(schema *hclext.BodySchema, option *GetModuleContentOption) (*hclext.BodyContent, error)

	// GetFile returns the hcl.File object.
	// This is low level API for accessing information such as comments and syntax.
	// When accessing resources, expressions, etc, it is recommended to use high-level APIs.
	GetFile(filename string) (*hcl.File, error)

	// GetFiles returns a map[string]hcl.File object, where the key is the file name.
	// This is low level API for accessing information such as comments and syntax.
	GetFiles() (map[string]*hcl.File, error)

	// WalkExpressions traverses expressions in all files by the passed walker.
	// The walker can be passed any structure that satisfies the `tflint.ExprWalker`
	// interface, or a `tflint.ExprWalkFunc`. Example of passing function:
	//
	// ```
	// runner.WalkExpressions(tflint.ExprWalkFunc(func (expr hcl.Expression) hcl.Diagnostics {
	//   // Write code here
	// }))
	// ```
	//
	// If you pass ExprWalkFunc, the function will be called for every expression.
	// Note that it behaves differently in native HCL syntax and JSON syntax.
	//
	// In the HCL syntax, `var.foo` and `var.bar` in `[var.foo, var.bar]` are
	// also passed to the walker. In other words, it traverses expressions recursively.
	// To avoid redundant checks, the walker should check the kind of expression.
	//
	// In the JSON syntax, only an expression of an attribute seen from the top
	// level of the file is passed. In other words, it doesn't traverse expressions
	// recursively. This is a limitation of JSON syntax.
	WalkExpressions(walker ExprWalker) hcl.Diagnostics

	// DecodeRuleConfig fetches the rule's configuration and reflects the result in the 2nd argument.
	// The argument is expected to be a pointer to a structure tagged with hclext:
	//
	// ```
	// type myRuleConfig struct {
	//   Style       string `hclext:"style"`
	//   Description string `hclext:"description,optional"`
	//   Detail      Detail `hclext:"detail,block"`
	// }
	//
	// config := &myRuleConfig{}
	// runner.DecodeRuleConfig("my_rule", config)
	// ```
	//
	// See the hclext.DecodeBody documentation and examples for more details.
	DecodeRuleConfig(ruleName string, ret interface{}) error

	// EvaluateExpr evaluates the given expression and assigns the value to the Go value target.
	// The target must be a pointer. Otherwise, it will cause a panic.
	//
	// If the value cannot be assigned to the target, it returns an error.
	// There are particularly examples such as:
	//
	//   1. Unknown value (e.g. variables without defaults, `aws_instance.foo.arn`)
	//   2. NULL
	//   3. Sensitive value (variables with `sensitive = true`)
	//
	// However, if the target is cty.Value, these errors will not be returned.
	// These errors can be handled with errors.Is().
	//
	// ```
	// var val string
	// err := runner.EvaluateExpr(expr, &val, nil)
	// if err != nil {
	//   if errors.Is(err, tflint.ErrUnknownValue) {
	//     // Ignore unknown values
	//	   return nil
	//   }
	//   if errors.Is(err, tflint.ErrNullValue) {
	//     // Ignore null values because null means that the value is not set
	//	   return nil
	//   }
	//   if errors.Is(err, tflint.ErrSensitive) {
	//     // Ignore sensitive values
	//     return nil
	//   }
	//   return err
	// }
	// ```
	//
	// The following are the types that can be passed as the target:
	//
	//   1. string
	//   2. int
	//   3. []string
	//   4. []int
	//   5. map[string]string
	//   6. map[string]int
	//   7. cty.Value
	//   8. func (v T) error
	//
	// Passing any other type will cause a panic. If you pass a function, the assigned value
	// will be used as an argument to execute the function. In this case, if a value cannot be
	// assigned to the argument type, instead of returning an error, the execution is skipped.
	// This is useful when it is always acceptable to ignore exceptional values.
	//
	// ```
	// runner.EvaluateExpr(expr, func (val string) error {
	//   // Test value
	// }, nil)
	// ```
	//
	// Besides this, you can pass a structure. In that case, you need to explicitly pass wantType.
	//
	// ```
	// type complexVal struct {
	//   Key     string `cty:"key"`
	//   Enabled bool   `cty:"enabled"`
	// }
	//
	// wantType := cty.List(cty.Object(map[string]cty.Type{
	//   "key":     cty.String,
	//   "enabled": cty.Bool,
	// }))
	//
	// var complexVals []complexVal
	// runner.EvaluateExpr(expr, &compleVals, &tflint.EvaluateExprOption{WantType: &wantType})
	// ```
	EvaluateExpr(expr hcl.Expression, target interface{}, option *EvaluateExprOption) error

	// EmitIssue sends an issue to TFLint. You need to pass the message of the issue and the range.
	EmitIssue(rule Rule, message string, issueRange hcl.Range) error

	// EnsureNoError is a helper for error handling. Depending on the type of error generated by EvaluateExpr,
	// determine whether to exit, skip, or continue. If it is continued, the passed function will be executed.
	//
	// Deprecated: Use EvaluateExpr with a function callback. e.g. EvaluateExpr(expr, func (val T) error {}, ...)
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
