package tflint

import (
	"fmt"
	"log"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

var _ RuleSet = &BuiltinRuleSet{}

// BuiltinRuleSet is the basis of the ruleset. Plugins can serve this ruleset directly.
// You can serve a custom ruleset by embedding this ruleset if you need special extensions.
type BuiltinRuleSet struct {
	Name    string
	Version string
	Rules   []Rule

	EnabledRules []Rule
}

// RuleSetName is the name of the ruleset.
// Generally, this is synonymous with the name of the plugin.
func (r *BuiltinRuleSet) RuleSetName() string {
	return r.Name
}

// RuleSetVersion is the version of the plugin.
func (r *BuiltinRuleSet) RuleSetVersion() string {
	return r.Version
}

// RuleNames is a list of rule names provided by the plugin.
func (r *BuiltinRuleSet) RuleNames() []string {
	names := make([]string, len(r.Rules))
	for idx, rule := range r.Rules {
		names[idx] = rule.Name()
	}
	return names
}

// ApplyGlobalConfig applies the common config to the ruleset.
// This is not supposed to be overridden from custom rulesets.
// Override the ApplyConfig if you want to apply the plugin's own configuration.
func (r *BuiltinRuleSet) ApplyGlobalConfig(config *Config) error {
	r.EnabledRules = []Rule{}

	if config.DisabledByDefault {
		log.Printf("[DEBUG] Only mode is enabled. Ignoring default plugin rules")
	}

	for _, rule := range r.Rules {
		enabled := rule.Enabled()
		if cfg := config.Rules[rule.Name()]; cfg != nil {
			enabled = cfg.Enabled
		} else if config.DisabledByDefault {
			enabled = false
		}

		if enabled {
			r.EnabledRules = append(r.EnabledRules, rule)
		}
	}
	return nil
}

// ConfigSchema returns the ruleset plugin config schema.
// This schema should be a schema inside of "plugin" block.
// Custom rulesets can override this method to return the plugin's own config schema.
func (r *BuiltinRuleSet) ConfigSchema() *hclext.BodySchema {
	return nil
}

// ApplyConfig applies the configuration to the ruleset.
// Custom rulesets can override this method to reflect the plugin's own configuration.
func (r *BuiltinRuleSet) ApplyConfig(content *hclext.BodyContent) error {
	return nil
}

// ApplyCommonConfig reflects common configurations regardless of plugins.
func (r *BuiltinRuleSet) ApplyCommonConfig(config *Config) {
	r.EnabledRules = []Rule{}

	if config.DisabledByDefault {
		log.Printf("[DEBUG] Only mode is enabled. Ignoring default plugin rules")
	}

	for _, rule := range r.Rules {
		enabled := rule.Enabled()
		if cfg := config.Rules[rule.Name()]; cfg != nil {
			enabled = cfg.Enabled
		} else if config.DisabledByDefault {
			enabled = false
		}

		if enabled {
			r.EnabledRules = append(r.EnabledRules, rule)
		}
	}
}

// Check runs inspection for each rule by applying Runner.
func (r *BuiltinRuleSet) Check(runner Runner) error {
	for _, rule := range r.EnabledRules {
		if err := rule.Check(runner); err != nil {
			return fmt.Errorf("Failed to check `%s` rule: %s", rule.Name(), err)
		}
	}
	return nil
}

func (r *BuiltinRuleSet) mustEmbedBuiltinRuleSet() {}
