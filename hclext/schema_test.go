package hclext

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestImpliedBodySchema(t *testing.T) {
	tests := []struct {
		Name string
		Val  interface{}
		Want *BodySchema
	}{
		{
			Name: "empty struct",
			Val:  struct{}{},
			Want: &BodySchema{},
		},
		{
			Name: "struct without tags",
			Val: struct {
				Ignored bool
			}{},
			Want: &BodySchema{},
		},
		{
			Name: "attribute tags",
			Val: struct {
				Attr1 bool `hclext:"attr1"`
				Attr2 bool `hclext:"attr2"`
			}{},
			Want: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name:     "attr1",
						Required: true,
					},
					{
						Name:     "attr2",
						Required: true,
					},
				},
			},
		},
		{
			Name: "pointer attribute tags",
			Val: struct {
				Attr *bool `hclext:"attr,attr"`
			}{},
			Want: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name:     "attr",
						Required: false,
					},
				},
			},
		},
		{
			Name: "optional attribute tags",
			Val: struct {
				Attr bool `hclext:"attr,optional"`
			}{},
			Want: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name:     "attr",
						Required: false,
					},
				},
			},
		},
		{
			Name: "block tags",
			Val: struct {
				Thing struct{} `hclext:"thing,block"`
			}{},
			Want: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type: "thing",
						Body: &BodySchema{},
					},
				},
			},
		},
		{
			Name: "block tags with labels",
			Val: struct {
				Thing struct {
					Type string `hclext:"type,label"`
					Name string `hclext:"name,label"`
				} `hclext:"thing,block"`
			}{},
			Want: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type:       "thing",
						LabelNames: []string{"type", "name"},
						Body:       &BodySchema{},
					},
				},
			},
		},
		{
			Name: "multiple block tags with labels",
			Val: struct {
				Thing []struct {
					Type string `hclext:"type,label"`
					Name string `hclext:"name,label"`
				} `hclext:"thing,block"`
			}{},
			Want: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type:       "thing",
						LabelNames: []string{"type", "name"},
						Body:       &BodySchema{},
					},
				},
			},
		},
		{
			Name: "pointer block tags with labels",
			Val: struct {
				Thing *struct {
					Type string `hclext:"type,label"`
					Name string `hclext:"name,label"`
				} `hclext:"thing,block"`
			}{},
			Want: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type:       "thing",
						LabelNames: []string{"type", "name"},
						Body:       &BodySchema{},
					},
				},
			},
		},
		{
			Name: "nested block tags with labels",
			Val: struct {
				Thing struct {
					Name      string `hclext:"name,label"`
					Something string `hclext:"something"`
				} `hclext:"thing,block"`
			}{},
			Want: &BodySchema{
				Blocks: []BlockSchema{
					{
						Type:       "thing",
						LabelNames: []string{"name"},
						Body: &BodySchema{
							Attributes: []AttributeSchema{
								{
									Name:     "something",
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "attribute/block tags with labels",
			Val: struct {
				Doodad string `hclext:"doodad"`
				Thing  struct {
					Name string `hclext:"name,label"`
				} `hclext:"thing,block"`
			}{},
			Want: &BodySchema{
				Attributes: []AttributeSchema{
					{
						Name:     "doodad",
						Required: true,
					},
				},
				Blocks: []BlockSchema{
					{
						Type:       "thing",
						LabelNames: []string{"name"},
						Body:       &BodySchema{},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := ImpliedBodySchema(test.Val)
			if diff := cmp.Diff(test.Want, got); diff != "" {
				t.Errorf("wrong schema\ndiff:  %s", diff)
			}
		})
	}
}
