package configs

import (
	"github.com/terraform-linters/tflint-plugin-sdk/terraform"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/experiments"
)

// Module is an alternative representation of configs.Module.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/module.go#L14-L45
type Module struct {
	SourceDir string

	CoreVersionConstraints []terraform.VersionConstraint

	ActiveExperiments experiments.Set

	Backend              *terraform.Backend
	ProviderConfigs      map[string]*Provider
	ProviderRequirements *RequiredProviders
	ProviderLocalNames   map[terraform.Provider]string
	ProviderMetas        map[terraform.Provider]*ProviderMeta

	Variables map[string]*Variable
	Locals    map[string]*Local
	Outputs   map[string]*Output

	ModuleCalls map[string]*terraform.ModuleCall

	ManagedResources map[string]*terraform.Resource
	DataResources    map[string]*terraform.Resource
}
