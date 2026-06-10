package runner

import (
	"errors"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

type ruleConfig struct {
	Format string `hclext:"format,optional"`
}

func TestDecodeRuleConfig(t *testing.T) {
	for _, tc := range []struct {
		name  string
		fetch FetchRuleConfigFunc
		want  string
	}{
		{
			name: "config is decoded",
			fetch: func(schema *hclext.BodySchema) (*hclext.BodyContent, error) {
				if len(schema.Attributes) != 1 || schema.Attributes[0].Name != "format" {
					t.Fatalf("expected implied schema with a format attribute, but got %#v", schema)
				}

				file, diags := hclsyntax.ParseConfig([]byte(`format = "snake_case"`), "config.tf", hcl.InitialPos)
				if diags.HasErrors() {
					return nil, diags
				}
				content, diags := hclext.Content(file.Body, schema)
				if diags.HasErrors() {
					return nil, diags
				}
				return content, nil
			},
			want: "snake_case",
		},
		{
			name: "rule is not configured",
			fetch: func(schema *hclext.BodySchema) (*hclext.BodyContent, error) {
				return nil, nil
			},
			want: "",
		},
		{
			name: "config is empty",
			fetch: func(schema *hclext.BodySchema) (*hclext.BodyContent, error) {
				return &hclext.BodyContent{}, nil
			},
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ret := &ruleConfig{}
			if err := DecodeRuleConfig(ret, tc.fetch); err != nil {
				t.Fatal(err)
			}
			if ret.Format != tc.want {
				t.Fatalf("expected %q, but got %q", tc.want, ret.Format)
			}
		})
	}
}

func TestDecodeRuleConfig_error(t *testing.T) {
	want := errors.New("unexpected")
	err := DecodeRuleConfig(&ruleConfig{}, func(schema *hclext.BodySchema) (*hclext.BodyContent, error) {
		return nil, want
	})
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, but got %v", want, err)
	}
}
