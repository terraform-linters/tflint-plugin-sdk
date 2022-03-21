package hclext

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestDecodeBody(t *testing.T) {
	makeInstantiateType := func(target interface{}) func() interface{} {
		return func() interface{} {
			return reflect.New(reflect.TypeOf(target)).Interface()
		}
	}
	parseExpr := func(src string) hcl.Expression {
		expr, diags := hclsyntax.ParseExpression([]byte(src), "", hcl.InitialPos)
		if diags.HasErrors() {
			panic(diags)
		}
		return expr
	}
	equals := func(other interface{}) func(v interface{}) bool {
		return func(v interface{}) bool {
			return cmp.Equal(v, other)
		}
	}
	noop := func(v interface{}) bool { return true }

	type withTwoAttributes struct {
		A string `hclext:"a,optional"`
		B string `hclext:"b,optional"`
	}

	type withNestedBlock struct {
		Plain  string             `hclext:"plain,optional"`
		Nested *withTwoAttributes `hclext:"nested,block"`
	}

	type withListofNestedBlocks struct {
		Nested []*withTwoAttributes `hclext:"nested,block"`
	}

	type withListofNestedBlocksNoPointers struct {
		Nested []withTwoAttributes `hclext:"nested,block"`
	}

	tests := []struct {
		Name      string
		Body      *BodyContent
		Target    func() interface{}
		Check     func(v interface{}) bool
		DiagCount int
	}{
		{
			Name:      "nil body",
			Body:      nil,
			Target:    makeInstantiateType(struct{}{}),
			Check:     equals(struct{}{}),
			DiagCount: 0,
		},
		{
			Name:      "empty body",
			Body:      &BodyContent{},
			Target:    makeInstantiateType(struct{}{}),
			Check:     equals(struct{}{}),
			DiagCount: 0,
		},
		{
			Name: "empty body with optional attr schema (pointer)",
			Body: &BodyContent{},
			Target: makeInstantiateType(struct {
				Name *string `hclext:"name"`
			}{}),
			Check: equals(struct {
				Name *string `hclext:"name"`
			}{}),
			DiagCount: 0,
		},
		{
			Name: "empty body with optional attr schema (label)",
			Body: &BodyContent{},
			Target: makeInstantiateType(struct {
				Name string `hclext:"name,optional"`
			}{}),
			Check: equals(struct {
				Name string `hclext:"name,optional"`
			}{}),
			DiagCount: 0,
		},
		{
			Name: "empty body with required attr schema",
			Body: &BodyContent{},
			Target: makeInstantiateType(struct {
				Name string `hclext:"name"`
			}{}),
			Check: equals(struct {
				Name string `hclext:"name"`
			}{}),
			DiagCount: 1, // attr is required by default
		},
		{
			Name: "required attr",
			Body: &BodyContent{
				Attributes: Attributes{
					"name": &Attribute{Name: "name", Expr: parseExpr(`"Ermintrude"`)},
				},
			},
			Target: makeInstantiateType(struct {
				Name string `hclext:"name"`
			}{}),
			Check: equals(struct {
				Name string `hclext:"name"`
			}{"Ermintrude"}),
			DiagCount: 0,
		},
		{
			Name: "extraneous attr",
			Body: &BodyContent{
				Attributes: Attributes{
					"name": &Attribute{Name: "name", Expr: parseExpr(`"Ermintrude"`)},
					"age":  &Attribute{Name: "age", Expr: parseExpr(`23`)},
				},
			},
			Target: makeInstantiateType(struct {
				Name string `hclext:"name"`
			}{}),
			Check: equals(struct {
				Name string `hclext:"name"`
			}{"Ermintrude"}),
			DiagCount: 0, // extraneous attr is ignored
		},
		{
			Name: "single block with required single block schema",
			Body: &BodyContent{
				Blocks: Blocks{&Block{Type: "noodle"}},
			},
			Target: makeInstantiateType(struct {
				Noodle struct{} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle struct{} `hclext:"noodle,block"`
			}{}),
			DiagCount: 0,
		},
		{
			Name: "single block with optional single block schema",
			Body: &BodyContent{
				Blocks: Blocks{&Block{Type: "noodle"}},
			},
			Target: makeInstantiateType(struct {
				Noodle *struct{} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle *struct{} `hclext:"noodle,block"`
			}{Noodle: &struct{}{}}),
			DiagCount: 0,
		},
		{
			Name: "single block with multiple block schema",
			Body: &BodyContent{
				Blocks: Blocks{&Block{Type: "noodle"}},
			},
			Target: makeInstantiateType(struct {
				Noodle []struct{} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle []struct{} `hclext:"noodle,block"`
			}{Noodle: []struct{}{{}}}),
			DiagCount: 0,
		},
		{
			Name: "multiple blocks with required single block schema",
			Body: &BodyContent{
				Blocks: Blocks{&Block{Type: "noodle"}, &Block{Type: "noodle"}},
			},
			Target: makeInstantiateType(struct {
				Noodle struct{} `hclext:"noodle,block"`
			}{}),
			Check:     noop,
			DiagCount: 1, // duplicate block is not allowed
		},
		{
			Name: "multiple blocks with optional single block schema",
			Body: &BodyContent{
				Blocks: Blocks{&Block{Type: "noodle"}, &Block{Type: "noodle"}},
			},
			Target: makeInstantiateType(struct {
				Noodle *struct{} `hclext:"noodle,block"`
			}{}),
			Check:     noop,
			DiagCount: 1, // duplicate block is not allowed
		},
		{
			Name: "multiple blocks with multiple block schema",
			Body: &BodyContent{
				Blocks: Blocks{&Block{Type: "noodle"}, &Block{Type: "noodle"}},
			},
			Target: makeInstantiateType(struct {
				Noodle []struct{} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle []struct{} `hclext:"noodle,block"`
			}{Noodle: []struct{}{{}, {}}}),
			DiagCount: 0,
		},
		{
			Name: "empty body with required single block schema",
			Body: &BodyContent{},
			Target: makeInstantiateType(struct {
				Noodle struct{} `hclext:"noodle,block"`
			}{}),
			Check:     noop,
			DiagCount: 1, // block is required by default
		},
		{
			Name: "empty body with optional single block schema",
			Body: &BodyContent{},
			Target: makeInstantiateType(struct {
				Noodle *struct{} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle *struct{} `hclext:"noodle,block"`
			}{nil}),
			DiagCount: 0,
		},
		{
			Name: "empty body with multiple block schema",
			Body: &BodyContent{},
			Target: makeInstantiateType(struct {
				Noodle []struct{} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle []struct{} `hclext:"noodle,block"`
			}{}),
			DiagCount: 0,
		},
		{
			Name: "non-labeled block with labeled block schema",
			Body: &BodyContent{
				Blocks: Blocks{&Block{Type: "noodle"}},
			},
			Target: makeInstantiateType(struct {
				Noodle struct {
					Name string `hclext:"name,label"`
				} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle struct {
					Name string `hclext:"name,label"`
				} `hclext:"noodle,block"`
			}{}),
			DiagCount: 1, // label is required by default
		},
		{
			Name: "labeled block with labeled block schema",
			Body: &BodyContent{
				Blocks: Blocks{&Block{Type: "noodle", Labels: []string{"foo"}}},
			},
			Target: makeInstantiateType(struct {
				Noodle struct {
					Name string `hclext:"name,label"`
				} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle struct {
					Name string `hclext:"name,label"`
				} `hclext:"noodle,block"`
			}{Noodle: struct {
				Name string `hclext:"name,label"`
			}{Name: "foo"}}),
			DiagCount: 0,
		},
		{
			Name: "multi-labeled blocks with labeled block schema",
			Body: &BodyContent{
				Blocks: Blocks{&Block{Type: "noodle", Labels: []string{"foo", "bar"}}},
			},
			Target: makeInstantiateType(struct {
				Noodle struct {
					Name string `hclext:"name,label"`
				} `hclext:"noodle,block"`
			}{}),
			Check:     noop,
			DiagCount: 1, // extraneous label is not allowed
		},
		{
			Name: "labeled blocks with multi-labeled block schema",
			Body: &BodyContent{
				Blocks: Blocks{&Block{Type: "noodle", Labels: []string{"foo"}}},
			},
			Target: makeInstantiateType(struct {
				Noodle struct {
					Name string `hclext:"name,label"`
					Type string `hclext:"type,label"`
				} `hclext:"noodle,block"`
			}{}),
			Check:     noop,
			DiagCount: 1, // missing label is not allowed
		},
		{
			Name: "multi-labeled blocks with multi-labeled block schema",
			Body: &BodyContent{
				Blocks: Blocks{&Block{Type: "noodle", Labels: []string{"foo", "bar"}}},
			},
			Target: makeInstantiateType(struct {
				Noodle struct {
					Name string `hclext:"name,label"`
					Type string `hclext:"type,label"`
				} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle struct {
					Name string `hclext:"name,label"`
					Type string `hclext:"type,label"`
				} `hclext:"noodle,block"`
			}{Noodle: struct {
				Name string `hclext:"name,label"`
				Type string `hclext:"type,label"`
			}{Name: "foo", Type: "bar"}}),
			DiagCount: 0,
		},
		{
			Name: "multiple non-labeled blocks with labeled block schema",
			Body: &BodyContent{
				Blocks: Blocks{
					&Block{Type: "noodle"},
					&Block{Type: "noodle"},
				},
			},
			Target: makeInstantiateType(struct {
				Noodle []struct {
					Name string `hclext:"name,label"`
				} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle []struct {
					Name string `hclext:"name,label"`
				} `hclext:"noodle,block"`
			}{Noodle: []struct {
				Name string `hclext:"name,label"`
			}{{Name: ""}, {Name: ""}}}),
			DiagCount: 2, // label is required by default
		},
		{
			Name: "multiple single labeled blocks with labeled block schema",
			Body: &BodyContent{
				Blocks: Blocks{
					&Block{Type: "noodle", Labels: []string{"foo"}},
					&Block{Type: "noodle", Labels: []string{"bar"}},
				},
			},
			Target: makeInstantiateType(struct {
				Noodle []struct {
					Name string `hclext:"name,label"`
				} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle []struct {
					Name string `hclext:"name,label"`
				} `hclext:"noodle,block"`
			}{Noodle: []struct {
				Name string `hclext:"name,label"`
			}{{Name: "foo"}, {Name: "bar"}}}),
			DiagCount: 0,
		},
		{
			Name: "labeled block with label/attr schema",
			Body: &BodyContent{
				Blocks: Blocks{
					&Block{
						Type:   "noodle",
						Labels: []string{"foo"},
						Body: &BodyContent{
							Attributes: Attributes{"type": &Attribute{Name: "type", Expr: parseExpr(`"rice"`)}},
						},
					},
				},
			},
			Target: makeInstantiateType(struct {
				Noodle struct {
					Name string `hclext:"name,label"`
					Type string `hclext:"type"`
				} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle struct {
					Name string `hclext:"name,label"`
					Type string `hclext:"type"`
				} `hclext:"noodle,block"`
			}{Noodle: struct {
				Name string `hclext:"name,label"`
				Type string `hclext:"type"`
			}{Name: "foo", Type: "rice"}}),
			DiagCount: 0,
		},
		{
			Name: "nested block",
			Body: &BodyContent{
				Blocks: Blocks{
					&Block{
						Type:   "noodle",
						Labels: []string{"foo"},
						Body: &BodyContent{
							Attributes: Attributes{"type": &Attribute{Name: "type", Expr: parseExpr(`"rice"`)}},
							Blocks: Blocks{
								&Block{
									Type:   "bread",
									Labels: []string{"bar"},
									Body: &BodyContent{
										Attributes: Attributes{"baked": &Attribute{Name: "baked", Expr: parseExpr(`true`)}},
									},
								},
							},
						},
					},
				},
			},
			Target: makeInstantiateType(struct {
				Noodle struct {
					Name  string `hclext:"name,label"`
					Type  string `hclext:"type"`
					Bread struct {
						Name  string `hclext:"name,label"`
						Baked bool   `hclext:"baked"`
					} `hclext:"bread,block"`
				} `hclext:"noodle,block"`
			}{}),
			Check: equals(struct {
				Noodle struct {
					Name  string `hclext:"name,label"`
					Type  string `hclext:"type"`
					Bread struct {
						Name  string `hclext:"name,label"`
						Baked bool   `hclext:"baked"`
					} `hclext:"bread,block"`
				} `hclext:"noodle,block"`
			}{Noodle: struct {
				Name  string `hclext:"name,label"`
				Type  string `hclext:"type"`
				Bread struct {
					Name  string `hclext:"name,label"`
					Baked bool   `hclext:"baked"`
				} `hclext:"bread,block"`
			}{
				Name: "foo",
				Type: "rice",
				Bread: struct {
					Name  string `hclext:"name,label"`
					Baked bool   `hclext:"baked"`
				}{
					Name:  "bar",
					Baked: true,
				},
			}}),
			DiagCount: 0,
		},
		{
			Name: "retain nested block",
			Body: &BodyContent{
				Attributes: Attributes{"plain": &Attribute{Name: "plain", Expr: parseExpr(`"foo"`)}},
			},
			Target: func() interface{} {
				return &withNestedBlock{
					Plain: "bar",
					Nested: &withTwoAttributes{
						A: "bar",
					},
				}
			},
			Check: equals(withNestedBlock{
				Plain: "foo",
				Nested: &withTwoAttributes{
					A: "bar",
				},
			}),
			DiagCount: 0,
		},
		{
			Name: "retain attrs in nested block",
			Body: &BodyContent{
				Blocks: Blocks{
					&Block{
						Type: "nested",
						Body: &BodyContent{
							Attributes: Attributes{"a": &Attribute{Name: "a", Expr: parseExpr(`"foo"`)}},
						},
					},
				},
			},
			Target: func() interface{} {
				return &withNestedBlock{
					Nested: &withTwoAttributes{
						B: "bar",
					},
				}
			},
			Check: equals(withNestedBlock{
				Nested: &withTwoAttributes{
					A: "foo",
					B: "bar",
				},
			}),
			DiagCount: 0,
		},
		{
			Name: "retain attrs in multiple nested blocks",
			Body: &BodyContent{
				Blocks: Blocks{
					&Block{
						Type: "nested",
						Body: &BodyContent{
							Attributes: Attributes{"a": &Attribute{Name: "a", Expr: parseExpr(`"foo"`)}},
						},
					},
				},
			},
			Target: func() interface{} {
				return &withListofNestedBlocks{
					Nested: []*withTwoAttributes{
						{B: "bar"},
					},
				}
			},
			Check: equals(withListofNestedBlocks{
				Nested: []*withTwoAttributes{
					{A: "foo", B: "bar"},
				},
			}),
			DiagCount: 0,
		},
		{
			Name: "remove additional elements from the list while decoding nested blocks",
			Body: &BodyContent{
				Blocks: Blocks{
					&Block{
						Type: "nested",
						Body: &BodyContent{
							Attributes: Attributes{"a": &Attribute{Name: "a", Expr: parseExpr(`"foo"`)}},
						},
					},
				},
			},
			Target: func() interface{} {
				return &withListofNestedBlocks{
					Nested: []*withTwoAttributes{
						{B: "bar"},
						{B: "bar"},
					},
				}
			},
			Check: equals(withListofNestedBlocks{
				Nested: []*withTwoAttributes{
					{A: "foo", B: "bar"},
				},
			}),
			DiagCount: 0,
		},
		{
			Name: "remove additional elements from the list while decoding nested blocks even if target are not pointer slices",
			Body: &BodyContent{
				Blocks: Blocks{
					&Block{
						Type: "nested",
						Body: &BodyContent{
							Attributes: Attributes{"b": &Attribute{Name: "b", Expr: parseExpr(`"bar"`)}},
						},
					},
					&Block{
						Type: "nested",
						Body: &BodyContent{
							Attributes: Attributes{"b": &Attribute{Name: "b", Expr: parseExpr(`"baz"`)}},
						},
					},
				},
			},
			Target: func() interface{} {
				return &withListofNestedBlocksNoPointers{
					Nested: []withTwoAttributes{
						{B: "foo"},
					},
				}
			},
			Check: equals(withListofNestedBlocksNoPointers{
				Nested: []withTwoAttributes{
					{B: "bar"},
					{B: "baz"},
				},
			}),
			DiagCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			targetVal := reflect.ValueOf(test.Target())

			diags := DecodeBody(test.Body, nil, targetVal.Interface())
			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}
			got := targetVal.Elem().Interface()
			if !test.Check(got) {
				t.Errorf("wrong result\ndiff:  %#v", got)
			}
		})
	}
}
