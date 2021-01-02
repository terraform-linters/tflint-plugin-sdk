package configs

import (
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
)

// VersionConstraint is an alternative representation of configs.VersionConstraint.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/version_constraint.go#L16-L19
type VersionConstraint struct {
	Required  version.Constraints
	DeclRange hcl.Range
}
