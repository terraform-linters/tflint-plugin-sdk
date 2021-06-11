package helper

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Runner is a mock that satisfies the Runner interface for plugin testing.
type Runner struct {
	files  map[string]*hcl.File
	Issues Issues

	tfconfig *configs.Config
	config   Config
}

// NewLocalRunner initialises a new test runner.
// Internal use only.
func NewLocalRunner(files map[string]*hcl.File, issues Issues) *Runner {
	return &Runner{files: map[string]*hcl.File{}, Issues: issues}
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

// WalkResourceAttributes visits all specified attributes from Files.
func (r *Runner) WalkResourceAttributes(resourceType, attributeName string, walker func(*hcl.Attribute) error) error {
	for _, resource := range r.tfconfig.Module.ManagedResources {
		if resource.Type != resourceType {
			continue
		}

		body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
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

	return nil
}

// WalkResourceBlocks visits all specified blocks from Files.
func (r *Runner) WalkResourceBlocks(resourceType, blockType string, walker func(*hcl.Block) error) error {
	for _, resource := range r.tfconfig.Module.ManagedResources {
		if resource.Type != resourceType {
			continue
		}

		body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
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

	return nil
}

// WalkResources visits all specified resources from Files.
func (r *Runner) WalkResources(resourceType string, walker func(*configs.Resource) error) error {
	for _, resource := range r.tfconfig.Module.ManagedResources {
		if resource.Type != resourceType {
			continue
		}

		err := walker(resource)
		if err != nil {
			return err
		}
	}

	return nil
}

// WalkModuleCalls visits all module calls from Files.
func (r *Runner) WalkModuleCalls(walker func(*configs.ModuleCall) error) error {
	for _, call := range r.tfconfig.Module.ModuleCalls {
		err := walker(call)
		if err != nil {
			return err
		}
	}

	return nil
}

// Backend returns the terraform backend configuration.
func (r *Runner) Backend() (*configs.Backend, error) {
	return r.tfconfig.Module.Backend, nil
}

// Config returns the Terraform configuration
func (r *Runner) Config() (*configs.Config, error) {
	return r.tfconfig, nil
}

// File returns the hcl.File object
func (r *Runner) File(filename string) (*hcl.File, error) {
	return r.files[filename], nil
}

// Files returns a map[string]hcl.File object
func (r *Runner) Files() (map[string]*hcl.File, error) {
	return r.files, nil
}

// RootProvider returns the provider configuration.
// In the helper runner, it always returns its own provider.
func (r *Runner) RootProvider(name string) (*configs.Provider, error) {
	return r.tfconfig.Module.ProviderConfigs[name], nil
}

// DecodeRuleConfig extracts the rule's configuration into the given value
func (r *Runner) DecodeRuleConfig(name string, ret interface{}) error {
	for _, rule := range r.config.Rules {
		if rule.Name == name {
			if diags := gohcl.DecodeBody(rule.Body, nil, ret); diags.HasErrors() {
				return diags
			}
			return nil
		}
	}

	return nil
}

// EvaluateExpr returns a value of the passed expression.
// Note that some features are limited
func (r *Runner) EvaluateExpr(expr hcl.Expression, ret interface{}, wantTy *cty.Type) error {
	var wantType cty.Type

	if wantTy != nil {
		wantType = *wantTy
	}
	if wantType == (cty.Type{}) {
		switch ret.(type) {
		case *string, string:
			wantType = cty.String
		case *int, int:
			wantType = cty.Number
		case *[]string, []string:
			wantType = cty.List(cty.String)
		case *[]int, []int:
			wantType = cty.List(cty.Number)
		case *map[string]string, map[string]string:
			wantType = cty.Map(cty.String)
		case *map[string]int, map[string]int:
			wantType = cty.Map(cty.Number)
		default:
			panic(fmt.Errorf("Unexpected result type: %T", ret))
		}
	}

	variables := map[string]cty.Value{}
	for _, variable := range r.tfconfig.Module.Variables {
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
	val, err := convert.Convert(rawVal, wantType)
	if err != nil {
		return err
	}

	return gocty.FromCtyValue(val, ret)
}

// EvaluateExprOnRootCtx returns a value of the passed expression.
// Note this is just alias of EvaluateExpr.
func (r *Runner) EvaluateExprOnRootCtx(expr hcl.Expression, ret interface{}, wantType *cty.Type) error {
	return r.EvaluateExpr(expr, ret, wantType)
}

// IsNullExpr checks whether the passed expression is null or not.
// Note that it does not eval the expression for simplify the implementation.
func (r *Runner) IsNullExpr(expr hcl.Expression) (bool, error) {
	val, diags := expr.Value(nil)
	if diags.HasErrors() {
		return false, diags
	}
	return val.IsNull(), nil
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

func (r *Runner) initFromFiles() error {
	r.tfconfig = &configs.Config{
		Module: &configs.Module{
			ModuleCalls:      map[string]*configs.ModuleCall{},
			ManagedResources: map[string]*configs.Resource{},
			Variables:        map[string]*configs.Variable{},
		},
	}

	for _, file := range r.files {
		content, diags := file.Body.Content(configFileSchema)
		if diags.HasErrors() {
			return diags
		}

		for _, block := range content.Blocks {
			switch block.Type {
			case "terraform":
				content, diags := block.Body.Content(terraformBlockSchema)
				if diags.HasErrors() {
					return diags
				}

				for _, block := range content.Blocks {
					switch block.Type {
					case "backend":
						r.tfconfig.Module.Backend = &configs.Backend{
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
				variable, diags := simpleDecodeVariableBlock(block)
				if diags.HasErrors() {
					return diags
				}
				r.tfconfig.Module.Variables[variable.Name] = variable
			case "locals":
				// TODO
			case "output":
				// TODO
			case "module":
				call, diags := simpleDecodeModuleCallBlock(block)
				if diags.HasErrors() {
					return diags
				}
				r.tfconfig.Module.ModuleCalls[call.Name] = call
			case "resource":
				resource, diags := simpleDecodeResouceBlock(block)
				if diags.HasErrors() {
					return diags
				}
				r.tfconfig.Module.ManagedResources[fmt.Sprintf("%s.%s", resource.Type, resource.Name)] = resource
			case "data":
				// TODO
			default:
				continue
			}
		}
	}

	return nil
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

func simpleDecodeVariableBlock(block *hcl.Block) (*configs.Variable, hcl.Diagnostics) {
	v := &configs.Variable{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	content, diags := block.Body.Content(&hcl.BodySchema{
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
