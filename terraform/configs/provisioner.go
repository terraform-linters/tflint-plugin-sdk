package configs

import "github.com/hashicorp/hcl/v2"

// Provisioner is an alternative representation of configs.Provisioner.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/provisioner.go#L11-L20
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
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/provisioner.go#L176-L180
type Connection struct {
	Config hcl.Body

	DeclRange hcl.Range
}

// ProvisionerWhen is an alternative representation of configs.ProvisionerWhen.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/provisioner.go#L182-L191
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
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/provisioner.go#L193-L203
type ProvisionerOnFailure int

const (
	// ProvisionerOnFailureInvalid is the zero value of ProvisionerOnFailure.
	ProvisionerOnFailureInvalid ProvisionerOnFailure = iota
	// ProvisionerOnFailureContinue indicates continuation on failure.
	ProvisionerOnFailureContinue
	// ProvisionerOnFailureFail indicates failure on failure.
	ProvisionerOnFailureFail
)
