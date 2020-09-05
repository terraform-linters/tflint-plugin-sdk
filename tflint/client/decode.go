package client

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/experiments"
)

// Config is an intermediate representation of configs.Config.
type Config struct {
	Path            addrs.Module
	Module          *Module
	CallRange       hcl.Range
	SourceAddr      string
	SourceAddrRange hcl.Range
	Version         string
}

func decodeConfig(config *Config) (*configs.Config, hcl.Diagnostics) {
	module, diags := decodeModule(config.Module)
	if diags.HasErrors() {
		return nil, diags
	}

	var ver *version.Version
	var err error
	if config.Version != "" {
		ver, err = version.NewVersion(config.Version)
		if err != nil {
			return nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Failed to reparse version",
					Detail:   err.Error(),
				},
			}
		}
	}

	return &configs.Config{
		Path:            config.Path,
		Module:          module,
		CallRange:       config.CallRange,
		SourceAddr:      config.SourceAddr,
		SourceAddrRange: config.SourceAddrRange,
		Version:         ver,
	}, nil
}

// Module is an intermediate representation of configs.Module.
type Module struct {
	SourceDir string

	CoreVersionConstraints      []string
	CoreVersionConstraintRanges []hcl.Range

	ActiveExperiments experiments.Set

	Backend              *Backend
	ProviderConfigs      map[string]*Provider
	ProviderRequirements *RequiredProviders
	ProviderLocalNames   map[terraform.Provider]string
	ProviderMetas        map[terraform.Provider]*ProviderMeta

	Variables map[string]*Variable
	Locals    map[string]*Local
	Outputs   map[string]*Output

	ModuleCalls map[string]*ModuleCall

	ManagedResources map[string]*Resource
	DataResources    map[string]*Resource
}

func decodeModule(module *Module) (*configs.Module, hcl.Diagnostics) {
	versionConstraints := make([]terraform.VersionConstraint, len(module.CoreVersionConstraints))
	for i, v := range module.CoreVersionConstraints {
		constraint, diags := parseVersionConstraint(v, module.CoreVersionConstraintRanges[i])
		if diags.HasErrors() {
			return nil, diags
		}
		versionConstraints[i] = constraint
	}

	backend, diags := decodeBackend(module.Backend)
	if diags.HasErrors() {
		return nil, diags
	}

	providers := map[string]*configs.Provider{}
	for k, v := range module.ProviderConfigs {
		p, diags := decodeProvider(v)
		if diags.HasErrors() {
			return nil, diags
		}
		providers[k] = p
	}

	requirements, diags := decodeRequiredProviders(module.ProviderRequirements)
	if diags.HasErrors() {
		return nil, diags
	}

	metas := map[terraform.Provider]*configs.ProviderMeta{}
	for k, v := range module.ProviderMetas {
		m, diags := decodeProviderMeta(v)
		if diags.HasErrors() {
			return nil, diags
		}
		metas[k] = m
	}

	variables := map[string]*configs.Variable{}
	for k, v := range module.Variables {
		variable, diags := decodeVariable(v)
		if diags.HasErrors() {
			return nil, diags
		}
		variables[k] = variable
	}

	locals := map[string]*configs.Local{}
	for k, v := range module.Locals {
		l, diags := decodeLocal(v)
		if diags.HasErrors() {
			return nil, diags
		}
		locals[k] = l
	}

	outputs := map[string]*configs.Output{}
	for k, v := range module.Outputs {
		o, diags := decodeOutput(v)
		if diags.HasErrors() {
			return nil, diags
		}
		outputs[k] = o
	}

	calls := map[string]*terraform.ModuleCall{}
	for k, v := range module.ModuleCalls {
		c, diags := decodeModuleCall(v)
		if diags.HasErrors() {
			return nil, diags
		}
		calls[k] = c
	}

	managed := map[string]*terraform.Resource{}
	for k, v := range module.ManagedResources {
		r, diags := decodeResource(v)
		if diags.HasErrors() {
			return nil, diags
		}
		managed[k] = r
	}

	data := map[string]*terraform.Resource{}
	for k, v := range module.DataResources {
		d, diags := decodeResource(v)
		if diags.HasErrors() {
			return nil, diags
		}
		data[k] = d
	}

	return &configs.Module{
		SourceDir: module.SourceDir,

		CoreVersionConstraints: versionConstraints,

		ActiveExperiments: module.ActiveExperiments,

		Backend:              backend,
		ProviderConfigs:      providers,
		ProviderRequirements: requirements,
		ProviderLocalNames:   module.ProviderLocalNames,
		ProviderMetas:        metas,

		Variables: variables,
		Locals:    locals,
		Outputs:   outputs,

		ModuleCalls: calls,

		ManagedResources: managed,
		DataResources:    data,
	}, nil
}

// Attribute is an intermediate representation of hcl.Attribute.
type Attribute struct {
	Name      string
	Expr      []byte
	ExprRange hcl.Range
	Range     hcl.Range
	NameRange hcl.Range
}

func decodeAttribute(attribute *Attribute) (*hcl.Attribute, hcl.Diagnostics) {
	expr, diags := parseExpression(attribute.Expr, attribute.ExprRange.Filename, attribute.ExprRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &hcl.Attribute{
		Name:      attribute.Name,
		Expr:      expr,
		Range:     attribute.Range,
		NameRange: attribute.NameRange,
	}, nil
}

// Block is an intermediate representation of hcl.Block.
type Block struct {
	Type      string
	Labels    []string
	Body      []byte
	BodyRange hcl.Range

	DefRange    hcl.Range
	TypeRange   hcl.Range
	LabelRanges []hcl.Range
}

func decodeBlock(block *Block) (*hcl.Block, hcl.Diagnostics) {
	file, diags := parseConfig(block.Body, block.BodyRange.Filename, block.BodyRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &hcl.Block{
		Type:        block.Type,
		Labels:      block.Labels,
		Body:        file.Body,
		DefRange:    block.DefRange,
		TypeRange:   block.TypeRange,
		LabelRanges: block.LabelRanges,
	}, nil
}

// Resource is an intermediate representation of terraform.Resource.
type Resource struct {
	Mode         terraform.ResourceMode
	Name         string
	Type         string
	Config       []byte
	ConfigRange  hcl.Range
	Count        []byte
	CountRange   hcl.Range
	ForEach      []byte
	ForEachRange hcl.Range

	ProviderConfigRef *terraform.ProviderConfigRef
	Provider          terraform.Provider

	Managed *ManagedResource

	DeclRange hcl.Range
	TypeRange hcl.Range
}

func decodeResource(resource *Resource) (*terraform.Resource, hcl.Diagnostics) {
	file, diags := parseConfig(resource.Config, resource.ConfigRange.Filename, resource.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	var count hcl.Expression
	if resource.Count != nil {
		count, diags = parseExpression(resource.Count, resource.CountRange.Filename, resource.CountRange.Start)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	var forEach hcl.Expression
	if resource.ForEach != nil {
		forEach, diags = parseExpression(resource.ForEach, resource.ForEachRange.Filename, resource.ForEachRange.Start)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	managed, diags := decodeManagedResource(resource.Managed)
	if diags.HasErrors() {
		return nil, diags
	}

	return &terraform.Resource{
		Mode:    resource.Mode,
		Name:    resource.Name,
		Type:    resource.Type,
		Config:  file.Body,
		Count:   count,
		ForEach: forEach,

		ProviderConfigRef: resource.ProviderConfigRef,
		Provider:          resource.Provider,

		Managed: managed,

		DeclRange: resource.DeclRange,
		TypeRange: resource.TypeRange,
	}, nil
}

// ManagedResource is an intermediate representation of terraform.ManagedResource.
type ManagedResource struct {
	Connection   *Connection
	Provisioners []*Provisioner

	CreateBeforeDestroy bool
	PreventDestroy      bool
	IgnoreAllChanges    bool

	CreateBeforeDestroySet bool
	PreventDestroySet      bool
}

func decodeManagedResource(resource *ManagedResource) (*terraform.ManagedResource, hcl.Diagnostics) {
	if resource == nil {
		return nil, nil
	}

	connection, diags := decodeConnection(resource.Connection)
	if diags.HasErrors() {
		return nil, diags
	}

	provisioners := make([]*terraform.Provisioner, len(resource.Provisioners))
	for i, p := range resource.Provisioners {
		provisioner, diags := decodeProvisioner(p)
		if diags.HasErrors() {
			return nil, diags
		}
		provisioners[i] = provisioner
	}

	return &terraform.ManagedResource{
		Connection:   connection,
		Provisioners: provisioners,

		CreateBeforeDestroy: resource.CreateBeforeDestroy,
		PreventDestroy:      resource.PreventDestroy,
		IgnoreAllChanges:    resource.IgnoreAllChanges,

		CreateBeforeDestroySet: resource.CreateBeforeDestroySet,
		PreventDestroySet:      resource.PreventDestroySet,
	}, nil
}

// ModuleCall is an intermediate representation of terraform.ModuleCall.
type ModuleCall struct {
	Name string

	SourceAddr      string
	SourceAddrRange hcl.Range
	SourceSet       bool

	Version      string
	VersionRange hcl.Range

	Config      []byte
	ConfigRange hcl.Range

	Count        []byte
	CountRange   hcl.Range
	ForEach      []byte
	ForEachRange hcl.Range

	Providers []PassedProviderConfig
	DeclRange hcl.Range

	// DependsOn []hcl.Traversal
}

// PassedProviderConfig is an intermediate representation of terraform.PassedProviderConfig.
type PassedProviderConfig struct {
	InChild  *terraform.ProviderConfigRef
	InParent *terraform.ProviderConfigRef
}

func decodeModuleCall(call *ModuleCall) (*terraform.ModuleCall, hcl.Diagnostics) {
	file, diags := parseConfig(call.Config, call.ConfigRange.Filename, call.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	var count hcl.Expression
	if call.Count != nil {
		count, diags = parseExpression(call.Count, call.CountRange.Filename, call.CountRange.Start)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	var forEach hcl.Expression
	if call.ForEach != nil {
		forEach, diags = parseExpression(call.ForEach, call.ForEachRange.Filename, call.ForEachRange.Start)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	providers := []terraform.PassedProviderConfig{}
	for _, provider := range call.Providers {
		providers = append(providers, terraform.PassedProviderConfig{
			InChild:  provider.InChild,
			InParent: provider.InParent,
		})
	}

	versionConstraint, diags := parseVersionConstraint(call.Version, call.VersionRange)
	if diags.HasErrors() {
		return nil, diags
	}

	return &terraform.ModuleCall{
		Name: call.Name,

		SourceAddr:      call.SourceAddr,
		SourceAddrRange: call.SourceAddrRange,
		SourceSet:       call.SourceSet,

		Config: file.Body,

		Version: versionConstraint,

		Count:   count,
		ForEach: forEach,

		Providers: providers,
		DeclRange: call.DeclRange,
	}, nil
}

// Backend is an intermediate representation of terraform.Backend.
type Backend struct {
	Type        string
	Config      []byte
	ConfigRange hcl.Range
	TypeRange   hcl.Range
	DeclRange   hcl.Range
}

func decodeBackend(backend *Backend) (*terraform.Backend, hcl.Diagnostics) {
	if backend == nil {
		return nil, nil
	}

	file, diags := parseConfig(backend.Config, backend.ConfigRange.Filename, backend.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &terraform.Backend{
		Type:      backend.Type,
		Config:    file.Body,
		TypeRange: backend.TypeRange,
		DeclRange: backend.DeclRange,
	}, nil
}

// Connection is an intermediate representation of terraform.Connection.
type Connection struct {
	Config      []byte
	ConfigRange hcl.Range

	DeclRange hcl.Range
}

func decodeConnection(connection *Connection) (*terraform.Connection, hcl.Diagnostics) {
	if connection == nil {
		return nil, hcl.Diagnostics{}
	}

	file, diags := parseConfig(connection.Config, connection.ConfigRange.Filename, connection.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &terraform.Connection{
		Config:    file.Body,
		DeclRange: connection.DeclRange,
	}, nil
}

// Provisioner is an intermediate representation of terraform.Provisioner.
type Provisioner struct {
	Type        string
	Config      []byte
	ConfigRange hcl.Range
	Connection  *Connection
	When        terraform.ProvisionerWhen
	OnFailure   terraform.ProvisionerOnFailure

	DeclRange hcl.Range
	TypeRange hcl.Range
}

func decodeProvisioner(provisioner *Provisioner) (*terraform.Provisioner, hcl.Diagnostics) {
	file, diags := parseConfig(provisioner.Config, provisioner.ConfigRange.Filename, provisioner.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	connection, diags := decodeConnection(provisioner.Connection)
	if diags.HasErrors() {
		return nil, diags
	}

	return &terraform.Provisioner{
		Type:       provisioner.Type,
		Config:     file.Body,
		Connection: connection,
		When:       provisioner.When,
		OnFailure:  provisioner.OnFailure,

		DeclRange: provisioner.DeclRange,
		TypeRange: provisioner.TypeRange,
	}, nil
}

func parseExpression(src []byte, filename string, start hcl.Pos) (hcl.Expression, hcl.Diagnostics) {
	if strings.HasSuffix(filename, ".tf") {
		return hclsyntax.ParseExpression(src, filename, start)
	}

	if strings.HasSuffix(filename, ".tf.json") {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "JSON configuration syntax is not supported",
				Subject:  &hcl.Range{Filename: filename, Start: start, End: start},
			},
		}
	}

	panic(fmt.Sprintf("Unexpected file: %s", filename))
}

func parseConfig(src []byte, filename string, start hcl.Pos) (*hcl.File, hcl.Diagnostics) {
	if strings.HasSuffix(filename, ".tf") {
		return hclsyntax.ParseConfig(src, filename, start)
	}

	if strings.HasSuffix(filename, ".tf.json") {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "JSON configuration syntax is not supported",
				Subject:  &hcl.Range{Filename: filename, Start: start, End: start},
			},
		}
	}

	panic(fmt.Sprintf("Unexpected file: %s", filename))
}

func parseVersionConstraint(versionStr string, versionRange hcl.Range) (terraform.VersionConstraint, hcl.Diagnostics) {
	versionConstraint := terraform.VersionConstraint{DeclRange: versionRange}
	if !versionRange.Empty() {
		required, err := version.NewConstraint(versionStr)
		if err != nil {
			detail := fmt.Sprintf(
				"Version constraint '%s' parse error: %s",
				versionStr,
				err,
			)

			return versionConstraint, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Failed to reparse version constraint",
					Detail:   detail,
					Subject:  &versionRange,
				},
			}
		}

		versionConstraint.Required = required
	}
	return versionConstraint, nil
}
