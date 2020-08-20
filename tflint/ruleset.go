package tflint

import (
	"fmt"
	"log"
)

// RuleSet is a list of rules that a plugin should provide.
type RuleSet struct {
	Name    string
	Version string
	Rules   []Rule
}

// RuleSetName is the name of the ruleset.
// Generally, this is synonymous with the name of the plugin.
func (r *RuleSet) RuleSetName() string {
	return r.Name
}

// RuleSetVersion is the version of the plugin.
func (r *RuleSet) RuleSetVersion() string {
	return r.Version
}

// RuleNames is a list of rule names provided by the plugin.
func (r *RuleSet) RuleNames() []string {
	names := []string{}
	for _, rule := range r.Rules {
		names = append(names, rule.Name())
	}
	return names
}

// ApplyConfig reflects the plugin configuration in the ruleset.
// Currently used only to enable/disable rules.
func (r *RuleSet) ApplyConfig(config *Config) {
	rules := []Rule{}

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
			rules = append(rules, rule)
		}
	}
	r.Rules = rules
}

// Check runs inspection for each rule by applying Runner.
func (r *RuleSet) Check(runner Runner) error {
	for _, rule := range r.Rules {
		if err := rule.Check(runner); err != nil {
			return fmt.Errorf("Failed to check `%s` rule: %s", rule.Name(), err)
		}
	}
	return nil
}
