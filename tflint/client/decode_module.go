package client

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/experiments"
)

// Module is an intermediate representation of configs.Module.
type Module struct {
	SourceDir string

	CoreVersionConstraints      []string
	CoreVersionConstraintRanges []hcl.Range

	ActiveExperiments experiments.Set

	Backend              *Backend
	ProviderConfigs      map[string]*Provider
	ProviderRequirements *RequiredProviders
	ProviderLocalNames   map[addrs.Provider]string
	ProviderMetas        map[addrs.Provider]*ProviderMeta

	Variables map[string]*Variable
	Locals    map[string]*Local
	Outputs   map[string]*Output

	ModuleCalls map[string]*ModuleCall

	ManagedResources map[string]*Resource
	DataResources    map[string]*Resource
}

func decodeModule(module *Module) (*configs.Module, hcl.Diagnostics) {
	versionConstraints := make([]configs.VersionConstraint, len(module.CoreVersionConstraints))
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

	metas := map[addrs.Provider]*configs.ProviderMeta{}
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

	calls := map[string]*configs.ModuleCall{}
	for k, v := range module.ModuleCalls {
		c, diags := decodeModuleCall(v)
		if diags.HasErrors() {
			return nil, diags
		}
		calls[k] = c
	}

	managed := map[string]*configs.Resource{}
	for k, v := range module.ManagedResources {
		r, diags := decodeResource(v)
		if diags.HasErrors() {
			return nil, diags
		}
		managed[k] = r
	}

	data := map[string]*configs.Resource{}
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
