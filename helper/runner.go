package helper

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Runner is a pseudo Runner client provided for plugin testing.
// Actually, it is provided as an RPC client, but for the sake of simplicity,
// only the methods that satisfy the minimum required Runner interface are implemented.
// Specifically, there are restrictions on evaluation, annotation comments, module inspection, and so on.
type Runner struct {
	Files  map[string]*hcl.File
	Issues Issues
}

// WalkResourceAttributes searches for resources and passes the appropriate attributes to the walker function
func (r *Runner) WalkResourceAttributes(resourceType, attributeName string, walker func(*hcl.Attribute) error) error {
	for _, file := range r.Files {
		resources, _, diags := file.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "resource",
					LabelNames: []string{"type", "name"},
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		for _, resource := range resources.Blocks {
			if resource.Labels[0] != resourceType {
				continue
			}

			body, _, diags := resource.Body.PartialContent(&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name: attributeName,
					},
				},
			})
			if diags.HasErrors() {
				return diags
			}

			if attribute, ok := body.Attributes[attributeName]; ok {
				err := walker(attribute)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// WalkResources searches for resources with a specific type and passes to the walker function
func (r *Runner) WalkResources(resourceType string, walker func(*tflint.Resource) error) error {
	for _, file := range r.Files {
		resources, _, diags := file.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "resource",
					LabelNames: []string{"type", "name"},
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		for _, resource := range resources.Blocks {
			if resource.Labels[0] != resourceType {
				continue
			}
			err := walker(&tflint.Resource{
				Name:      resource.Labels[1],
				Type:      resource.Labels[0],
				DeclRange: resource.DefRange,
				TypeRange: resource.LabelRanges[0],
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// EvaluateExpr returns a value of the passed expression.
// Note that there is no evaluation, no type conversion, etc.
func (r *Runner) EvaluateExpr(expr hcl.Expression, ret interface{}) error {
	val, diags := expr.Value(&hcl.EvalContext{})
	if diags.HasErrors() {
		return diags
	}
	return gocty.FromCtyValue(val, ret)
}

// EmitIssue adds an issue into the self
func (r *Runner) EmitIssue(rule tflint.Rule, message string, location hcl.Range, meta tflint.Metadata) error {
	r.Issues = append(r.Issues, &Issue{
		Rule:    rule,
		Message: message,
		Range:   location,
	})
	return nil
}

// EnsureNoError is a method that simply run a function if there is no error
func (r *Runner) EnsureNoError(err error, proc func() error) error {
	if err == nil {
		return proc()
	}
	return err
}
