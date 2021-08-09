package fromproto

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/schema"
)

func BodySchema(body *proto.BodySchema) *schema.BodySchema {
	if body == nil {
		return nil
	}

	attributes := make([]schema.AttributeSchema, len(body.Attributes))
	for idx, attr := range body.Attributes {
		attributes[idx] = schema.AttributeSchema{Name: attr.Name}
	}

	blocks := make([]schema.BlockSchema, len(body.Blocks))
	for idx, block := range body.Blocks {
		blocks[idx] = schema.BlockSchema{
			Type:       block.Type,
			LabelNames: block.LabelNames,
			Body:       BodySchema(block.Body),
		}
	}

	return &schema.BodySchema{
		Attributes: attributes,
		Blocks:     blocks,
	}
}

func BodyContent(body *proto.BodyContent) (*schema.BodyContent, hcl.Diagnostics) {
	if body == nil {
		return nil, nil
	}

	attributes := schema.Attributes{}
	for key, attr := range body.Attributes {
		expr, diags := parseExpression(attr.Expr, attr.Range.Filename, Pos(attr.Range.Start))
		// TODO: append
		if diags.HasErrors() {
			return nil, diags
		}

		attributes[key] = &schema.Attribute{
			Name:      attr.Name,
			Expr:      expr,
			Range:     Range(attr.Range),
			NameRange: Range(attr.NameRange),
			ExprBytes: attr.Expr,
		}
	}

	blocks := make(schema.Blocks, len(body.Blocks))
	for idx, block := range body.Blocks {
		blockBody, diags := BodyContent(block.Body)
		// TODO: append
		if diags.HasErrors() {
			return nil, diags
		}

		labelRanges := make([]hcl.Range, len(block.LabelRanges))
		for idx, labelRange := range block.LabelRanges {
			labelRanges[idx] = Range(labelRange)
		}

		blocks[idx] = &schema.Block{
			Type:        block.Type,
			Labels:      block.Labels,
			Body:        blockBody,
			DefRange:    Range(block.DefRange),
			TypeRange:   Range(block.TypeRange),
			LabelRanges: labelRanges,
		}
	}

	return &schema.BodyContent{
		Attributes:       attributes,
		Blocks:           blocks,
		MissingItemRange: Range(body.MissingItemRange),
	}, nil
}

func Range(rng *proto.Range) hcl.Range {
	if rng == nil {
		return hcl.Range{}
	}

	return hcl.Range{
		Filename: rng.Filename,
		Start:    Pos(rng.Start),
		End:      Pos(rng.End),
	}
}

func Pos(pos *proto.Range_Pos) hcl.Pos {
	if pos == nil {
		return hcl.Pos{}
	}

	return hcl.Pos{
		Line:   int(pos.Line),
		Column: int(pos.Column),
		Byte:   int(pos.Byte),
	}
}

// TODO: Move to another package
func parseExpression(src []byte, filename string, start hcl.Pos) (hcl.Expression, hcl.Diagnostics) {
	if strings.HasSuffix(filename, ".tf") {
		// HACK: Always add a newline to avoid heredoc parse errors.
		// @see https://github.com/hashicorp/hcl/issues/441
		src = []byte(string(src) + "\n")
		return hclsyntax.ParseExpression(src, filename, start)
	}

	if strings.HasSuffix(filename, ".tf.json") {
		return json.ParseExpressionWithStartPos(src, filename, start)
	}

	panic(fmt.Sprintf("Unexpected file: %s", filename))
}
