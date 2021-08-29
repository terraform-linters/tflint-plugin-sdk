package schema

import (
	"github.com/hashicorp/hcl/v2"
)

type BodyContent struct {
	Attributes Attributes
	Blocks     Blocks

	MissingItemRange hcl.Range
}

type Blocks []*Block

type Block struct {
	Type   string
	Labels []string
	Body   *BodyContent

	DefRange    hcl.Range
	TypeRange   hcl.Range
	LabelRanges []hcl.Range
}

type Attributes map[string]*Attribute

type Attribute struct {
	Name string
	Expr hcl.Expression

	Range     hcl.Range
	NameRange hcl.Range
}

func Content(body hcl.Body, schema *BodySchema) (*BodyContent, hcl.Diagnostics) {
	hclS := &hcl.BodySchema{
		Attributes: make([]hcl.AttributeSchema, len(schema.Attributes)),
		Blocks:     make([]hcl.BlockHeaderSchema, len(schema.Blocks)),
	}
	for idx, attrS := range schema.Attributes {
		hclS.Attributes[idx] = hcl.AttributeSchema{Name: attrS.Name}
	}
	childS := map[string]*BodySchema{}
	for idx, blockS := range schema.Blocks {
		hclS.Blocks[idx] = hcl.BlockHeaderSchema{Type: blockS.Type, LabelNames: blockS.LabelNames}
		childS[blockS.Type] = blockS.Body
	}

	content, _, diags := body.PartialContent(hclS)

	ret := &BodyContent{
		Attributes:       Attributes{},
		Blocks:           make(Blocks, len(content.Blocks)),
		MissingItemRange: content.MissingItemRange,
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
