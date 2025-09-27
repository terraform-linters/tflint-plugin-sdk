package tflint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testRule struct {
	DefaultRule
	name string
}

func (r *testRule) Name() string       { return r.name }
func (r *testRule) Enabled() bool      { return true }
func (r *testRule) Severity() Severity { return ERROR }
func (r *testRule) Check(Runner) error { return nil }

type testRule1 struct {
	testRule
}

type testRule2 struct {
	testRule
}

type testRule3 struct {
	testRule
}

func TestApplyGlobalConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   []string
	}{
		{
			name:   "default",
			config: &Config{},
			want:   []string{"test_rule1", "test_rule2", "test_rule3"},
		},
		{
			name:   "disabled by default",
			config: &Config{DisabledByDefault: true},
			want:   []string{},
		},
		{
			name: "rule config",
			config: &Config{
				Rules: map[string]*RuleConfig{
					"test_rule1": {
						Name:    "test_rule1",
						Enabled: false,
					},
				},
			},
			want: []string{"test_rule2", "test_rule3"},
		},
		{
			name:   "only",
			config: &Config{Only: []string{"test_rule1"}},
			want:   []string{"test_rule1"},
		},
		{
			name: "disabled by default + rule config",
			config: &Config{
				Rules: map[string]*RuleConfig{
					"test_rule2": {
						Name:    "test_rule2",
						Enabled: true,
					},
				},
				DisabledByDefault: true,
			},
			want: []string{"test_rule2"},
		},
		{
			name: "only + rule config",
			config: &Config{
				Rules: map[string]*RuleConfig{
					"test_rule1": {
						Name:    "test_rule1",
						Enabled: false,
					},
				},
				Only: []string{"test_rule1", "test_rule2"},
			},
			want: []string{"test_rule1", "test_rule2"},
		},
		{
			name: "disabled by default + only",
			config: &Config{
				DisabledByDefault: true,
				Only:              []string{"test_rule1", "test_rule2"},
			},
			want: []string{"test_rule1", "test_rule2"},
		},
		{
			name: "disabled by default + only + rule config",
			config: &Config{
				Rules: map[string]*RuleConfig{
					"test_rule2": {
						Name:    "test_rule2",
						Enabled: true,
					},
					"test_rule3": {
						Name:    "test_rule3",
						Enabled: false,
					},
				},
				DisabledByDefault: true,
				Only:              []string{"test_rule1", "test_rule3"},
			},
			want: []string{"test_rule1", "test_rule3"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ruleset := &BuiltinRuleSet{
				Rules: []Rule{
					&testRule1{testRule: testRule{name: "test_rule1"}},
					&testRule2{testRule: testRule{name: "test_rule2"}},
					&testRule3{testRule: testRule{name: "test_rule3"}},
				},
			}

			if err := ruleset.ApplyGlobalConfig(test.config); err != nil {
				t.Fatal(err)
			}

			got := make([]string, len(ruleset.EnabledRules))
			for i, r := range ruleset.EnabledRules {
				got[i] = r.Name()
			}

			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Error(diff)
			}
		})
	}
}
