package terraform

// ResourceMode is an alternative representation of addrs.ResourceMode.
// https://github.com/hashicorp/terraform/blob/v0.12.26/addrs/resource.go#L253-L271
type ResourceMode rune

const (
	// InvalidResourceMode is the zero value of ResourceMode.
	InvalidResourceMode ResourceMode = 0
	// ManagedResourceMode indicates a managed resource.
	ManagedResourceMode ResourceMode = 'M'
	// DataResourceMode indicates a data resource.
	DataResourceMode ResourceMode = 'D'
)
