package hclext

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func ExampleContent() {
	src := `
noodle "foo" "bar" {
	type = "rice"

	bread "baz" {
		type  = "focaccia"
		baked = true
	}
	bread "quz" {
		type = "rye"
	}
}`
	file, diags := hclsyntax.ParseConfig([]byte(src), "test.tf", hcl.InitialPos)
	if diags.HasErrors() {
		panic(diags)
	}

	body, diags := Content(file.Body, &BodySchema{
		Blocks: []BlockSchema{
			{
				Type:       "noodle",
				LabelNames: []string{"name", "subname"},
				Body: &BodySchema{
					Attributes: []AttributeSchema{{Name: "type"}},
					Blocks: []BlockSchema{
						{
							Type:       "bread",
							LabelNames: []string{"name"},
							Body: &BodySchema{
								Attributes: []AttributeSchema{
									{Name: "type", Required: true},
									{Name: "baked"},
								},
							},
						},
					},
				},
			},
		},
	})
	if diags.HasErrors() {
		panic(diags)
	}

	for i, noodle := range body.Blocks {
		fmt.Printf("- noodle[%d]: labels=%s, attributes=%d\n", i, noodle.Labels, len(noodle.Body.Attributes))
		for i, bread := range noodle.Body.Blocks {
			fmt.Printf("  - bread[%d]: labels=%s, attributes=%d\n", i, bread.Labels, len(bread.Body.Attributes))
		}
	}
	// Output:
	// - noodle[0]: labels=[foo bar], attributes=1
	//   - bread[0]: labels=[baz], attributes=2
	//   - bread[1]: labels=[quz], attributes=1
}

func ExamplePartialContent() {
	src := `
noodle "foo" "bar" {
	type = "rice"

	bread "baz" {
		type  = "focaccia"
		baked = true
	}
	bread "quz" {
		type = "rye"
	}
}`
	file, diags := hclsyntax.ParseConfig([]byte(src), "test.tf", hcl.InitialPos)
	if diags.HasErrors() {
		panic(diags)
	}

	body, diags := PartialContent(file.Body, &BodySchema{
		Blocks: []BlockSchema{
			{
				Type:       "noodle",
				LabelNames: []string{"name", "subname"},
				Body: &BodySchema{
					Blocks: []BlockSchema{
						{
							Type:       "bread",
							LabelNames: []string{"name"},
							Body: &BodySchema{
								Attributes: []AttributeSchema{
									{Name: "type", Required: true},
								},
							},
						},
					},
				},
			},
		},
	})
	if diags.HasErrors() {
		panic(diags)
	}

	for i, noodle := range body.Blocks {
		fmt.Printf("- noodle[%d]: labels=%s, attributes=%d\n", i, noodle.Labels, len(noodle.Body.Attributes))
		for i, bread := range noodle.Body.Blocks {
			fmt.Printf("  - bread[%d]: labels=%s, attributes=%d\n", i, bread.Labels, len(bread.Body.Attributes))
		}
	}
	// Output:
	// - noodle[0]: labels=[foo bar], attributes=0
	//   - bread[0]: labels=[baz], attributes=1
	//   - bread[1]: labels=[quz], attributes=1
}
