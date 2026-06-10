package helper

import (
	"errors"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/internal"
	"github.com/terraform-linters/tflint-plugin-sdk/internal/runner"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// Runner is a mock that satisfies the Runner interface for plugin testing.
type Runner struct {
	Issues Issues

	files     map[string]*hcl.File
	sources   map[string][]byte
	config    Config
	variables map[string]*Variable
	fixer     *internal.Fixer
}

// Variable is an implementation of variables in Terraform language
type Variable struct {
	Name      string
	Default   cty.Value
	DeclRange hcl.Range
}

// Config is a pseudo TFLint config file object for testing from plugins.
type Config struct {
	Rules []RuleConfig `hcl:"rule,block"`
}

// RuleConfig is a pseudo TFLint config file object for testing from plugins.
type RuleConfig struct {
	Name    string   `hcl:"name,label"`
	Enabled bool     `hcl:"enabled"`
	Body    hcl.Body `hcl:",remain"`
}

var _ tflint.Runner = &Runner{}

// GetOriginalwd always returns the current directory
func (r *Runner) GetOriginalwd() (string, error) {
	return os.Getwd()
}

// GetModulePath always returns the root module path address
func (r *Runner) GetModulePath() (addrs.Module, error) {
	return []string{}, nil
}

// GetModuleContent gets a content of the current module
func (r *Runner) GetModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	content := &hclext.BodyContent{
		Attributes: hclext.Attributes{},
		Blocks:     hclext.Blocks{},
	}
	diags := hcl.Diagnostics{}

	for _, f := range r.files {
		c, d := hclext.PartialContent(f.Body, schema)
		diags = diags.Extend(d)
		for name, attr := range c.Attributes {
			content.Attributes[name] = attr
		}
		content.Blocks = append(content.Blocks, c.Blocks...)
	}

	if diags.HasErrors() {
		return nil, diags
	}
	return content, nil
}

// GetResourceContent gets a resource content of the current module
func (r *Runner) GetResourceContent(name string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	return runner.GetResourceContent(r, name, schema, opts)
}

// GetProviderContent gets a provider content of the current module
func (r *Runner) GetProviderContent(name string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	return runner.GetProviderContent(r, name, schema, opts)
}

// GetFile returns the hcl.File object
func (r *Runner) GetFile(filename string) (*hcl.File, error) {
	return r.files[filename], nil
}

// GetFiles returns all hcl.File
func (r *Runner) GetFiles() (map[string]*hcl.File, error) {
	return r.files, nil
}

type nativeWalker struct {
	walker tflint.ExprWalker
}

func (w *nativeWalker) Enter(node hclsyntax.Node) hcl.Diagnostics {
	if expr, ok := node.(hcl.Expression); ok {
		return w.walker.Enter(expr)
	}
	return nil
}

func (w *nativeWalker) Exit(node hclsyntax.Node) hcl.Diagnostics {
	if expr, ok := node.(hcl.Expression); ok {
		return w.walker.Exit(expr)
	}
	return nil
}

// WalkExpressions traverses expressions in all files by the passed walker.
func (r *Runner) WalkExpressions(walker tflint.ExprWalker) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	for _, file := range r.files {
		if body, ok := file.Body.(*hclsyntax.Body); ok {
			walkDiags := hclsyntax.Walk(body, &nativeWalker{walker: walker})
			diags = diags.Extend(walkDiags)
			continue
		}

		// In JSON syntax, everything can be walked as an attribute.
		attrs, jsonDiags := file.Body.JustAttributes()
		if jsonDiags.HasErrors() {
			diags = diags.Extend(jsonDiags)
			continue
		}

		for _, attr := range attrs {
			enterDiags := walker.Enter(attr.Expr)
			diags = diags.Extend(enterDiags)
			exitDiags := walker.Exit(attr.Expr)
			diags = diags.Extend(exitDiags)
		}
	}

	return diags
}

// DecodeRuleConfig extracts the rule's configuration into the given value
func (r *Runner) DecodeRuleConfig(name string, ret interface{}) error {
	return runner.DecodeRuleConfig(ret, func(schema *hclext.BodySchema) (*hclext.BodyContent, error) {
		for _, rule := range r.config.Rules {
			if rule.Name == name {
				body, diags := hclext.Content(rule.Body, schema)
				if diags.HasErrors() {
					return nil, diags
				}
				return body, nil
			}
		}
		return nil, nil
	})
}

// EvaluateExpr returns a value of the passed expression.
// Note that some features are limited
func (r *Runner) EvaluateExpr(expr hcl.Expression, target interface{}, opts *tflint.EvaluateExprOption) error {
	return runner.EvaluateExpr(expr, target, opts, r.evaluateExpr)
}

func (r *Runner) evaluateExpr(expr hcl.Expression, target interface{}, opts *tflint.EvaluateExprOption) error {
	ty := runner.WantType(target, opts)

	variables := map[string]cty.Value{}
	for _, variable := range r.variables {
		variables[variable.Name] = variable.Default
	}
	workspace, success := os.LookupEnv("TF_WORKSPACE")
	if !success {
		workspace = "default"
	}
	rawVal, diags := expr.Value(&hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var": cty.ObjectVal(variables),
			"terraform": cty.ObjectVal(map[string]cty.Value{
				"workspace": cty.StringVal(workspace),
			}),
		},
	})
	if diags.HasErrors() {
		return diags
	}
	val, err := convert.Convert(rawVal, ty)
	if err != nil {
		return err
	}

	return runner.DecodeValue(val, ty, expr.Range(), target)
}

// EmitIssue adds an issue to the runner itself.
func (r *Runner) EmitIssue(rule tflint.Rule, message string, location hcl.Range) error {
	r.Issues = append(r.Issues, &Issue{
		Rule:    rule,
		Message: message,
		Range:   location,
	})
	return nil
}

// EmitIssueWithFix adds an issue and invoke fix.
func (r *Runner) EmitIssueWithFix(rule tflint.Rule, message string, location hcl.Range, fixFunc func(f tflint.Fixer) error) error {
	r.fixer.StashChanges()
	if err := fixFunc(r.fixer); err != nil {
		if errors.Is(err, tflint.ErrFixNotSupported) {
			r.fixer.PopChangesFromStash()
			return r.EmitIssue(rule, message, location)
		}
		return err
	}
	return r.EmitIssue(rule, message, location)
}

// Changes returns formatted changes by the fixer.
func (r *Runner) Changes() map[string][]byte {
	r.fixer.FormatChanges()
	return r.fixer.Changes()
}

// EnsureNoError is a method that simply runs a function if there is no error.
//
// Deprecated: Use EvaluateExpr with a function callback. e.g. EvaluateExpr(expr, func (val T) error {}, ...)
func (r *Runner) EnsureNoError(err error, proc func() error) error {
	return runner.EnsureNoError(err, proc)
}

// newLocalRunner initialises a new test runner.
func newLocalRunner(files map[string]*hcl.File, issues Issues) *Runner {
	return &Runner{
		files:     map[string]*hcl.File{},
		sources:   map[string][]byte{},
		variables: map[string]*Variable{},
		Issues:    issues,
	}
}

// addLocalFile adds a new file to the current mapped files.
// For testing only. Normally, the main TFLint process is responsible for loading files.
func (r *Runner) addLocalFile(name string, file *hcl.File) bool {
	if _, exists := r.files[name]; exists {
		return false
	}

	r.files[name] = file
	r.sources[name] = file.Bytes
	return true
}

// initFromFiles initializes the runner from locally added files.
// For testing only.
func (r *Runner) initFromFiles() error {
	for _, file := range r.files {
		content, _, diags := file.Body.PartialContent(configFileSchema)
		if diags.HasErrors() {
			return diags
		}

		for _, block := range content.Blocks {
			switch block.Type {
			case "variable":
				variable, diags := decodeVariableBlock(block)
				if diags.HasErrors() {
					return diags
				}
				r.variables[variable.Name] = variable
			default:
				continue
			}
		}
	}
	r.fixer = internal.NewFixer(r.sources)

	return nil
}

func decodeVariableBlock(block *hcl.Block) (*Variable, hcl.Diagnostics) {
	v := &Variable{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	content, _, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "default",
			},
			{
				Name: "sensitive",
			},
			{
				Name: "ephemeral",
			},
		},
	})
	if diags.HasErrors() {
		return v, diags
	}

	if attr, exists := content.Attributes["default"]; exists {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return v, diags
		}

		v.Default = val
	} else {
		v.Default = cty.DynamicVal
	}
	if attr, exists := content.Attributes["sensitive"]; exists {
		var sensitive bool
		diags := gohcl.DecodeExpression(attr.Expr, nil, &sensitive)
		if diags.HasErrors() {
			return v, diags
		}

		if sensitive {
			v.Default = v.Default.Mark(marks.Sensitive)
		}
	}
	if attr, exists := content.Attributes["ephemeral"]; exists {
		var ephemeral bool
		diags := gohcl.DecodeExpression(attr.Expr, nil, &ephemeral)
		if diags.HasErrors() {
			return v, diags
		}

		if ephemeral {
			v.Default = v.Default.Mark(marks.Ephemeral)
		}
	}

	return v, nil
}

var configFileSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "variable",
			LabelNames: []string{"name"},
		},
	},
}
