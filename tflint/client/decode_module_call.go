package client

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
)

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
	InChild  *configs.ProviderConfigRef
	InParent *configs.ProviderConfigRef
}

func decodeModuleCall(call *ModuleCall) (*configs.ModuleCall, hcl.Diagnostics) {
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

	providers := []configs.PassedProviderConfig{}
	for _, provider := range call.Providers {
		providers = append(providers, configs.PassedProviderConfig{
			InChild:  provider.InChild,
			InParent: provider.InParent,
		})
	}

	versionConstraint, diags := parseVersionConstraint(call.Version, call.VersionRange)
	if diags.HasErrors() {
		return nil, diags
	}

	return &configs.ModuleCall{
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
