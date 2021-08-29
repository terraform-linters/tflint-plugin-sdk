package fromproto

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/schema"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
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
		expr, diags := parseExpression(attr.Expr, attr.ExprRange.Filename, Pos(attr.ExprRange.Start))
		// TODO: append
		if diags.HasErrors() {
			return nil, diags
		}

		attributes[key] = &schema.Attribute{
			Name:      attr.Name,
			Expr:      expr,
			Range:     Range(attr.Range),
			NameRange: Range(attr.NameRange),
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

type Rule struct {
	Data RuleData
}

type RuleData struct {
	Name     string
	Enabled  bool
	Severity string // TODO: enum?
	Link     string
}

func (r *Rule) Name() string              { return r.Data.Name }
func (r *Rule) Enabled() bool             { return r.Data.Enabled }
func (r *Rule) Severity() string          { return r.Data.Severity }
func (r *Rule) Link() string              { return r.Data.Link }
func (r *Rule) Check(tflint.Runner) error { return nil }

func EmitIssue_Rule(rule *proto.EmitIssue_Rule) *Rule {
	if rule == nil {
		return nil
	}

	return &Rule{
		Data: RuleData{
			Name:     rule.Name,
			Enabled:  rule.Enabled,
			Severity: EmitIssue_Severity(rule.Severity),
			Link:     rule.Link,
		},
	}
}

func EmitIssue_Severity(severity proto.EmitIssue_Severity) string {
	switch severity {
	case proto.EmitIssue_ERROR:
		return tflint.ERROR
	case proto.EmitIssue_WARNING:
		return tflint.WARNING
	case proto.EmitIssue_NOTICE:
		return tflint.NOTICE
	}

	return tflint.ERROR
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
