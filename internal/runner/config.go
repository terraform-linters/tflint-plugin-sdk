package runner

import (
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

// FetchRuleConfigFunc returns the rule config content matching the schema.
// A nil content means the rule is not configured.
type FetchRuleConfigFunc func(schema *hclext.BodySchema) (*hclext.BodyContent, error)

// DecodeRuleConfig infers the schema from ret, fetches the matching rule
// config content, and decodes it into ret. Missing or empty content leaves
// ret unchanged.
func DecodeRuleConfig(ret any, fetch FetchRuleConfigFunc) error {
	content, err := fetch(hclext.ImpliedBodySchema(ret))
	if err != nil {
		return err
	}
	if content == nil || content.IsEmpty() {
		return nil
	}

	if diags := hclext.DecodeBody(content, nil, ret); diags.HasErrors() {
		return diags
	}
	return nil
}
