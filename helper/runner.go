package helper

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/internal"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"
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
	content := &hclext.BodyContent{}
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
	body, err := r.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}, Body: schema},
		},
	}, opts)
	if err != nil {
		return nil, err
	}

	content := &hclext.BodyContent{Blocks: []*hclext.Block{}}
	for _, resource := range body.Blocks {
		if resource.Labels[0] != name {
			continue
		}

		content.Blocks = append(content.Blocks, resource)
	}

	return content, nil
}

// GetProviderContent gets a provider content of the current module
func (r *Runner) GetProviderContent(name string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	body, err := r.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{Type: "provider", LabelNames: []string{"name"}, Body: schema},
		},
	}, opts)
	if err != nil {
		return nil, err
	}

	content := &hclext.BodyContent{Blocks: []*hclext.Block{}}
	for _, provider := range body.Blocks {
		if provider.Labels[0] != name {
			continue
		}

		content.Blocks = append(content.Blocks, provider)
	}

	return content, nil
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
	schema := hclext.ImpliedBodySchema(ret)

	for _, rule := range r.config.Rules {
		if rule.Name == name {
			body, diags := hclext.Content(rule.Body, schema)
			if diags.HasErrors() {
				return diags
			}
			if diags := hclext.DecodeBody(body, nil, ret); diags.HasErrors() {
				return diags
			}
			return nil
		}
	}

	return nil
}

var errRefTy = reflect.TypeOf((*error)(nil)).Elem()

// EvaluateExpr returns a value of the passed expression.
// Note that some features are limited
func (r *Runner) EvaluateExpr(expr hcl.Expression, target interface{}, opts *tflint.EvaluateExprOption) error {
	rval := reflect.ValueOf(target)
	rty := rval.Type()

	var callback bool
	switch rty.Kind() {
	case reflect.Func:
		// Callback must meet the following requirements:
		//   - It must be a function
		//   - It must take an argument
		//   - It must return an error
		if !(rty.NumIn() == 1 && rty.NumOut() == 1 && rty.Out(0).Implements(errRefTy)) {
			panic(`callback must be of type "func (v T) error"`)
		}
		callback = true
		target = reflect.New(rty.In(0)).Interface()

	case reflect.Pointer:
		// ok
	default:
		panic("target value is not a pointer or function")
	}

	err := r.evaluateExpr(expr, target, opts)
	if !callback {
		// error should be handled in the caller
		return err
	}

	if err != nil {
		// If it cannot be represented as a Go value, exit without invoking the callback rather than returning an error.
		if errors.Is(err, tflint.ErrUnknownValue) || errors.Is(err, tflint.ErrNullValue) || errors.Is(err, tflint.ErrSensitive) || errors.Is(err, tflint.ErrUnevaluable) {
			return nil
		}
		return err
	}

	rerr := rval.Call([]reflect.Value{reflect.ValueOf(target).Elem()})
	if rerr[0].IsNil() {
		return nil
	}
	return rerr[0].Interface().(error)
}

func (r *Runner) evaluateExpr(expr hcl.Expression, target interface{}, opts *tflint.EvaluateExprOption) error {
	if opts == nil {
		opts = &tflint.EvaluateExprOption{}
	}

	var ty cty.Type
	if opts.WantType != nil {
		ty = *opts.WantType
	} else {
		switch target.(type) {
		case *string:
			ty = cty.String
		case *int:
			ty = cty.Number
		case *bool:
			ty = cty.Bool
		case *[]string:
			ty = cty.List(cty.String)
		case *[]int:
			ty = cty.List(cty.Number)
		case *[]bool:
			ty = cty.List(cty.Bool)
		case *map[string]string:
			ty = cty.Map(cty.String)
		case *map[string]int:
			ty = cty.Map(cty.Number)
		case *map[string]bool:
			ty = cty.Map(cty.Bool)
		case *cty.Value:
			ty = cty.DynamicPseudoType
		default:
			return fmt.Errorf("unsupported target type: %T", target)
		}
	}

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

	return gocty.FromCtyValue(val, target)
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
	if err == nil {
		return proc()
	}
	return err
}

// NewLocalRunner initialises a new test runner.
// Internal use only.
func NewLocalRunner(files map[string]*hcl.File, issues Issues) *Runner {
	return &Runner{
		files:     map[string]*hcl.File{},
		sources:   map[string][]byte{},
		variables: map[string]*Variable{},
		Issues:    issues,
	}
}

// AddLocalFile adds a new file to the current mapped files.
// Internal use only.
func (r *Runner) AddLocalFile(name string, file *hcl.File) bool {
	if _, exists := r.files[name]; exists {
		return false
	}

	r.files[name] = file
	r.sources[name] = file.Bytes
	return true
}

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
