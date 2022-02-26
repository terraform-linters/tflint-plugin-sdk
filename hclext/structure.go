package hclext

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
)

// BodyContent is the result of applying a hclext.BodySchema to a hcl.Body.
// Unlike hcl.BodyContent, this does not have MissingItemRange.
// This difference is because hcl.BodyContent is the result for a single HCL file,
// while hclext.BodyContent is the result for a Terraform module.
type BodyContent struct {
	Attributes Attributes
	Blocks     Blocks
}

// Blocks is a sequence of Block.
type Blocks []*Block

// Block represents a nested block within a hcl.Body.
// Unlike hcl.Block, this has Body as hclext.BodyContent (struct), not hcl.Body (interface).
// Since interface is hard to send over a wire protocol, it is designed to always return only the attributes based on the schema.
// Instead, the hclext.BlockSchema can now be nested to extract the attributes within the nested block.
type Block struct {
	Type   string
	Labels []string
	Body   *BodyContent

	DefRange    hcl.Range
	TypeRange   hcl.Range
	LabelRanges []hcl.Range
}

// Attributes is a set of attributes keyed by their names.
// Please note that this is not strictly. Since hclext.BodyContent is the body from multiple files,
// top-level attributes can have the same name (it is not possible to specify the same name within a block).
// This exception is not considered here, as Terraform syntax does not allow top-level attributes.
type Attributes map[string]*Attribute

// Attribute represents an attribute from within a body.
type Attribute struct {
	Name string
	Expr hcl.Expression

	Range     hcl.Range
	NameRange hcl.Range
}

// Content is a wrapper for hcl.Content for working with nested schemas.
// Convert hclext.BodySchema to hcl.BodySchema, and convert hcl.BodyContent
// to hclext.BodyContent. It processes the nested body recursively.
func Content(body hcl.Body, schema *BodySchema) (*BodyContent, hcl.Diagnostics) {
	if reflect.ValueOf(body).IsNil() {
		return &BodyContent{}, hcl.Diagnostics{}
	}
	if schema == nil {
		schema = &BodySchema{}
	}

	hclS := &hcl.BodySchema{
		Attributes: make([]hcl.AttributeSchema, len(schema.Attributes)),
		Blocks:     make([]hcl.BlockHeaderSchema, len(schema.Blocks)),
	}
	for idx, attrS := range schema.Attributes {
		hclS.Attributes[idx] = hcl.AttributeSchema{Name: attrS.Name, Required: attrS.Required}
	}
	childS := map[string]*BodySchema{}
	for idx, blockS := range schema.Blocks {
		hclS.Blocks[idx] = hcl.BlockHeaderSchema{Type: blockS.Type, LabelNames: blockS.LabelNames}
		childS[blockS.Type] = blockS.Body
	}

	content, diags := body.Content(hclS)

	ret := &BodyContent{
		Attributes: Attributes{},
		Blocks:     make(Blocks, len(content.Blocks)),
	}
	for name, attr := range content.Attributes {
		ret.Attributes[name] = &Attribute{
			Name:      attr.Name,
			Expr:      attr.Expr,
			Range:     attr.Range,
			NameRange: attr.NameRange,
		}
	}
	for idx, block := range content.Blocks {
		child, childDiags := Content(block.Body, childS[block.Type])
		diags = diags.Extend(childDiags)

		ret.Blocks[idx] = &Block{
			Type:        block.Type,
			Labels:      block.Labels,
			Body:        child,
			DefRange:    block.DefRange,
			TypeRange:   block.TypeRange,
			LabelRanges: block.LabelRanges,
		}
	}

	return ret, diags
}

// PartialContent is a wrapper for hcl.PartialContent for working with nested schemas.
// Convert hclext.BodySchema to hcl.BodySchema, and convert hcl.BodyContent
// to hclext.BodyContent. It processes the nested body recursively.
// Unlike hcl.PartialContent, it does not return the rest of the body.
func PartialContent(body hcl.Body, schema *BodySchema) (*BodyContent, hcl.Diagnostics) {
	if reflect.ValueOf(body).IsNil() {
		return &BodyContent{}, hcl.Diagnostics{}
	}
	if schema == nil {
		schema = &BodySchema{}
	}

	hclS := &hcl.BodySchema{
		Attributes: make([]hcl.AttributeSchema, len(schema.Attributes)),
		Blocks:     make([]hcl.BlockHeaderSchema, len(schema.Blocks)),
	}
	for idx, attrS := range schema.Attributes {
		hclS.Attributes[idx] = hcl.AttributeSchema{Name: attrS.Name, Required: attrS.Required}
	}
	childS := map[string]*BodySchema{}
	for idx, blockS := range schema.Blocks {
		hclS.Blocks[idx] = hcl.BlockHeaderSchema{Type: blockS.Type, LabelNames: blockS.LabelNames}
		childS[blockS.Type] = blockS.Body
	}

	content, _, diags := body.PartialContent(hclS)

	ret := &BodyContent{
		Attributes: Attributes{},
		Blocks:     make(Blocks, len(content.Blocks)),
	}
	for name, attr := range content.Attributes {
		ret.Attributes[name] = &Attribute{
			Name:      attr.Name,
			Expr:      attr.Expr,
			Range:     attr.Range,
			NameRange: attr.NameRange,
		}
	}
	for idx, block := range content.Blocks {
		child, childDiags := PartialContent(block.Body, childS[block.Type])
		diags = diags.Extend(childDiags)

		ret.Blocks[idx] = &Block{
			Type:        block.Type,
			Labels:      block.Labels,
			Body:        child,
			DefRange:    block.DefRange,
			TypeRange:   block.TypeRange,
			LabelRanges: block.LabelRanges,
		}
	}

	return ret, diags
}

// ByType transforms the receiving block sequence into a map from type
// name to block sequences of only that type.
func (els Blocks) ByType() map[string]Blocks {
	ret := make(map[string]Blocks)
	for _, el := range els {
		ty := el.Type
		if ret[ty] == nil {
			ret[ty] = make(Blocks, 0, 1)
		}
		ret[ty] = append(ret[ty], el)
	}
	return ret
}
