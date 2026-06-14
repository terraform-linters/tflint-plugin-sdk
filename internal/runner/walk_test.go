package runner

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	hcljson "github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func TestWalkExpressions(t *testing.T) {
	for _, tc := range []struct {
		name   string
		files  map[string]string
		walked []hcl.Range
	}{
		{
			name: "native syntax walks recursively",
			files: map[string]string{
				"resource.tf": `
resource "null_resource" "test" {
  key = "foo"
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			name: "object based json",
			files: map[string]string{
				"resource.tf.json": `
{
  "resource": {
    "null_resource": {
      "test": {
        "key": "foo"
      }
    }
  }
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 15}, End: hcl.Pos{Line: 9, Column: 4}},
			},
		},
		{
			name: "array based json",
			files: map[string]string{
				"main.tf.json": `[
  {
    "resource": {
      "null_resource": {
        "foo": {}
      }
    }
  },
  {
    "variable": {
      "example": {
        "type": "string"
      }
    }
  }
]`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 17}, End: hcl.Pos{Line: 7, Column: 6}},
				{Start: hcl.Pos{Line: 10, Column: 17}, End: hcl.Pos{Line: 14, Column: 6}},
			},
		},
		{
			name: "multiple files",
			files: map[string]string{
				"main.tf": `
locals {
  key = "foo"
}`,
				"main.tf.json": `{"locals": {"key": "bar"}}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}, Filename: "main.tf"},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}, Filename: "main.tf"},
				{Start: hcl.Pos{Line: 1, Column: 12}, End: hcl.Pos{Line: 1, Column: 26}, Filename: "main.tf.json"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]*hcl.File{}
			for name, src := range tc.files {
				var file *hcl.File
				var diags hcl.Diagnostics
				if strings.HasSuffix(name, ".json") {
					file, diags = hcljson.Parse([]byte(src), name)
				} else {
					file, diags = hclsyntax.ParseConfig([]byte(src), name, hcl.InitialPos)
				}
				if diags.HasErrors() {
					t.Fatal(diags)
				}
				files[name] = file
			}

			walked := []hcl.Range{}
			diags := WalkExpressions(files, tflint.ExprWalkFunc(func(expr hcl.Expression) hcl.Diagnostics {
				walked = append(walked, expr.Range())
				return nil
			}))
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			opts := cmp.Options{
				cmpopts.IgnoreFields(hcl.Range{}, "Filename"),
				cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
				cmpopts.SortSlices(func(x, y hcl.Range) bool { return x.String() > y.String() }),
			}
			if diff := cmp.Diff(walked, tc.walked, opts); diff != "" {
				t.Error(diff)
			}
		})
	}
}
