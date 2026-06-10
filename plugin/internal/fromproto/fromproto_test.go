package fromproto

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/internal/proto"
)

func TestExpression(t *testing.T) {
	tests := []struct {
		Name      string
		Expr      *proto.Expression
		Want      hcl.Range
		WantDiags hcl.Diagnostics
	}{
		{
			Name: "valid expression",
			Expr: &proto.Expression{
				Bytes: []byte("var.foo"),
				Range: &proto.Range{
					Filename: "main.tf",
					Start:    &proto.Range_Pos{Line: 1, Column: 1, Byte: 0},
					End:      &proto.Range_Pos{Line: 1, Column: 8, Byte: 7},
				},
			},
			Want: hcl.Range{
				Filename: "main.tf",
				Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
			},
		},
		{
			Name: "nil expression",
			Expr: nil,
			WantDiags: hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Failed to decode expression",
					Detail:   "expression should not be null",
				},
			},
		},
		{
			Name: "expression with nil range",
			Expr: &proto.Expression{Bytes: []byte("var.foo")},
			WantDiags: hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Failed to decode expression",
					Detail:   "expression.range should not be null",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			expr, diags := Expression(test.Expr)

			if diff := cmp.Diff(test.WantDiags, diags, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("diagnostics mismatch: %s", diff)
			}

			if test.WantDiags.HasErrors() {
				if expr != nil {
					t.Errorf("expected nil expression, got %#v", expr)
				}
				return
			}

			if expr == nil {
				t.Fatal("expected non-nil expression")
			}
			if diff := cmp.Diff(test.Want, expr.Range()); diff != "" {
				t.Errorf("range mismatch: %s", diff)
			}

			vars := expr.Variables()
			if len(vars) != 1 || vars[0].RootName() != "var" {
				t.Errorf("expected a single `var` traversal, got %#v", vars)
			}
		})
	}
}
