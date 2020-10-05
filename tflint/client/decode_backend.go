package client

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
)

// Backend is an intermediate representation of terraform.Backend.
type Backend struct {
	Type        string
	Config      []byte
	ConfigRange hcl.Range
	TypeRange   hcl.Range
	DeclRange   hcl.Range
}

func decodeBackend(backend *Backend) (*configs.Backend, hcl.Diagnostics) {
	if backend == nil {
		return nil, nil
	}

	file, diags := parseConfig(backend.Config, backend.ConfigRange.Filename, backend.ConfigRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &configs.Backend{
		Type:      backend.Type,
		Config:    file.Body,
		TypeRange: backend.TypeRange,
		DeclRange: backend.DeclRange,
	}, nil
}
