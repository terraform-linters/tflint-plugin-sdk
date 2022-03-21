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
			Name:      "HCL but file extension is invalid (*.json)",
			Source:    `"baz"`,
			Filename:  "test.json",
			DiagCount: 1,
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
