package toproto

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/schema"
)

func ConfigSchema_Response(body *schema.BodySchema) *proto.ConfigSchema_Response {
	return &proto.ConfigSchema_Response{
		Body: BodySchema(body),
	}
}

func BodySchema(body *schema.BodySchema) *proto.BodySchema {
	attributes := make([]*proto.BodySchema_Attribute, len(body.Attributes))
	for idx, attr := range body.Attributes {
		attributes[idx] = &proto.BodySchema_Attribute{Name: attr.Name}
	}

	blocks := make([]*proto.BodySchema_Block, len(body.Blocks))
	for idx, block := range body.Blocks {
		blocks[idx] = &proto.BodySchema_Block{
			Type:       block.Type,
			LabelNames: block.LabelNames,
			Body:       BodySchema(block.Body),
		}
	}

	return &proto.BodySchema{
		Attributes: attributes,
		Blocks:     blocks,
	}
}

func ApplyConfig_Request(body *schema.BodyContent) *proto.ApplyConfig_Request {
	return &proto.ApplyConfig_Request{
		Body: ApplyConfig_Request_Body(body),
	}
}

func ApplyConfig_Request_Body(body *schema.BodyContent) *proto.BodyContent {
	attributes := map[string]*proto.BodyContent_Attribute{}
	for key, attr := range body.Attributes {
		attributes[key] = &proto.BodyContent_Attribute{
			Name:      attr.Name,
			Expr:      attr.ExprBytes,
			Range:     Range(attr.Range),
			NameRange: Range(attr.NameRange),
		}
	}

	blocks := make([]*proto.BodyContent_Block, len(body.Blocks))
	for idx, block := range body.Blocks {
		labelRanges := make([]*proto.Range, len(block.LabelRanges))
		for i, labelRange := range block.LabelRanges {
			labelRanges[i] = Range(labelRange)
		}

		blocks[idx] = &proto.BodyContent_Block{
			Type:        block.Type,
			Labels:      block.Labels,
			Body:        ApplyConfig_Request_Body(block.Body),
			DefRange:    Range(block.DefRange),
			TypeRange:   Range(block.TypeRange),
			LabelRanges: labelRanges,
		}
	}

	return &proto.BodyContent{
		Attributes:       attributes,
		Blocks:           blocks,
		MissingItemRange: Range(body.MissingItemRange),
	}
}

func Range(rng hcl.Range) *proto.Range {
	return &proto.Range{
		Filename: rng.Filename,
		Start:    Range_Pos(rng.Start),
		End:      Range_Pos(rng.End),
	}
}

func Range_Pos(pos hcl.Pos) *proto.Range_Pos {
	return &proto.Range_Pos{
		Line:   int64(pos.Line),
		Column: int64(pos.Column),
		Byte:   int64(pos.Byte),
	}
}

func BodyContent(body *schema.BodyContent) *proto.BodyContent {
	attributes := map[string]*proto.BodyContent_Attribute{}
	for idx, attr := range body.Attributes {
		attributes[idx] = &proto.BodyContent_Attribute{
			Name:      attr.Name,
			Expr:      attr.Expr.Range().SliceBytes(attr.ExprBytes),
			Range:     Range(attr.Range),
			NameRange: Range(attr.NameRange),
		}
	}

	blocks := make([]*proto.BodyContent_Block, len(body.Blocks))
	for idx, block := range body.Blocks {
		labelRanges := make([]*proto.Range, len(block.LabelRanges))
		for idx, labelRange := range block.LabelRanges {
			labelRanges[idx] = Range(labelRange)
		}

		blocks[idx] = &proto.BodyContent_Block{
			Type:        block.Type,
			Labels:      block.Labels,
			Body:        BodyContent(block.Body),
			DefRange:    Range(block.DefRange),
			TypeRange:   Range(block.TypeRange),
			LabelRanges: labelRanges,
		}
	}

	return &proto.BodyContent{
		Attributes: attributes,
		Blocks:     blocks,
	}
}
