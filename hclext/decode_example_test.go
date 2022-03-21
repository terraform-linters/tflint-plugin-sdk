package hclext

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func ExampleDecodeBody() {
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

	type Bread struct {
		// The `*,label` tag matches "bread" block labels.
		// The count of tags should be matched to count of block labels.
		Name string `hclext:"name,label"`
		// The `type` tag matches a "type" attribute inside of "bread" block.
		Type string `hclext:"type"`
		// The `baked,optional` tag matches a "baked" attribute, but it is optional.
		Baked bool `hclext:"baked,optional"`
	}
	type Noodle struct {
		Name    string `hclext:"name,label"`
		SubName string `hclext:"subname,label"`
		Type    string `hclext:"type"`
		// The `bread,block` tag matches "bread" blocks.
		// Multiple blocks are allowed because the field type is slice.
		Breads []Bread `hclext:"bread,block"`
	}
	type Config struct {
		// Only 1 block must be needed because the field type is not slice, not a pointer.
		Noodle Noodle `hclext:"noodle,block"`
	}

	target := &Config{}

	schema := ImpliedBodySchema(target)
	body, diags := Content(file.Body, schema)
	if diags.HasErrors() {
		panic(diags)
	}

	diags = DecodeBody(body, nil, target)
	if diags.HasErrors() {
		panic(diags)
	}

	fmt.Printf("- noodle: name=%s, subname=%s type=%s\n", target.Noodle.Name, target.Noodle.SubName, target.Noodle.Type)
	for i, bread := range target.Noodle.Breads {
		fmt.Printf("  - bread[%d]: name=%s, type=%s baked=%t\n", i, bread.Name, bread.Type, bread.Baked)
	}
	// Output:
	// - noodle: name=foo, subname=bar type=rice
	//   - bread[0]: name=baz, type=focaccia baked=true
	//   - bread[1]: name=quz, type=rye baked=false
}
