package schema

import "github.com/hashicorp/hcl/v2"

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

	ExprBytes []byte
}
