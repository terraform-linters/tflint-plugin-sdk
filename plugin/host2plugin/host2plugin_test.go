package host2plugin

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/plugin2host"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

func startTestGRPCPluginServer(t *testing.T, ruleset tflint.RuleSet) *GRPCClient {
	client, _ := plugin.TestPluginGRPCConn(t, map[string]plugin.Plugin{
		"ruleset": &RuleSetPlugin{impl: ruleset},
	})
	raw, err := client.Dispense("ruleset")
	if err != nil {
		t.Fatalf("failed to dispense: %s", err)
	}
	return raw.(*GRPCClient)
}

var _ tflint.RuleSet = &mockRuleSet{}

type mockRuleSet struct {
	tflint.BuiltinRuleSet

	impl mockRuleSetImpl
}

type mockRuleSetImpl struct {
	ruleNames         func() []string
	configSchema      func() *hclext.BodySchema
	applyGlobalConfig func(*tflint.Config) error
	applyConfig       func(*hclext.BodyContent) error
	check             func(tflint.Runner) error
}

func (r *mockRuleSet) RuleNames() []string {
	if r.impl.ruleNames != nil {
		return r.impl.ruleNames()
	}
	return []string{}
}

func (r *mockRuleSet) ConfigSchema() *hclext.BodySchema {
	if r.impl.configSchema != nil {
		return r.impl.configSchema()
	}
	return &hclext.BodySchema{}
}

func (r *mockRuleSet) ApplyGlobalConfig(config *tflint.Config) error {
	if r.impl.applyGlobalConfig != nil {
		return r.impl.applyGlobalConfig(config)
	}
	return nil
}

func (r *mockRuleSet) ApplyConfig(content *hclext.BodyContent) error {
	if r.impl.applyConfig != nil {
		return r.impl.applyConfig(content)
	}
	return nil
}

func (r *mockRuleSet) Check(runner tflint.Runner) error {
	if r.impl.check != nil {
		return r.impl.check(runner)
	}
	return nil
}

func newMockRuleSet(name, version string, impl mockRuleSetImpl) *mockRuleSet {
	return &mockRuleSet{
		BuiltinRuleSet: tflint.BuiltinRuleSet{
			Name:    name,
			Version: version,
		},
		impl: impl,
	}
}

func TestRuleSetName(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		Name        string
		RuleSetName string
		Want        string
		ErrCheck    func(error) bool
	}{
		{
			Name:        "rule set name",
			RuleSetName: "test_ruleset",
			Want:        "test_ruleset",
			ErrCheck:    neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCPluginServer(t, newMockRuleSet(test.RuleSetName, "0.1.0", mockRuleSetImpl{}))

			got, err := client.RuleSetName()
			if test.ErrCheck(err) {
				t.Fatalf("failed to call RuleSetName: %s", err)
			}

			if got != test.Want {
				t.Errorf("expected `%s`, but got `%s`", test.Want, got)
			}
		})
	}
}

func TestRuleSetVersion(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		Name           string
		RuleSetVersion string
		Want           string
		ErrCheck       func(error) bool
	}{
		{
			Name:           "rule set version",
			RuleSetVersion: "0.1.0",
			Want:           "0.1.0",
			ErrCheck:       neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCPluginServer(t, newMockRuleSet("test_ruleset", test.RuleSetVersion, mockRuleSetImpl{}))

			got, err := client.RuleSetVersion()
			if test.ErrCheck(err) {
				t.Fatalf("failed to call RuleSetVersion: %s", err)
			}

			if got != test.Want {
				t.Errorf("expected `%s`, but got `%s`", test.Want, got)
			}
		})
	}
}

func TestRuleNames(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		Name       string
		ServerImpl func() []string
		Want       []string
		ErrCheck   func(error) bool
	}{
		{
			Name: "rule names",
			ServerImpl: func() []string {
				return []string{"test1", "test2"}
			},
			Want:     []string{"test1", "test2"},
			ErrCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCPluginServer(t, newMockRuleSet("test_ruleset", "0.1.0", mockRuleSetImpl{ruleNames: test.ServerImpl}))

			got, err := client.RuleNames()
			if test.ErrCheck(err) {
				t.Fatalf("failed to call RuleNames: %s", err)
			}

			if diff := cmp.Diff(got, test.Want); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestConfigSchema(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	// nested schema example
	schema := &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{
			{Name: "foo", Required: true},
		},
		Blocks: []hclext.BlockSchema{
			{
				Type:       "bar",
				LabelNames: []string{"baz", "qux"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "qux", Required: true},
					},
					Blocks: []hclext.BlockSchema{
						{
							Type:       "baz",
							LabelNames: []string{"foo", "bar"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{},
								Blocks:     []hclext.BlockSchema{},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		Name       string
		ServerImpl func() *hclext.BodySchema
		Want       *hclext.BodySchema
		ErrCheck   func(error) bool
	}{
		{
			Name: "nested schema",
			ServerImpl: func() *hclext.BodySchema {
				return schema
			},
			Want:     schema,
			ErrCheck: neverHappend,
		},
		{
			Name: "nil schema",
			ServerImpl: func() *hclext.BodySchema {
				return nil
			},
			Want: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{},
				Blocks:     []hclext.BlockSchema{},
			},
			ErrCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCPluginServer(t, newMockRuleSet("test_ruleset", "0.1.0", mockRuleSetImpl{configSchema: test.ServerImpl}))

			got, err := client.ConfigSchema()
			if test.ErrCheck(err) {
				t.Fatalf("failed to call ConfigSchema: %s", err)
			}

			if diff := cmp.Diff(got, test.Want); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestApplyGlobalConfig(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		Name       string
		Arg        *tflint.Config
		ServerImpl func(*tflint.Config) error
		ErrCheck   func(error) bool
	}{
		{
			Name: "nil config",
			Arg:  nil,
			ServerImpl: func(config *tflint.Config) error {
				if len(config.Rules) != 0 {
					return fmt.Errorf("config rules should be empty, but %#v", config.Rules)
				}
				if config.DisabledByDefault != false {
					return errors.New("disabled by default should be false")
				}
				return nil
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "full config",
			Arg: &tflint.Config{
				Rules: map[string]*tflint.RuleConfig{
					"test1": {Name: "test1", Enabled: true},
					"test2": {Name: "test2", Enabled: false},
				},
				DisabledByDefault: true,
			},
			ServerImpl: func(config *tflint.Config) error {
				want := &tflint.Config{
					Rules: map[string]*tflint.RuleConfig{
						"test1": {Name: "test1", Enabled: true},
						"test2": {Name: "test2", Enabled: false},
					},
					DisabledByDefault: true,
				}

				if diff := cmp.Diff(config, want); diff != "" {
					return fmt.Errorf("diff: %s", diff)
				}
				return nil
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "server returns an error",
			Arg:  nil,
			ServerImpl: func(config *tflint.Config) error {
				return errors.New("unexpected error")
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCPluginServer(t, newMockRuleSet("test_ruleset", "0.1.0", mockRuleSetImpl{applyGlobalConfig: test.ServerImpl}))

			err := client.ApplyGlobalConfig(test.Arg)
			if test.ErrCheck(err) {
				t.Fatalf("failed to call ApplyGlobalConfig: %s", err)
			}
		})
	}
}

func TestApplyConfig(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	// test util functions
	hclFile := func(filename string, code string) *hcl.File {
		file, diags := hclsyntax.ParseConfig([]byte(code), filename, hcl.InitialPos)
		if diags.HasErrors() {
			panic(diags)
		}
		return file
	}
	schema := &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{{Name: "name"}},
		Blocks: []hclext.BlockSchema{
			{
				Type:       "block",
				LabelNames: []string{"bar"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "nested"}},
				},
			},
		},
	}

	tests := []struct {
		Name       string
		Args       func() (*hclext.BodyContent, map[string][]byte)
		ServerImpl func(*hclext.BodyContent) error
		ErrCheck   func(error) bool
	}{
		{
			Name: "nil content",
			Args: func() (*hclext.BodyContent, map[string][]byte) {
				return nil, nil
			},
			ServerImpl: func(content *hclext.BodyContent) error {
				want := &hclext.BodyContent{
					Attributes: hclext.Attributes{},
					Blocks:     hclext.Blocks{},
				}

				if diff := cmp.Diff(content, want); diff != "" {
					return fmt.Errorf("diff: %s", diff)
				}
				return nil
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "nested content",
			Args: func() (*hclext.BodyContent, map[string][]byte) {
				file := hclFile(".tflint.hcl", `
name = "foo"
block "bar" {
	nested = "baz"
}`)
				content, diags := hclext.Content(file.Body, schema)
				if diags.HasErrors() {
					panic(diags)
				}

				return content, map[string][]byte{".tflint.hcl": file.Bytes}
			},
			ServerImpl: func(content *hclext.BodyContent) error {
				file := hclFile(".tflint.hcl", `
name = "foo"
block "bar" {
	nested = "baz"
}`)
				want, diags := hclext.Content(file.Body, schema)
				if diags.HasErrors() {
					return diags
				}

				opts := cmp.Options{
					cmp.Comparer(func(x, y cty.Value) bool {
						return x.GoString() == y.GoString()
					}),
				}
				if diff := cmp.Diff(content, want, opts); diff != "" {
					return fmt.Errorf("diff: %s", diff)
				}
				return nil
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "server returns an error",
			Args: func() (*hclext.BodyContent, map[string][]byte) {
				return nil, nil
			},
			ServerImpl: func(content *hclext.BodyContent) error {
				return errors.New("unexpected error")
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCPluginServer(t, newMockRuleSet("test_ruleset", "0.1.0", mockRuleSetImpl{applyConfig: test.ServerImpl}))

			err := client.ApplyConfig(test.Args())
			if test.ErrCheck(err) {
				t.Fatalf("failed to call ApplyConfig: %s", err)
			}
		})
	}
}

var _ plugin2host.Server = &mockServer{}

type mockServer struct {
	impl mockServerImpl
}

type mockServerImpl struct {
	getFile func(string) (*hcl.File, error)
}

func (s *mockServer) GetModuleContent(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
	return &hclext.BodyContent{}, hcl.Diagnostics{}
}

func (s *mockServer) GetFile(filename string) (*hcl.File, error) {
	if s.impl.getFile != nil {
		return s.impl.getFile(filename)
	}
	return nil, nil
}

func (s *mockServer) GetRuleConfigContent(name string, schema *hclext.BodySchema) (*hclext.BodyContent, map[string][]byte, error) {
	return &hclext.BodyContent{}, map[string][]byte{}, nil
}

func (s *mockServer) EvaluateExpr(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
	return cty.Value{}, nil
}

func (s *mockServer) EmitIssue(rule tflint.Rule, message string, location hcl.Range) error {
	return nil
}

func (s *mockServer) GetFiles(tflint.ModuleCtxType) map[string][]byte {
	return map[string][]byte{}
}

func TestCheck(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	// test util functions
	hclFile := func(filename string, code string) (*hcl.File, error) {
		file, diags := hclsyntax.ParseConfig([]byte(code), filename, hcl.InitialPos)
		if diags.HasErrors() {
			return file, diags
		}
		return file, nil
	}

	tests := []struct {
		Name       string
		Arg        func() plugin2host.Server
		ServerImpl func(tflint.Runner) error
		ErrCheck   func(error) bool
	}{
		{
			Name: "bidirectional",
			Arg: func() plugin2host.Server {
				return &mockServer{
					impl: mockServerImpl{
						getFile: func(filename string) (*hcl.File, error) {
							return hclFile("test.tf", `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`)
						},
					},
				}
			},
			ServerImpl: func(runner tflint.Runner) error {
				got, err := runner.GetFile("test.tf")
				if err != nil {
					return err
				}

				want, err := hclFile("test.tf", `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`)
				if err != nil {
					return err
				}

				opts := cmp.Options{
					cmp.Comparer(func(x, y cty.Value) bool {
						return x.GoString() == y.GoString()
					}),
					cmp.AllowUnexported(hclsyntax.Body{}),
					cmpopts.IgnoreFields(hcl.File{}, "Nav"),
				}
				if diff := cmp.Diff(got, want, opts); diff != "" {
					return fmt.Errorf("diff: %s", diff)
				}
				return nil
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "host server returns an error",
			Arg: func() plugin2host.Server {
				return &mockServer{
					impl: mockServerImpl{
						getFile: func(filename string) (*hcl.File, error) {
							return nil, errors.New("unexpected error")
						},
					},
				}
			},
			ServerImpl: func(runner tflint.Runner) error {
				_, err := runner.GetFile("test.tf")
				return err
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
		},
		{
			Name: "plugin server returns an error",
			Arg: func() plugin2host.Server {
				return &mockServer{}
			},
			ServerImpl: func(runner tflint.Runner) error {
				return errors.New("unexpected error")
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCPluginServer(t, newMockRuleSet("test_ruleset", "0.1.0", mockRuleSetImpl{check: test.ServerImpl}))

			err := client.Check(test.Arg())
			if test.ErrCheck(err) {
				t.Fatalf("failed to call Check: %s", err)
			}
		})
	}
}
