package hclext

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
)

// ParseExpression is a wrapper that calls ParseExpression of hclsyntax and json based on the file extension.
// This function specializes in parsing intermediate expressions in the file,
// so it takes into account the hack on trailing newlines in heredoc.
func ParseExpression(src []byte, filename string, start hcl.Pos) (hcl.Expression, hcl.Diagnostics) {
	// Handle HCL files: .tf (Terraform HCL) and .hcl (HCL config like .tflint.hcl)
	if strings.HasSuffix(filename, ".tf") || strings.HasSuffix(filename, ".hcl") {
		// HACK: Always add a newline to avoid heredoc parse errors.
		// @see https://github.com/hashicorp/hcl/issues/441
		src = []byte(string(src) + "\n")
		return hclsyntax.ParseExpression(src, filename, start)
	}

	// Handle JSON files:
	// We accept any .json file (including .tf.json), not just specific ones like .tflint.json.
	// The calling functions are responsible for validating that the file should be processed.
	// If the content is not valid HCL-compatible JSON, the JSON parser will return appropriate diagnostics.
	if strings.HasSuffix(filename, ".tf.json") || strings.HasSuffix(filename, ".json") {
		return json.ParseExpressionWithStartPos(src, filename, start)
	}

	return nil, hcl.Diagnostics{
		{
			Severity: hcl.DiagError,
			Summary:  "Unexpected file extension",
			Detail:   fmt.Sprintf("The file name `%s` is a file with an unexpected extension. Valid extensions are `.tf`, `.tf.json`, `.hcl`, and `.json`.", filename),
		},
	}
}
