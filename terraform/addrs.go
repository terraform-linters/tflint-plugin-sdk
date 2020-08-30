package terraform

import svchost "github.com/hashicorp/terraform-svchost"

// Provider is an alternative representation of addrs.Provider.
// https://github.com/hashicorp/terraform/blob/v0.13.1/addrs/provider.go#L16-L20
type Provider struct {
	Type      string
	Namespace string
	Hostname  svchost.Hostname
}

// ResourceMode is an alternative representation of addrs.ResourceMode.
// https://github.com/hashicorp/terraform/blob/v0.13.1/addrs/resource.go#L326-L344
type ResourceMode rune

const (
	// InvalidResourceMode is the zero value of ResourceMode.
	InvalidResourceMode ResourceMode = 0
	// ManagedResourceMode indicates a managed resource.
	ManagedResourceMode ResourceMode = 'M'
	// DataResourceMode indicates a data resource.
	DataResourceMode ResourceMode = 'D'
)
