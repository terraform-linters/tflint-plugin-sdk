package client

import (
	"github.com/hashicorp/go-version"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
)

// Config is an intermediate representation of configs.Config.
type Config struct {
	Path            addrs.Module
	Module          *Module
	CallRange       hcl.Range
	SourceAddr      string
	SourceAddrRange hcl.Range
	Version         string
}

func decodeConfig(config *Config) (*configs.Config, hcl.Diagnostics) {
	module, diags := decodeModule(config.Module)
	if diags.HasErrors() {
		return nil, diags
	}

	var ver *version.Version
	var err error
	if config.Version != "" {
		ver, err = version.NewVersion(config.Version)
		if err != nil {
			return nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Failed to reparse version",
					Detail:   err.Error(),
				},
			}
		}
	}

	return &configs.Config{
		Path:            config.Path,
		Module:          module,
		CallRange:       config.CallRange,
		SourceAddr:      config.SourceAddr,
		SourceAddrRange: config.SourceAddrRange,
		Version:         ver,
	}, nil
}
