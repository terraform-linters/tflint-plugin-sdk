package client

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
)

// Provider is an intermediate representation of configs.Provider.
type Provider struct {
	Name       string
	NameRange  hcl.Range
	Alias      string
	AliasRange *hcl.Range // nil if no alias set

	Version      string
	VersionRange hcl.Range

	Config      []byte
	ConfigRange hcl.Range

	DeclRange hcl.Range
}

func decodeProvider(provider *Provider) (*configs.Provider, hcl.Diagnostics) {
	if provider == nil {
		return nil, nil
	}

	versionConstraint, diags := parseVersionConstraint(provider.Version, provider.VersionRange)
	if diags.HasErrors() {
		return nil, diags
	}

	file, diags := parseConfig(provider.Config, provider.ConfigRange.Filename, provider.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &configs.Provider{
		Name:       provider.Name,
		NameRange:  provider.NameRange,
		Alias:      provider.Alias,
		AliasRange: provider.AliasRange,

		Version: versionConstraint,

		Config: file.Body,

		DeclRange: provider.DeclRange,
	}, nil
}

// ProviderMeta is an intermediate representation of configs.ProviderMeta.
type ProviderMeta struct {
	Provider    string
	Config      []byte
	ConfigRange hcl.Range

	ProviderRange hcl.Range
	DeclRange     hcl.Range
}

func decodeProviderMeta(meta *ProviderMeta) (*configs.ProviderMeta, hcl.Diagnostics) {
	file, diags := parseConfig(meta.Config, meta.ConfigRange.Filename, meta.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &configs.ProviderMeta{
		Provider: meta.Provider,
		Config:   file.Body,

		ProviderRange: meta.ProviderRange,
		DeclRange:     meta.DeclRange,
	}, nil
}

// RequiredProvider is an intermediate representation of configs.RequiredProvider.
type RequiredProvider struct {
	Name             string
	Source           string
	Type             addrs.Provider
	Requirement      string
	RequirementRange hcl.Range
	DeclRange        hcl.Range
}

func decodeRequiredProvider(provider *RequiredProvider) (*configs.RequiredProvider, hcl.Diagnostics) {
	versionConstraint, diags := parseVersionConstraint(provider.Requirement, provider.RequirementRange)
	if diags.HasErrors() {
		return nil, diags
	}

	return &configs.RequiredProvider{
		Name:        provider.Name,
		Source:      provider.Source,
		Type:        provider.Type,
		Requirement: versionConstraint,
		DeclRange:   provider.DeclRange,
	}, nil
}

// RequiredProviders is an intermediate representation of configs.RequiredProviders.
type RequiredProviders struct {
	RequiredProviders map[string]*RequiredProvider
	DeclRange         hcl.Range
}

func decodeRequiredProviders(providers *RequiredProviders) (*configs.RequiredProviders, hcl.Diagnostics) {
	ret := map[string]*configs.RequiredProvider{}
	for k, v := range providers.RequiredProviders {
		p, diags := decodeRequiredProvider(v)
		if diags.HasErrors() {
			return nil, diags
		}
		ret[k] = p
	}

	return &configs.RequiredProviders{
		RequiredProviders: ret,
		DeclRange:         providers.DeclRange,
	}, nil
}
