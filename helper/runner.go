package helper

import (
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Runner is a mock that satisfies the Runner interface for plugin testing.
type Runner struct {
	Files  map[string]*hcl.File
	Issues Issues
}

// WalkResourceAttributes visits all specified attributes from Files.
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

// WalkResourceBlocks visits all specified blocks from Files.
func (r *Runner) WalkResourceBlocks(resourceType, blockType string, walker func(*hcl.Block) error) error {
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
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: blockType,
					},
				},
			})
			if diags.HasErrors() {
				return diags
			}

			for _, block := range body.Blocks {
				err := walker(block)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// WalkResources visits all specified resources from Files.
func (r *Runner) WalkResources(resourceType string, walker func(*terraform.Resource) error) error {
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

		for _, block := range resources.Blocks {
			resource, diags := simpleDecodeResouceBlock(block)
			if diags.HasErrors() {
				return diags
			}

			if resource.Type != resourceType {
				continue
			}

			err := walker(resource)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// WalkModuleCalls visits all module calls from Files.
func (r *Runner) WalkModuleCalls(walker func(*terraform.ModuleCall) error) error {
	for _, file := range r.Files {
		calls, _, diags := file.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "module",
					LabelNames: []string{"name"},
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		for _, block := range calls.Blocks {
			call, diags := simpleDecodeModuleCallBlock(block)
			if diags.HasErrors() {
				return diags
			}

			err := walker(call)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Backend returns the terraform backend configuration.
func (r *Runner) Backend() (*terraform.Backend, error) {
	for _, file := range r.Files {
		tfcfg, _, diags := file.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{Type: "terraform"},
			},
		})
		if diags.HasErrors() {
			return nil, diags
		}

		for _, block := range tfcfg.Blocks {
			backendCfg, _, diags := block.Body.PartialContent(&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{Type: "backend", LabelNames: []string{"type"}},
				},
			})
			if diags.HasErrors() {
				return nil, diags
			}

			for _, backendBlock := range backendCfg.Blocks {
				return &terraform.Backend{
					Type:      backendBlock.Labels[0],
					TypeRange: backendBlock.LabelRanges[0],
					Config:    backendBlock.Body,
					DeclRange: backendBlock.DefRange,
				}, nil
			}
		}
	}

	return nil, nil
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

// EmitIssueOnExpr adds an issue to the runner itself.
func (r *Runner) EmitIssueOnExpr(rule tflint.Rule, message string, expr hcl.Expression) error {
	r.Issues = append(r.Issues, &Issue{
		Rule:    rule,
		Message: message,
		Range:   expr.Range(),
	})
	return nil
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

// simpleDecodeResourceBlock decodes the data equivalent to configs.Resource from hcl.Block
// without depending on Terraform. Some operations have been omitted for ease of implementation.
// As such, it is expected to parse the minimal code needed for testing.
// https://github.com/hashicorp/terraform/blob/v0.12.26/configs/resource.go#L78-L288
func simpleDecodeResouceBlock(resource *hcl.Block) (*terraform.Resource, hcl.Diagnostics) {
	content, resourceRemain, diags := resource.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "count",
			},
			{
				Name: "for_each",
			},
			{
				Name: "provider",
			},
			{
				Name: "depends_on",
			},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "lifecycle"},
			{Type: "connection"},
			{Type: "provisioner", LabelNames: []string{"type"}},
		},
	})
	if diags.HasErrors() {
		return nil, diags
	}

	var count hcl.Expression
	if attr, exists := content.Attributes["count"]; exists {
		count = attr.Expr
	}

	var forEach hcl.Expression
	if attr, exists := content.Attributes["for_each"]; exists {
		forEach = attr.Expr
	}

	var ref *terraform.ProviderConfigRef
	if attr, exists := content.Attributes["provider"]; exists {
		ref, diags = decodeProviderConfigRef(attr.Expr)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	managed := &terraform.ManagedResource{}
	for _, block := range content.Blocks {
		switch block.Type {
		case "lifecycle":
			content, _, diags := block.Body.PartialContent(&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name: "create_before_destroy",
					},
					{
						Name: "prevent_destroy",
					},
					{
						Name: "ignore_changes",
					},
				},
			})
			if diags.HasErrors() {
				return nil, diags
			}

			if attr, exists := content.Attributes["create_before_destroy"]; exists {
				if diags := gohcl.DecodeExpression(attr.Expr, nil, &managed.CreateBeforeDestroy); diags.HasErrors() {
					return nil, diags
				}
				managed.CreateBeforeDestroySet = true
			}
			if attr, exists := content.Attributes["prevent_destroy"]; exists {
				if diags := gohcl.DecodeExpression(attr.Expr, nil, &managed.PreventDestroy); diags.HasErrors() {
					return nil, diags
				}
				managed.PreventDestroySet = true
			}
			if attr, exists := content.Attributes["ignore_changes"]; exists {
				if hcl.ExprAsKeyword(attr.Expr) == "all" {
					managed.IgnoreAllChanges = true
				}
			}
		case "connection":
			managed.Connection = &terraform.Connection{
				Config:    block.Body,
				DeclRange: block.DefRange,
			}
		case "provisioner":
			pv := &terraform.Provisioner{
				Type:      block.Labels[0],
				TypeRange: block.LabelRanges[0],
				DeclRange: block.DefRange,
				When:      terraform.ProvisionerWhenCreate,
				OnFailure: terraform.ProvisionerOnFailureFail,
			}

			content, config, diags := block.Body.PartialContent(&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{Name: "when"},
					{Name: "on_failure"},
				},
				Blocks: []hcl.BlockHeaderSchema{
					{Type: "connection"},
				},
			})
			if diags.HasErrors() {
				return nil, diags
			}
			pv.Config = config

			if attr, exists := content.Attributes["when"]; exists {
				switch hcl.ExprAsKeyword(attr.Expr) {
				case "create":
					pv.When = terraform.ProvisionerWhenCreate
				case "destroy":
					pv.When = terraform.ProvisionerWhenDestroy
				}
			}

			if attr, exists := content.Attributes["on_failure"]; exists {
				switch hcl.ExprAsKeyword(attr.Expr) {
				case "continue":
					pv.OnFailure = terraform.ProvisionerOnFailureContinue
				case "fail":
					pv.OnFailure = terraform.ProvisionerOnFailureFail
				}
			}

			for _, block := range content.Blocks {
				pv.Connection = &terraform.Connection{
					Config:    block.Body,
					DeclRange: block.DefRange,
				}
			}

			managed.Provisioners = append(managed.Provisioners, pv)
		}
	}

	return &terraform.Resource{
		Mode:    terraform.ManagedResourceMode,
		Name:    resource.Labels[1],
		Type:    resource.Labels[0],
		Config:  resourceRemain,
		Count:   count,
		ForEach: forEach,

		ProviderConfigRef: ref,

		Managed: managed,

		DeclRange: resource.DefRange,
		TypeRange: resource.LabelRanges[0],
	}, nil
}

func simpleDecodeModuleCallBlock(block *hcl.Block) (*terraform.ModuleCall, hcl.Diagnostics) {
	content, remain, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "source", Required: true},
			{Name: "version"},
			{Name: "providers"},
		},
	})
	if diags.HasErrors() {
		return nil, diags
	}

	var sourceAddr string
	var sourceAddrRange hcl.Range
	if attr, exists := content.Attributes["source"]; exists {
		if diags := gohcl.DecodeExpression(attr.Expr, nil, &sourceAddr); diags.HasErrors() {
			return nil, diags
		}
		sourceAddrRange = attr.Expr.Range()
	}

	providers := []terraform.PassedProviderConfig{}
	if attr, exists := content.Attributes["providers"]; exists {
		pairs, diags := hcl.ExprMap(attr.Expr)
		if diags.HasErrors() {
			return nil, diags
		}

		for _, pair := range pairs {
			key, diags := decodeProviderConfigRef(pair.Key)
			if diags.HasErrors() {
				return nil, diags
			}

			value, diags := decodeProviderConfigRef(pair.Value)
			if diags.HasErrors() {
				return nil, diags
			}

			providers = append(providers, terraform.PassedProviderConfig{
				InChild:  key,
				InParent: value,
			})
		}
	}

	var versionRequired version.Constraints
	var versionValue string
	var versionRange hcl.Range
	var err error
	if attr, exists := content.Attributes["version"]; exists {
		versionRange = attr.Expr.Range()

		if diags := gohcl.DecodeExpression(attr.Expr, nil, &versionValue); diags.HasErrors() {
			return nil, diags
		}

		versionRequired, err = version.NewConstraint(versionValue)
		if err != nil {
			return nil, hcl.Diagnostics{
				{Severity: hcl.DiagError, Summary: "Invalid version constraint"},
			}
		}
	}

	return &terraform.ModuleCall{
		Name: block.Labels[0],

		SourceAddr:      sourceAddr,
		SourceAddrRange: sourceAddrRange,
		SourceSet:       !sourceAddrRange.Empty(),

		Config:      remain,
		ConfigRange: block.DefRange,

		Version: terraform.VersionConstraint{
			Required:  versionRequired,
			DeclRange: versionRange,
		},

		Providers: providers,

		DeclRange: block.DefRange,
	}, nil
}

func decodeProviderConfigRef(expr hcl.Expression) (*terraform.ProviderConfigRef, hcl.Diagnostics) {
	traversal, diags := hcl.AbsTraversalForExpr(expr)
	if diags.HasErrors() {
		return nil, diags
	}

	ref := &terraform.ProviderConfigRef{
		Name:      traversal.RootName(),
		NameRange: traversal[0].SourceRange(),
	}

	if len(traversal) > 1 {
		aliasStep := traversal[1].(hcl.TraverseAttr)
		ref.Alias = aliasStep.Name
		ref.AliasRange = aliasStep.SourceRange().Ptr()
	}

	return ref, nil
}
