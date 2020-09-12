package configs

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
)

// Provider is an alternative representation of configs.Provider.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/provider.go#L17-L28
type Provider struct {
	Name       string
	NameRange  hcl.Range
	Alias      string
	AliasRange *hcl.Range // nil if no alias set

	Version VersionConstraint

	Config hcl.Body

	DeclRange hcl.Range
}

// ProviderMeta is an alternative representation of configs.ProviderMeta.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/provider_meta.go#L7-L13
type ProviderMeta struct {
	Provider string
	Config   hcl.Body

	ProviderRange hcl.Range
	DeclRange     hcl.Range
}

// RequiredProvider is an alternative representation of configs.RequiredProvider.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/provider_requirements.go#L14-L20
type RequiredProvider struct {
	Name        string
	Source      string
	Type        addrs.Provider
	Requirement VersionConstraint
	DeclRange   hcl.Range
}

// RequiredProviders is an alternative representation of configs.RequiredProviders.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/provider_requirements.go#L22-L25
type RequiredProviders struct {
	RequiredProviders map[string]*RequiredProvider
	DeclRange         hcl.Range
}
