package hclext

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2"
)

func TestParseExpression(t *testing.T) {
	tests := []struct {
		Name      string
		Source    string
		Filename  string
		Want      string
		DiagCount int
	}{
		{
			Name:      "HCL (*.tf)",
			Source:    `"foo"`,
			Filename:  "test.tf",
			Want:      `cty.StringVal("foo")`,
			DiagCount: 0,
		},
		{
			Name:      "HCL (*.hcl)",
			Source:    `"bar"`,
			Filename:  "test.hcl",
			Want:      `cty.StringVal("bar")`,
			DiagCount: 0,
		},
		{
			Name:      "JSON (*.json)",
			Source:    `"baz"`,
			Filename:  "test.json",
			Want:      `cty.StringVal("baz")`,
			DiagCount: 0,
		},
		{
			Name:      "JSON (.tflint.json)",
			Source:    `{"config": {"force": true}}`,
			Filename:  ".tflint.json",
			Want:      `cty.ObjectVal(map[string]cty.Value{"config":cty.ObjectVal(map[string]cty.Value{"force":cty.True})})`,
			DiagCount: 0,
		},
		{
			Name: "HCL heredoc with trailing newline",
			Source: `<<EOF
foo
EOF
`,
			Filename:  "test.tf",
			Want:      `cty.StringVal("foo\n")`,
			DiagCount: 0,
		},
		{
			Name: "HCL heredoc without trailing newline",
			Source: `<<EOF
foo
EOF`,
			Filename:  "test.tf",
			Want:      `cty.StringVal("foo\n")`,
			DiagCount: 0,
		},
		{
			Name:      "json",
			Source:    `{"foo":"bar","baz":1}`,
			Filename:  "test.tf.json",
			Want:      `cty.ObjectVal(map[string]cty.Value{"baz":cty.NumberIntVal(1), "foo":cty.StringVal("bar")})`,
			DiagCount: 0,
		},
		{
			Name:      "Invalid JSON content",
			Source:    `{invalid json content}`,
			Filename:  "test.json",
			DiagCount: 2, // JSON parser returns 2 diagnostics for this invalid JSON
		},
		{
			Name:      "Invalid file extension",
			Source:    `"test"`,
			Filename:  "test.yaml",
			DiagCount: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			expr, diags := ParseExpression([]byte(test.Source), test.Filename, hcl.InitialPos)
			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}
			if diags.HasErrors() {
				return
			}

			value, diags := expr.Value(nil)
			if diags.HasErrors() {
				t.Errorf("got %d diagnostics on decode value; want 0", len(diags))
				for _, d := range diags {
					t.Logf("  - %s", d.Error())
				}
			}

			got := fmt.Sprintf("%#v", value)
			if got != test.Want {
				t.Errorf("got %s, but want %s", got, test.Want)
			}
		})
	}
}
