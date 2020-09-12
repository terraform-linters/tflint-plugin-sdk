package helper

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
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
func (r *Runner) WalkResources(resourceType string, walker func(*configs.Resource) error) error {
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
func (r *Runner) WalkModuleCalls(walker func(*configs.ModuleCall) error) error {
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
func (r *Runner) Backend() (*configs.Backend, error) {
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
				return &configs.Backend{
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

// Config returns the Terraform configuration
func (r *Runner) Config() (*configs.Config, error) {
	config := &configs.Config{
		Module: &configs.Module{},
	}

	for _, file := range r.Files {
		content, diags := file.Body.Content(configFileSchema)
		if diags.HasErrors() {
			return nil, diags
		}

		for _, block := range content.Blocks {
			switch block.Type {
			case "terraform":
				content, diags := block.Body.Content(terraformBlockSchema)
				if diags.HasErrors() {
					return nil, diags
				}

				for _, block := range content.Blocks {
					switch block.Type {
					case "backend":
						config.Module.Backend = &configs.Backend{
							Type:      block.Labels[0],
							TypeRange: block.LabelRanges[0],
							Config:    block.Body,
							DeclRange: block.DefRange,
						}
					case "required_providers":
						// TODO
					case "provider_meta":
						// TODO
					default:
						continue
					}
				}
			case "provider":
				// TODO
			case "variable":
				// TODO
			case "locals":
				// TODO
			case "output":
				// TODO
			case "module":
				call, diags := simpleDecodeModuleCallBlock(block)
				if diags.HasErrors() {
					return nil, diags
				}
				config.Module.ModuleCalls[call.Name] = call
			case "resource":
				resource, diags := simpleDecodeResouceBlock(block)
				if diags.HasErrors() {
					return nil, diags
				}
				config.Module.ManagedResources[fmt.Sprintf("%s.%s", resource.Type, resource.Name)] = resource
			case "data":
				// TODO
			default:
				continue
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
// https://github.com/hashicorp/terraform/blob/v0.13.2/configs/resource.go#L80-L290
func simpleDecodeResouceBlock(resource *hcl.Block) (*configs.Resource, hcl.Diagnostics) {
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

	var ref *configs.ProviderConfigRef
	if attr, exists := content.Attributes["provider"]; exists {
		ref, diags = decodeProviderConfigRef(attr.Expr)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	managed := &configs.ManagedResource{}
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
			managed.Connection = &configs.Connection{
				Config:    block.Body,
				DeclRange: block.DefRange,
			}
		case "provisioner":
			pv := &configs.Provisioner{
				Type:      block.Labels[0],
				TypeRange: block.LabelRanges[0],
				DeclRange: block.DefRange,
				When:      configs.ProvisionerWhenCreate,
				OnFailure: configs.ProvisionerOnFailureFail,
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
					pv.When = configs.ProvisionerWhenCreate
				case "destroy":
					pv.When = configs.ProvisionerWhenDestroy
				}
			}

			if attr, exists := content.Attributes["on_failure"]; exists {
				switch hcl.ExprAsKeyword(attr.Expr) {
				case "continue":
					pv.OnFailure = configs.ProvisionerOnFailureContinue
				case "fail":
					pv.OnFailure = configs.ProvisionerOnFailureFail
				}
			}

			for _, block := range content.Blocks {
				pv.Connection = &configs.Connection{
					Config:    block.Body,
					DeclRange: block.DefRange,
				}
			}

			managed.Provisioners = append(managed.Provisioners, pv)
		}
	}

	return &configs.Resource{
		Mode:    addrs.ManagedResourceMode,
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

func simpleDecodeModuleCallBlock(block *hcl.Block) (*configs.ModuleCall, hcl.Diagnostics) {
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

	providers := []configs.PassedProviderConfig{}
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

			providers = append(providers, configs.PassedProviderConfig{
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

	return &configs.ModuleCall{
		Name: block.Labels[0],

		SourceAddr:      sourceAddr,
		SourceAddrRange: sourceAddrRange,
		SourceSet:       !sourceAddrRange.Empty(),

		Config: remain,

		Version: configs.VersionConstraint{
			Required:  versionRequired,
			DeclRange: versionRange,
		},

		Providers: providers,

		DeclRange: block.DefRange,
	}, nil
}

func decodeProviderConfigRef(expr hcl.Expression) (*configs.ProviderConfigRef, hcl.Diagnostics) {
	traversal, diags := hcl.AbsTraversalForExpr(expr)
	if diags.HasErrors() {
		return nil, diags
	}

	ref := &configs.ProviderConfigRef{
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

// configFileSchema is the schema for the top-level of a config file.
// @see https://github.com/hashicorp/terraform/blob/v0.13.2/configs/parser_config.go#L197-L239
var configFileSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "terraform",
		},
		{
			Type: "required_providers",
		},
		{
			Type:       "provider",
			LabelNames: []string{"name"},
		},
		{
			Type:       "variable",
			LabelNames: []string{"name"},
		},
		{
			Type: "locals",
		},
		{
			Type:       "output",
			LabelNames: []string{"name"},
		},
		{
			Type:       "module",
			LabelNames: []string{"name"},
		},
		{
			Type:       "resource",
			LabelNames: []string{"type", "name"},
		},
		{
			Type:       "data",
			LabelNames: []string{"type", "name"},
		},
	},
}

// terraformBlockSchema is the schema for a top-level "terraform" block in a configuration file.
// @see https://github.com/hashicorp/terraform/blob/v0.13.2/configs/parser_config.go#L241-L261
var terraformBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "required_version"},
		{Name: "experiments"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "backend",
			LabelNames: []string{"type"},
		},
		{
			Type: "required_providers",
		},
		{
			Type:       "provider_meta",
			LabelNames: []string{"provider"},
		},
	},
}
