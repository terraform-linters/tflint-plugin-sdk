package toproto

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/schema"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
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

func ApplyConfig_Request(body *schema.BodyContent, sources map[string][]byte) *proto.ApplyConfig_Request {
	return &proto.ApplyConfig_Request{
		Body: ApplyConfig_Request_Body(body, sources),
	}
}

func ApplyConfig_Request_Body(body *schema.BodyContent, sources map[string][]byte) *proto.BodyContent {
	attributes := map[string]*proto.BodyContent_Attribute{}
	for key, attr := range body.Attributes {
		attributes[key] = &proto.BodyContent_Attribute{
			Name:      attr.Name,
			Expr:      attr.Expr.Range().SliceBytes(sources[attr.Range.Filename]),
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
			Body:        ApplyConfig_Request_Body(block.Body, sources),
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

func BodyContent(body *schema.BodyContent, sources map[string][]byte) *proto.BodyContent {
	attributes := map[string]*proto.BodyContent_Attribute{}
	for idx, attr := range body.Attributes {
		attributes[idx] = &proto.BodyContent_Attribute{
			Name:      attr.Name,
			Expr:      attr.Expr.Range().SliceBytes(sources[attr.Range.Filename]),
			Range:     Range(attr.Range),
			NameRange: Range(attr.NameRange),
			ExprRange: Range(attr.Expr.Range()),
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
			Body:        BodyContent(block.Body, sources),
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

func EmitIssue_Rule(rule tflint.Rule) *proto.EmitIssue_Rule {
	return &proto.EmitIssue_Rule{
		Name:     rule.Name(),
		Enabled:  rule.Enabled(),
		Severity: EmitIssue_Severity(rule.Severity()),
		Link:     rule.Link(),
	}
}

func EmitIssue_Severity(severity string) proto.EmitIssue_Severity {
	switch severity {
	case tflint.ERROR:
		return proto.EmitIssue_ERROR
	case tflint.WARNING:
		return proto.EmitIssue_WARNING
	case tflint.NOTICE:
		return proto.EmitIssue_NOTICE
	}

	return proto.EmitIssue_ERROR
}
