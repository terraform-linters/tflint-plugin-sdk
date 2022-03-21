package tflint

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
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

	// Check runs inspection for each rule by applying Runner.
	// This is a entrypoint for all inspections and can be used as a hook to inject a custom runner.
	Check(Runner) error

	// All Ruleset must embed the builtin ruleset.
	mustEmbedBuiltinRuleSet()
}

// Runner acts as a client for each plugin to query the host process about the Terraform configurations.
// The actual implementation can be found in plugin/plugin2host.GRPCClient.
type Runner interface {
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

	// GetModuleContent retrieves the content of the module based on the passed schema.
	// GetResourceContent is syntactic sugar for GetModuleContent, which you can use to access structures other than resources.
	GetModuleContent(schema *hclext.BodySchema, option *GetModuleContentOption) (*hclext.BodyContent, error)

	// GetFile returns the hcl.File object.
	// This is low level API for accessing information such as comments and syntax.
	// When accessing resources, expressions, etc, it is recommended to use high-level APIs.
	GetFile(filename string) (*hcl.File, error)

	// GetFiles returns a map[string]hcl.File object, where the key is the file name.
	// This is low level API for accessing information such as comments and syntax.
	GetFiles() (map[string]*hcl.File, error)

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

	// EvaluateExpr evaluates the passed expression and reflects the result in the 2nd argument.
	// In addition to the obvious errors, this function returns an error if:
	//   - The expression contains unknown variables (e.g. variables without defaults)
	//   - The expression contains null variables
	//   - The expression contains unevaluable references (e.g. `aws_instance.arn`)
	//
	// To ignore these, use EnsureNoError for the returned error:
	//
	// ```
	// var val string
	// err := runner.EvaluateExpr(expr, &val, nil)
	// err = runner.EnsureNoError(err, func () error {
	//   // Only when no error occurs
	// })
	// if err != nil {
	//   // Only for obvious errors, excluding the above errors
	// }
	// ```
	//
	// The types that can be passed to the 2nd argument are assumed to be as follows:
	//   - string
	//   - int
	//   - []string
	//   - []int
	//   - map[string]string
	//   - map[string]int
	//   - cty.Value
	//
	// Besides this, you can pass a structure. In that case, you need to explicitly pass
	// the type in the option of the 3rd argument:
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
	EvaluateExpr(expr hcl.Expression, ret interface{}, option *EvaluateExprOption) error

	// EmitIssue sends an issue to TFLint. You need to pass the message of the issue and the range.
	EmitIssue(rule Rule, message string, issueRange hcl.Range) error

	// EnsureNoError is a helper for error handling. Depending on the type of error generated by EvaluateExpr,
	// determine whether to exit, skip, or continue. If it is continued, the passed function will be executed.
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
