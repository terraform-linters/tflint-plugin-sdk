package configs

import (
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
)

// Config is an alternative representation of configs.Config.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/config.go#L22-L78
type Config struct {
	// Root            *Config
	// Parent          *Config
	Path addrs.Module
	// Children        map[string]*Config
	Module          *Module
	CallRange       hcl.Range
	SourceAddr      string
	SourceAddrRange hcl.Range
	Version         *version.Version
}
