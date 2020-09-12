package configs

import "github.com/hashicorp/hcl/v2"

// ModuleCall is an alternative representation of configs.ModuleCall.
// https://github.com/hashicorp/terraform/blob/v0.13.2/configs/module_call.go#L12-L31
// DependsOn is not supported due to the difficulty of intermediate representation.
type ModuleCall struct {
	Name string

	SourceAddr      string
	SourceAddrRange hcl.Range
	SourceSet       bool

	Config hcl.Body

	Version VersionConstraint

	Count   hcl.Expression
	ForEach hcl.Expression

	Providers []PassedProviderConfig

	// DependsOn []hcl.Traversal

	DeclRange hcl.Range
}

// PassedProviderConfig is an alternative representation of configs.PassedProviderConfig.
// https://github.com/hashicorp/terraform/blob/v0.13.2/configs/module_call.go#L140-L143
type PassedProviderConfig struct {
	InChild  *ProviderConfigRef
	InParent *ProviderConfigRef
}
