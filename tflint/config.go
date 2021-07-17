package tflint

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Config is a TFLint configuration applied to the plugin.
// The Body contains the contents declared in the "plugin" block.
type Config struct {
	Rules             map[string]*RuleConfig
	DisabledByDefault bool
	Body              hcl.Body
}

// RuleConfig is a TFLint's rule configuration.
type RuleConfig struct {
	Name    string
	Enabled bool
}

// MarshalledConfig is an intermediate representation of Config for communicating over RPC
type MarshalledConfig struct {
	Rules             map[string]*RuleConfig
	DisabledByDefault bool
	BodyBytes         []byte
	BodyRange         hcl.Range
}

// Unmarshal converts intermediate representations into the Config object.
func (c *MarshalledConfig) Unmarshal() (*Config, error) {
	// HACK: Always add a newline to avoid heredoc parse errors.
	// @see https://github.com/hashicorp/hcl/issues/441
	src := []byte(string(c.BodyBytes) + "\n")
	file, diags := hclsyntax.ParseConfig(src, c.BodyRange.Filename, c.BodyRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &Config{
		Rules:             c.Rules,
		DisabledByDefault: c.DisabledByDefault,
		Body:              file.Body,
	}, nil
}
