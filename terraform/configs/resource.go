package configs

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
)

// Resource is an alternative representation of configs.Resource.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/resource.go#L14-L34
// DependsOn is not supported due to the difficulty of intermediate representation.
type Resource struct {
	Mode    addrs.ResourceMode
	Name    string
	Type    string
	Config  hcl.Body
	Count   hcl.Expression
	ForEach hcl.Expression

	ProviderConfigRef *ProviderConfigRef
	Provider          addrs.Provider

	// DependsOn []hcl.Traversal

	Managed *ManagedResource

	DeclRange hcl.Range
	TypeRange hcl.Range
}

// ManagedResource is an alternative representation of configs.ManagedResource.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/resource.go#L37-L48
// IgnoreChanges is not supported due to the difficulty of intermediate representation.
type ManagedResource struct {
	Connection   *Connection
	Provisioners []*Provisioner

	CreateBeforeDestroy bool
	PreventDestroy      bool
	// IgnoreChanges       []hcl.Traversal
	IgnoreAllChanges bool

	CreateBeforeDestroySet bool
	PreventDestroySet      bool
}

// ProviderConfigRef is an alternative representation of configs.ProviderConfigRef.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/resource.go#L384-L389
type ProviderConfigRef struct {
	Name       string
	NameRange  hcl.Range
	Alias      string
	AliasRange *hcl.Range
}
