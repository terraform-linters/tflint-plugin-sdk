package runner

import (
	stdjson "encoding/json"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// WalkExpressions traverses expressions in the given files by the passed
// walker. Note that it behaves differently in native HCL syntax and JSON
// syntax.
//
// In the HCL syntax, `var.foo` and `var.bar` in `[var.foo, var.bar]` are
// also passed to the walker. In other words, it traverses expressions
// recursively. To avoid redundant checks, the walker should check the kind
// of expression.
//
// In the JSON syntax, only an expression of an attribute seen from the top
// level of the file is passed. In other words, it doesn't traverse
// expressions recursively. This is a limitation of JSON syntax.
func WalkExpressions(files map[string]*hcl.File, walker tflint.ExprWalker) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	for _, file := range files {
		if body, ok := file.Body.(*hclsyntax.Body); ok {
			walkDiags := hclsyntax.Walk(body, &nativeWalker{walker: walker})
			diags = diags.Extend(walkDiags)
			continue
		}

		// In JSON syntax, everything can be walked as an attribute.
		attrs, jsonDiags := getJSONAttributes(file.Body, file.Bytes)
		if jsonDiags.HasErrors() {
			diags = diags.Extend(jsonDiags)
			continue
		}

		for _, attr := range attrs {
			enterDiags := walker.Enter(attr.Expr)
			diags = diags.Extend(enterDiags)
			exitDiags := walker.Exit(attr.Expr)
			diags = diags.Extend(exitDiags)
		}
	}

	return diags
}

type nativeWalker struct {
	walker tflint.ExprWalker
}

func (w *nativeWalker) Enter(node hclsyntax.Node) hcl.Diagnostics {
	if expr, ok := node.(hcl.Expression); ok {
		return w.walker.Enter(expr)
	}
	return nil
}

func (w *nativeWalker) Exit(node hclsyntax.Node) hcl.Diagnostics {
	if expr, ok := node.(hcl.Expression); ok {
		return w.walker.Exit(expr)
	}
	return nil
}

// extractJSONKeys extracts attribute names from JSON bytes using encoding/json.
// This works for both object-based JSON {"foo": ...} and array-based JSON [{"foo": ...}].
func extractJSONKeys(bytes []byte) ([]string, error) {
	// Try to unmarshal as an object first
	var obj map[string]any
	if err := stdjson.Unmarshal(bytes, &obj); err == nil {
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		return keys, nil
	}

	// Try as an array of objects
	var arr []map[string]any
	if err := stdjson.Unmarshal(bytes, &arr); err != nil {
		return nil, err
	}

	// Collect all unique keys from all objects in the array
	keysMap := make(map[string]bool)
	for _, obj := range arr {
		for k := range obj {
			keysMap[k] = true
		}
	}

	keys := make([]string, 0, len(keysMap))
	for k := range keysMap {
		keys = append(keys, k)
	}
	return keys, nil
}

// getJSONAttributes gets all attributes from a JSON body, supporting both object
// and array-based syntax. For array-based JSON like [{"import": {...}}], it
// extracts attribute names using encoding/json and builds a schema to extract them.
func getJSONAttributes(body hcl.Body, bytes []byte) (hcl.Attributes, hcl.Diagnostics) {
	// First, try JustAttributes (works for object-based JSON)
	attrs, diags := body.JustAttributes()
	if !diags.HasErrors() {
		return attrs, nil
	}

	// Extract keys using encoding/json
	keys, err := extractJSONKeys(bytes)
	if err != nil {
		return attrs, diags // Return original JustAttributes error
	}

	// Build a schema with all discovered keys
	schema := &hcl.BodySchema{
		Attributes: make([]hcl.AttributeSchema, len(keys)),
	}
	for i, key := range keys {
		schema.Attributes[i] = hcl.AttributeSchema{Name: key}
	}

	// Use PartialContent to get proper *json.expression objects
	content, _, partialDiags := body.PartialContent(schema)
	return content.Attributes, partialDiags
}
