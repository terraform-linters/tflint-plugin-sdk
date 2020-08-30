package terraform

import (
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
)

// Resource is an alternative representation of configs.Resource.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/resource.go#L14-L34
// DependsOn is not supported due to the difficulty of intermediate representation.
type Resource struct {
	Mode    ResourceMode
	Name    string
	Type    string
	Config  hcl.Body
	Count   hcl.Expression
	ForEach hcl.Expression

	ProviderConfigRef *ProviderConfigRef
	Provider          Provider

	// DependsOn []hcl.Traversal

	Managed *ManagedResource

	DeclRange hcl.Range
	TypeRange hcl.Range
}

// ManagedResource is an alternative representation of configs.ManagedResource.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/resource.go#L37-L48
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

// PassedProviderConfig is an alternative representation of configs.PassedProviderConfig.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/module_call.go#L140-L143
type PassedProviderConfig struct {
	InChild  *ProviderConfigRef
	InParent *ProviderConfigRef
}

// ProviderConfigRef is an alternative representation of configs.ProviderConfigRef.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/resource.go#L373-L378
type ProviderConfigRef struct {
	Name       string
	NameRange  hcl.Range
	Alias      string
	AliasRange *hcl.Range
}

// Provisioner is an alternative representation of configs.Provisioner.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/provisioner.go#L11-L20
type Provisioner struct {
	Type       string
	Config     hcl.Body
	Connection *Connection
	When       ProvisionerWhen
	OnFailure  ProvisionerOnFailure

	DeclRange hcl.Range
	TypeRange hcl.Range
}

// Connection is an alternative representation of configs.Connection.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/provisioner.go#L166-L170
type Connection struct {
	Config hcl.Body

	DeclRange hcl.Range
}

// Backend is an alternative representation of configs.Backend.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/backend.go#L12-L18
type Backend struct {
	Type   string
	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range
}

// ModuleCall is an alternative representation of configs.ModuleCall.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/module_call.go#L12-L31
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

// VersionConstraint is an alternative representation of configs.VersionConstraint.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/version_constraint.go#L16-L19
type VersionConstraint struct {
	Required  version.Constraints
	DeclRange hcl.Range
}

// ProvisionerWhen is an alternative representation of configs.ProvisionerWhen.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/provisioner.go#L172-L181
type ProvisionerWhen int

const (
	// ProvisionerWhenInvalid is the zero value of ProvisionerWhen.
	ProvisionerWhenInvalid ProvisionerWhen = iota
	// ProvisionerWhenCreate indicates the time of creation.
	ProvisionerWhenCreate
	// ProvisionerWhenDestroy indicates the time of deletion.
	ProvisionerWhenDestroy
)

// ProvisionerOnFailure is an alternative representation of configs.ProvisionerOnFailure.
// https://github.com/hashicorp/terraform/blob/v0.13.1/configs/provisioner.go#L183-L193
type ProvisionerOnFailure int

const (
	// ProvisionerOnFailureInvalid is the zero value of ProvisionerOnFailure.
	ProvisionerOnFailureInvalid ProvisionerOnFailure = iota
	// ProvisionerOnFailureContinue indicates continuation on failure.
	ProvisionerOnFailureContinue
	// ProvisionerOnFailureFail indicates failure on failure.
	ProvisionerOnFailureFail
)
