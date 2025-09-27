package lang

import (
	"sort"
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
)

func TestReferencesInExpr(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  []*addrs.Reference
	}{
		{
			Name:  "input variable",
			Input: `var.foo`,
			Want: []*addrs.Reference{
				{
					Subject: addrs.InputVariable{
						Name: "foo",
					},
					SourceRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
				},
			},
		},
		{
			Name:  "multiple input variables",
			Input: `"${var.foo}-${var.bar}"`,
			Want: []*addrs.Reference{
				{
					Subject: addrs.InputVariable{
						Name: "foo",
					},
					SourceRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   hcl.Pos{Line: 1, Column: 11, Byte: 10},
					},
				},
				{
					Subject: addrs.InputVariable{
						Name: "bar",
					},
					SourceRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 15, Byte: 14},
						End:   hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
				},
			},
		},
		{
			Name:  "input variable and resource",
			Input: `"${boop_instance.foo}_${var.foo}"`,
			Want: []*addrs.Reference{
				{
					Subject: addrs.Resource{
						Mode: addrs.ManagedResourceMode,
						Type: "boop_instance",
						Name: "foo",
					},
					SourceRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
					},
				},
				{
					Subject: addrs.InputVariable{
						Name: "foo",
					},
					SourceRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 25, Byte: 24},
						End:   hcl.Pos{Line: 1, Column: 32, Byte: 31},
					},
				},
			},
		},
		{
			Name:  "contains invalid references",
			Input: `"${boop_instance}_${var.foo}"`, // A reference to a resource type must be followed by at least one attribute access, specifying the resource name.
			Want: []*addrs.Reference{
				{
					Subject: addrs.InputVariable{
						Name: "foo",
					},
					SourceRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 21, Byte: 20},
						End:   hcl.Pos{Line: 1, Column: 28, Byte: 27},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			expr, diags := hclsyntax.ParseExpression([]byte(test.Input), "", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			got := ReferencesInExpr(expr)
			sort.Slice(got, func(i, j int) bool {
				return got[i].SourceRange.String() > got[j].SourceRange.String()
			})

			for _, problem := range deep.Equal(got, test.Want) {
				t.Error(problem)
			}
		})
	}
}
