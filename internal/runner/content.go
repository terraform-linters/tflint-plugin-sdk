package runner

import (
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// ModuleContentGetter fetches module content for the content shorthands.
// Runner implementations satisfy it with their GetModuleContent method.
type ModuleContentGetter interface {
	GetModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error)
}

// GetResourceContent gets the contents of resources of the given type based
// on the schema. This is shorthand of GetModuleContent for resources.
func GetResourceContent(g ModuleContentGetter, name string, inner *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if opts == nil {
		opts = &tflint.GetModuleContentOption{}
	}
	opts.Hint.ResourceType = name

	return getBlockContent(g, "resource", []string{"type", "name"}, name, inner, opts)
}

// GetProviderContent gets the contents of providers with the given name based
// on the schema. This is shorthand of GetModuleContent for providers.
func GetProviderContent(g ModuleContentGetter, name string, inner *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if opts == nil {
		opts = &tflint.GetModuleContentOption{}
	}

	return getBlockContent(g, "provider", []string{"name"}, name, inner, opts)
}

func getBlockContent(g ModuleContentGetter, blockType string, labelNames []string, name string, inner *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	body, err := g.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{Type: blockType, LabelNames: labelNames, Body: inner},
		},
	}, opts)
	if err != nil {
		return nil, err
	}

	content := &hclext.BodyContent{Blocks: []*hclext.Block{}}
	for _, block := range body.Blocks {
		if block.Labels[0] != name {
			continue
		}

		content.Blocks = append(content.Blocks, block)
	}

	return content, nil
}
