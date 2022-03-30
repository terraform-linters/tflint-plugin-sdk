package helper

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Runner is a mock that satisfies the Runner interface for plugin testing.
type Runner struct {
	Issues Issues

	files     map[string]*hcl.File
	config    Config
	variables map[string]*Variable
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

// GetFile returns the hcl.File object
func (r *Runner) GetFile(filename string) (*hcl.File, error) {
	return r.files[filename], nil
}

// GetFiles returns all hcl.File
func (r *Runner) GetFiles() (map[string]*hcl.File, error) {
	return r.files, nil
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

	return fmt.Errorf("rule `%s` not found", name)
}

// EvaluateExpr returns a value of the passed expression.
// Note that some features are limited
func (r *Runner) EvaluateExpr(expr hcl.Expression, ret interface{}, opts *tflint.EvaluateExprOption) error {
	if opts == nil {
		opts = &tflint.EvaluateExprOption{}
	}

	var ty cty.Type
	if opts.WantType != nil {
		ty = *opts.WantType
	} else {
		switch ret.(type) {
		case *string, string:
			ty = cty.String
		case *int, int:
			ty = cty.Number
		case *[]string, []string:
			ty = cty.List(cty.String)
		case *[]int, []int:
			ty = cty.List(cty.Number)
		case *map[string]string, map[string]string:
			ty = cty.Map(cty.String)
		case *map[string]int, map[string]int:
			ty = cty.Map(cty.Number)
		case cty.Value, *cty.Value:
			ty = cty.DynamicPseudoType
		default:
			return fmt.Errorf("unsupported result type: %T", ret)
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

	return gocty.FromCtyValue(val, ret)
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

// EnsureNoError is a method that simply runs a function if there is no error.
func (r *Runner) EnsureNoError(err error, proc func() error) error {
	if err == nil {
		return proc()
	}
	return err
}

// NewLocalRunner initialises a new test runner.
// Internal use only.
func NewLocalRunner(files map[string]*hcl.File, issues Issues) *Runner {
	return &Runner{files: map[string]*hcl.File{}, variables: map[string]*Variable{}, Issues: issues}
}

// AddLocalFile adds a new file to the current mapped files.
// Internal use only.
func (r *Runner) AddLocalFile(name string, file *hcl.File) bool {
	if _, exists := r.files[name]; exists {
		return false
	}

	r.files[name] = file
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
