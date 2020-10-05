package configs

import "github.com/hashicorp/hcl/v2"

// Backend is an alternative representation of configs.Backend.
// https://github.com/hashicorp/terraform/blob/v0.13.2/configs/backend.go#L12-L18
type Backend struct {
	Type   string
	Config hcl.Body

	TypeRange hcl.Range
	DeclRange hcl.Range
}
