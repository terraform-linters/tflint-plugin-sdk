package hclext

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestContent_PartialContent(t *testing.T) {
	tests := []struct {
		Name      string
		Body      *hclsyntax.Body
		Schema    *BodySchema
		Partial   bool
		Want      *BodyContent
		DiagCount int
	}{
		{
			Name:      "nil body with nil schema",
			Body:      nil,
			Schema:    nil,
			Partial:   false,
			Want:      &BodyContent{},
			DiagCount: 0,
		},
		{
			Name:      "nil body with empty schema",
			Body:      nil,
			Schema:    &BodySchema{},
			Partial:   false,
			Want:      &BodyContent{},
			DiagCount: 0,
		},
		{
			Name:      "empty body with nil schema",
			Body:      &hclsyntax.Body{},
			Schema:    nil,
			Partial:   false,
			Want:      &BodyContent{Attributes: Attributes{}, Blocks: Blocks{}},
			DiagCount: 0,
		},
		{
			Name:      "empty body with empty schema",
			Body:      &hclsyntax.Body{},
			Schema:    &BodySchema{},
			Partial:   false,
			Want:      &BodyContent{Attributes: Attributes{}, Blocks: Blocks{}},
			DiagCount: 0,
		},
		{
			Name: "attributes",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"foo": &hclsyntax.Attribute{
						Name: "foo",
					},
				},
			},
			Schema: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "foo",
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{"foo": &Attribute{Name: "foo"}},
				Blocks:     Blocks{},
			},
			DiagCount: 0,
		},
		{
			Name: "attributes with empty schema",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"foo": &hclsyntax.Attribute{
						Name: "foo",
					},
				},
			},
			Schema:  &BodySchema{},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks:     Blocks{},
			},
			DiagCount: 1, // extra attribute is not allowed
		},
		{
			Name: "attributes with partial empty schema",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"foo": &hclsyntax.Attribute{
						Name: "foo",
					},
				},
			},
			Schema:  &BodySchema{},
			Partial: true,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks:     Blocks{},
			},
			DiagCount: 0, // extra attribute is allowed in partial schema
		},
		{
			Name: "empty body with attribute schema",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{},
			},
			Schema: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "foo",
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks:     Blocks{},
			},
			DiagCount: 0, // attribute is not required by default
		},
		{
			Name: "empty body with required attribute schema",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{},
			},
			Schema: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name:     "foo",
						Required: true,
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks:     Blocks{},
			},
			DiagCount: 1, // attribute is required
		},
		{
			Name: "attributes with block schema",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"foo": &hclsyntax.Attribute{
						Name: "foo",
					},
				},
			},
			Schema: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type: "foo",
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks:     Blocks{},
			},
			DiagCount: 1, // "foo" is defined as attribute, but should be defined as block
		},
		{
			Name: "blocks",
			Body: &hclsyntax.Body{
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "foo",
					},
				},
			},
			Schema: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type: "foo",
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks: Blocks{
					{
						Type: "foo",
						Body: &BodyContent{},
					},
				},
			},
			DiagCount: 0,
		},
		{
			Name: "multiple blocks",
			Body: &hclsyntax.Body{
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "foo",
					},
					&hclsyntax.Block{
						Type: "foo",
					},
				},
			},
			Schema: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type: "foo",
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks: Blocks{
					{
						Type: "foo",
						Body: &BodyContent{},
					},
					{
						Type: "foo",
						Body: &BodyContent{},
					},
				},
			},
			DiagCount: 0,
		},
		{
			Name: "multiple blocks which including unexpected schema",
			Body: &hclsyntax.Body{
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "foo",
					},
					&hclsyntax.Block{
						Type: "bar",
					},
				},
			},
			Schema: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type: "foo",
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks: Blocks{
					{
						Type: "foo",
						Body: &BodyContent{},
					},
				},
			},
			DiagCount: 1, // "bar" is not expected
		},
		{
			Name: "multiple blocks which including unexpected schema with partial schema",
			Body: &hclsyntax.Body{
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "foo",
					},
					&hclsyntax.Block{
						Type: "bar",
					},
				},
			},
			Schema: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type: "foo",
					},
				},
			},
			Partial: true,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks: Blocks{
					{
						Type: "foo",
						Body: &BodyContent{},
					},
				},
			},
			DiagCount: 0, // extra schema block is allowed in partial schema
		},
		{
			Name: "labeled block",
			Body: &hclsyntax.Body{
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type:   "foo",
						Labels: []string{"bar"},
					},
				},
			},
			Schema: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type:       "foo",
						LabelNames: []string{"name"},
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks: Blocks{
					{
						Type:   "foo",
						Labels: []string{"bar"},
						Body:   &BodyContent{},
					},
				},
			},
			DiagCount: 0,
		},
		{
			Name: "non-labeled block with labeled block schema",
			Body: &hclsyntax.Body{
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "foo",
					},
				},
			},
			Schema: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type:       "foo",
						LabelNames: []string{"name"},
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks:     Blocks{},
			},
			DiagCount: 1, // missing label is not allowed
		},
		{
			Name: "labeled block with non-labeled block schema",
			Body: &hclsyntax.Body{
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type:        "foo",
						Labels:      []string{"bar"},
						LabelRanges: []hcl.Range{{}},
					},
				},
			},
			Schema: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type: "foo",
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks:     Blocks{},
			},
			DiagCount: 1, // extraneous label is not allowed
		},
		{
			Name: "multi-labeled block with single-labeled block schema",
			Body: &hclsyntax.Body{
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type:        "foo",
						Labels:      []string{"bar", "baz"},
						LabelRanges: []hcl.Range{{}, {}},
					},
				},
			},
			Schema: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type:       "foo",
						LabelNames: []string{"name"},
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks:     Blocks{},
			},
			DiagCount: 1, // extraneous label is not allowed
		},
		{
			Name: "block with attribute schema",
			Body: &hclsyntax.Body{
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "foo",
					},
				},
			},
			Schema: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "foo",
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks:     Blocks{},
			},
			DiagCount: 1, // "foo" is defined as block, but should be defined as attribute
		},
		{
			Name: "nested blocks",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"foo": &hclsyntax.Attribute{Name: "foo"},
				},
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "bar",
						Body: &hclsyntax.Body{
							Attributes: hclsyntax.Attributes{
								"baz": &hclsyntax.Attribute{Name: "baz"},
							},
						},
					},
				},
			},
			Schema: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "foo",
					},
				},
				Blocks: []BlockSchema{
					{
						Type: "bar",
						Body: &BodySchema{
							Attributes: []AttributeSchema{
								{
									Name: "baz",
								},
							},
						},
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{
					"foo": &Attribute{Name: "foo"},
				},
				Blocks: Blocks{
					{
						Type: "bar",
						Body: &BodyContent{
							Attributes: Attributes{
								"baz": &Attribute{Name: "baz"},
							},
							Blocks: Blocks{},
						},
					},
				},
			},
			DiagCount: 0,
		},
		{
			Name: "attributes with empty schema in nested blocks",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"foo": &hclsyntax.Attribute{Name: "foo"},
				},
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "bar",
						Body: &hclsyntax.Body{
							Attributes: hclsyntax.Attributes{
								"baz": &hclsyntax.Attribute{Name: "baz"},
							},
						},
					},
				},
			},
			Schema: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "foo",
					},
				},
				Blocks: []BlockSchema{
					{
						Type: "bar",
						Body: &BodySchema{},
					},
				},
			},
			Partial: false,
			Want: &BodyContent{
				Attributes: Attributes{
					"foo": &Attribute{Name: "foo"},
				},
				Blocks: Blocks{
					{
						Type: "bar",
						Body: &BodyContent{
							Attributes: Attributes{},
							Blocks:     Blocks{},
						},
					},
				},
			},
			DiagCount: 1, // extra attribute in nested blocks is not allowed
		},
		{
			Name: "attributes with partial empty schema in nested blocks",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"foo": &hclsyntax.Attribute{Name: "foo"},
				},
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "bar",
						Body: &hclsyntax.Body{
							Attributes: hclsyntax.Attributes{
								"baz": &hclsyntax.Attribute{Name: "baz"},
							},
						},
					},
				},
			},
			Schema: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "foo",
					},
				},
				Blocks: []BlockSchema{
					{
						Type: "bar",
						Body: &BodySchema{},
					},
				},
			},
			Partial: true,
			Want: &BodyContent{
				Attributes: Attributes{
					"foo": &Attribute{Name: "foo"},
				},
				Blocks: Blocks{
					{
						Type: "bar",
						Body: &BodyContent{
							Attributes: Attributes{},
							Blocks:     Blocks{},
						},
					},
				},
			},
			DiagCount: 0, // extra attribute in nested blocks is allowed in partial schema
		},
		{
			Name: "empty body with attribute schema in nested blocks",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"foo": &hclsyntax.Attribute{Name: "foo"},
				},
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "bar",
						Body: &hclsyntax.Body{},
					},
				},
			},
			Schema: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "foo",
					},
				},
				Blocks: []BlockSchema{
					{
						Type: "bar",
						Body: &BodySchema{
							Attributes: []AttributeSchema{
								{
									Name: "baz",
								},
							},
						},
					},
				},
			},
			Partial: true,
			Want: &BodyContent{
				Attributes: Attributes{
					"foo": &Attribute{Name: "foo"},
				},
				Blocks: Blocks{
					{
						Type: "bar",
						Body: &BodyContent{
							Attributes: Attributes{},
							Blocks:     Blocks{},
						},
					},
				},
			},
			DiagCount: 0, // attribute in nested blocks is not required by default
		},
		{
			Name: "empty body with required attribute schema in nested blocks",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"foo": &hclsyntax.Attribute{Name: "foo"},
				},
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "bar",
						Body: &hclsyntax.Body{},
					},
				},
			},
			Schema: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name: "foo",
					},
				},
				Blocks: []BlockSchema{
					{
						Type: "bar",
						Body: &BodySchema{
							Attributes: []AttributeSchema{
								{
									Name:     "baz",
									Required: true,
								},
							},
						},
					},
				},
			},
			Partial: true,
			Want: &BodyContent{
				Attributes: Attributes{
					"foo": &Attribute{Name: "foo"},
				},
				Blocks: Blocks{
					{
						Type: "bar",
						Body: &BodyContent{
							Attributes: Attributes{},
							Blocks:     Blocks{},
						},
					},
				},
			},
			DiagCount: 1, // attribute in nested blocks is required
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var got *BodyContent
			var diags hcl.Diagnostics
			if test.Partial {
				got, diags = PartialContent(test.Body, test.Schema)
			} else {
				got, diags = Content(test.Body, test.Schema)
			}

			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if diff := cmp.Diff(test.Want, got); diff != "" {
				t.Errorf("wrong result\ndiff: %s", diff)
			}
		})
	}
}

func TestContent_JustAttributes(t *testing.T) {
	tests := []struct {
		Name      string
		Body      *hclsyntax.Body
		Schema    *BodySchema
		Partial   bool
		Want      *BodyContent
		DiagCount int
	}{
		{
			Name: "just attributes in the top level",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"foo": &hclsyntax.Attribute{Name: "foo"},
					"bar": &hclsyntax.Attribute{Name: "bar"},
					"baz": &hclsyntax.Attribute{Name: "baz"},
				},
			},
			Schema: &BodySchema{Mode: SchemaJustAttributesMode},
			Want: &BodyContent{
				Attributes: Attributes{
					"foo": &Attribute{Name: "foo"},
					"bar": &Attribute{Name: "bar"},
					"baz": &Attribute{Name: "baz"},
				},
				Blocks: Blocks{},
			},
		},
		{
			Name: "just attributes in nested blocks",
			Body: &hclsyntax.Body{
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "bar",
						Body: &hclsyntax.Body{
							Attributes: hclsyntax.Attributes{
								"foo": &hclsyntax.Attribute{Name: "foo"},
								"bar": &hclsyntax.Attribute{Name: "bar"},
								"baz": &hclsyntax.Attribute{Name: "baz"},
							},
						},
					},
				},
			},
			Schema: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type: "bar",
						Body: &BodySchema{Mode: SchemaJustAttributesMode},
					},
				},
			},
			Want: &BodyContent{
				Attributes: Attributes{},
				Blocks: Blocks{
					{
						Type: "bar",
						Body: &BodyContent{
							Attributes: Attributes{
								"foo": &Attribute{Name: "foo"},
								"bar": &Attribute{Name: "bar"},
								"baz": &Attribute{Name: "baz"},
							},
							Blocks: Blocks{},
						},
					},
				},
			},
		},
		{
			Name: "just attributes in body with blocks",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"foo": &hclsyntax.Attribute{Name: "foo"},
					"bar": &hclsyntax.Attribute{Name: "bar"},
					"baz": &hclsyntax.Attribute{Name: "baz"},
				},
				Blocks: hclsyntax.Blocks{
					&hclsyntax.Block{
						Type: "bar",
						Body: &hclsyntax.Body{},
					},
				},
			},
			Schema: &BodySchema{Mode: SchemaJustAttributesMode},
			Want: &BodyContent{
				Attributes: Attributes{
					"foo": &Attribute{Name: "foo"},
					"bar": &Attribute{Name: "bar"},
					"baz": &Attribute{Name: "baz"},
				},
				Blocks: Blocks{},
			},
			DiagCount: 1, // Unexpected "bar" block; Blocks are not allowed here.
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var got *BodyContent
			var diags hcl.Diagnostics
			if test.Partial {
				got, diags = PartialContent(test.Body, test.Schema)
			} else {
				got, diags = Content(test.Body, test.Schema)
			}

			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if diff := cmp.Diff(test.Want, got); diff != "" {
				t.Errorf("wrong result\ndiff: %s", diff)
			}
		})
	}
}

func Test_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		body *BodyContent
		want bool
	}{
		{
			name: "body is not empty",
			body: &BodyContent{
				Attributes: Attributes{
					"foo": &Attribute{Name: "foo"},
				},
			},
			want: false,
		},
		{
			name: "body has empty attributes and empty blocks",
			body: &BodyContent{Attributes: Attributes{}, Blocks: Blocks{}},
			want: true,
		},
		{
			name: "body has nil attributes and nil blocks",
			body: &BodyContent{},
			want: true,
		},
		{
			name: "body is nil",
			body: nil,
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.body.IsEmpty() != test.want {
				t.Errorf("%t is expected, but got %t", test.want, test.body.IsEmpty())
			}
		})
	}
}

func TestCopy_BodyContent(t *testing.T) {
	body := &BodyContent{
		Attributes: Attributes{
			"foo": {Name: "foo"},
		},
		Blocks: Blocks{
			{
				Body: &BodyContent{
					Attributes: Attributes{
						"bar": {Name: "bar"},
					},
					Blocks: Blocks{
						{
							Body: &BodyContent{
								Attributes: Attributes{
									"baz": {Name: "baz"},
								},
							},
						},
					},
				},
			},
			{
				Body: &BodyContent{
					Attributes: Attributes{
						"aaa": {Name: "aaa"},
						"bbb": {Name: "bbb"},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(body.Copy(), body); diff != "" {
		t.Error(diff)
	}
}

func TestWalkAttributes(t *testing.T) {
	body := &BodyContent{
		Attributes: Attributes{
			"foo": {Name: "foo"},
		},
		Blocks: Blocks{
			{
				Body: &BodyContent{
					Attributes: Attributes{
						"bar": {Name: "bar"},
					},
					Blocks: Blocks{
						{
							Body: &BodyContent{
								Attributes: Attributes{
									"baz": {Name: "baz"},
								},
							},
						},
					},
				},
			},
			{
				Body: &BodyContent{
					Attributes: Attributes{
						"aaa": {Name: "aaa"},
						"bbb": {Name: "bbb"},
					},
				},
			},
		},
	}

	got := []string{}
	diags := body.WalkAttributes(func(a *Attribute) hcl.Diagnostics {
		got = append(got, a.Name)
		return nil
	})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	want := []string{"foo", "bar", "baz", "aaa", "bbb"}

	opt := cmpopts.SortSlices(func(x, y string) bool { return x < y })
	if diff := cmp.Diff(got, want, opt); diff != "" {
		t.Error(diff)
	}
}

func TestCopy_Attribute(t *testing.T) {
	attribute := &Attribute{
		Name:      "foo",
		Expr:      hcl.StaticExpr(cty.StringVal("foo"), hcl.Range{}),
		Range:     hcl.Range{Start: hcl.Pos{Line: 2}},
		NameRange: hcl.Range{Start: hcl.Pos{Line: 1}},
	}

	must := func(v cty.Value, diags hcl.Diagnostics) cty.Value {
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		return v
	}
	opts := cmp.Options{
		cmp.Comparer(func(x, y hcl.Expression) bool {
			return must(x.Value(nil)) == must(y.Value(nil))
		}),
	}
	if diff := cmp.Diff(attribute.Copy(), attribute, opts); diff != "" {
		t.Error(diff)
	}
}

func TestCopy_Block(t *testing.T) {
	block := &Block{
		Type:   "foo",
		Labels: []string{"bar", "baz"},
		Body: &BodyContent{
			Attributes: Attributes{
				"foo": {Name: "foo"},
			},
			Blocks: Blocks{},
		},
		DefRange:  hcl.Range{Start: hcl.Pos{Line: 1}},
		TypeRange: hcl.Range{Start: hcl.Pos{Line: 2}},
		LabelRanges: []hcl.Range{
			{Start: hcl.Pos{Line: 3}},
			{Start: hcl.Pos{Line: 4}},
		},
	}

	if diff := cmp.Diff(block.Copy(), block); diff != "" {
		t.Error(diff)
	}
}
