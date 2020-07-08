package terraform

import "github.com/hashicorp/hcl/v2"

// Resource is an alternative representation of configs.Resource.
// https://github.com/hashicorp/terraform/blob/v0.12.26/configs/resource.go#L13-L33
// DependsOn is not supported due to the difficulty of intermediate representation.
type Resource struct {
	Mode    ResourceMode
	Name    string
	Type    string
	Config  hcl.Body
	Count   hcl.Expression
	ForEach hcl.Expression

	ProviderConfigRef *ProviderConfigRef

	// DependsOn []hcl.Traversal

	Managed *ManagedResource

	DeclRange hcl.Range
	TypeRange hcl.Range
}

// ManagedResource is an alternative representation of configs.ManagedResource.
// https://github.com/hashicorp/terraform/blob/v0.12.26/configs/resource.go#L35-L47
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
// https://github.com/hashicorp/terraform/blob/v0.12.26/configs/resource.go#L371-L376
type ProviderConfigRef struct {
	Name       string
	NameRange  hcl.Range
	Alias      string
	AliasRange *hcl.Range
}

// Provisioner is an alternative representation of configs.Provisioner.
// https://github.com/hashicorp/terraform/blob/v0.12.26/configs/provisioner.go#L9-L20
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
// https://github.com/hashicorp/terraform/blob/v0.12.26/configs/provisioner.go#L164-L170
type Connection struct {
	Config hcl.Body

	DeclRange hcl.Range
}

// Backend is an alternative representation of configs.Backend.
// https://github.com/hashicorp/terraform/blob/v0.12.26/configs/backend.go#L12-L18
type Backend struct {
	Type        string
	Config      hcl.Body
	ConfigRange hcl.Range
	TypeRange   hcl.Range
	DeclRange   hcl.Range
}

// ProvisionerWhen is an alternative representation of configs.ProvisionerWhen.
// https://github.com/hashicorp/terraform/blob/v0.12.26/configs/provisioner.go#L172-L181
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
// https://github.com/hashicorp/terraform/blob/v0.12.26/configs/provisioner.go#L183-L193
type ProvisionerOnFailure int

const (
	// ProvisionerOnFailureInvalid is the zero value of ProvisionerOnFailure.
	ProvisionerOnFailureInvalid ProvisionerOnFailure = iota
	// ProvisionerOnFailureContinue indicates continuation on failure.
	ProvisionerOnFailureContinue
	// ProvisionerOnFailureFail indicates failure on failure.
	ProvisionerOnFailureFail
)
