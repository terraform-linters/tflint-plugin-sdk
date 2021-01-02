package configs

import (
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/experiments"
)

// Module is an alternative representation of configs.Module.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/module.go#L14-L45
type Module struct {
	SourceDir string

	CoreVersionConstraints []VersionConstraint

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
