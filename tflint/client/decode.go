package client

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
)

// Attribute is an intermediate representation of hcl.Attribute.
type Attribute struct {
	Name      string
	Expr      []byte
	ExprRange hcl.Range
	Range     hcl.Range
	NameRange hcl.Range
}

func decodeAttribute(attribute *Attribute) (*hcl.Attribute, hcl.Diagnostics) {
	expr, diags := parseExpression(attribute.Expr, attribute.ExprRange.Filename, attribute.ExprRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &hcl.Attribute{
		Name:      attribute.Name,
		Expr:      expr,
		Range:     attribute.Range,
		NameRange: attribute.NameRange,
	}, nil
}

// Block is an intermediate representation of hcl.Block.
type Block struct {
	Type      string
	Labels    []string
	Body      []byte
	BodyRange hcl.Range

	DefRange    hcl.Range
	TypeRange   hcl.Range
	LabelRanges []hcl.Range
}

func decodeBlock(block *Block) (*hcl.Block, hcl.Diagnostics) {
	file, diags := parseConfig(block.Body, block.BodyRange.Filename, block.BodyRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &hcl.Block{
		Type:        block.Type,
		Labels:      block.Labels,
		Body:        file.Body,
		DefRange:    block.DefRange,
		TypeRange:   block.TypeRange,
		LabelRanges: block.LabelRanges,
	}, nil
}

func parseExpression(src []byte, filename string, start hcl.Pos) (hcl.Expression, hcl.Diagnostics) {
	if strings.HasSuffix(filename, ".tf") {
		return hclsyntax.ParseExpression(src, filename, start)
	}

	if strings.HasSuffix(filename, ".tf.json") {
		return json.ParseExpressionWithStartPos(src, filename, start)
	}

	panic(fmt.Sprintf("Unexpected file: %s", filename))
}

func parseConfig(src []byte, filename string, start hcl.Pos) (*hcl.File, hcl.Diagnostics) {
	if strings.HasSuffix(filename, ".tf") {
		return hclsyntax.ParseConfig(src, filename, start)
	}

	if strings.HasSuffix(filename, ".tf.json") {
		return json.ParseWithStartPos(src, filename, start)
	}

	panic(fmt.Sprintf("Unexpected file: %s", filename))
}

func parseVersionConstraint(versionStr string, versionRange hcl.Range) (configs.VersionConstraint, hcl.Diagnostics) {
	versionConstraint := configs.VersionConstraint{DeclRange: versionRange}
	if !versionRange.Empty() {
		required, err := version.NewConstraint(versionStr)
		if err != nil {
			detail := fmt.Sprintf(
				"Version constraint '%s' parse error: %s",
				versionStr,
				err,
			)

			return versionConstraint, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Failed to reparse version constraint",
					Detail:   detail,
					Subject:  &versionRange,
				},
			}
		}

		versionConstraint.Required = required
	}
	return versionConstraint, nil
}
