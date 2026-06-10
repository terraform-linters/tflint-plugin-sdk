package runner

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type fakeGetter struct {
	schema  *hclext.BodySchema
	opts    *tflint.GetModuleContentOption
	content *hclext.BodyContent
	err     error
}

func (g *fakeGetter) GetModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	g.schema = schema
	g.opts = opts
	return g.content, g.err
}

func TestGetResourceContent(t *testing.T) {
	inner := &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "instance_type"}}}
	getter := &fakeGetter{
		content: &hclext.BodyContent{
			Blocks: hclext.Blocks{
				{Type: "resource", Labels: []string{"aws_instance", "foo"}},
				{Type: "resource", Labels: []string{"aws_s3_bucket", "bar"}},
				{Type: "resource", Labels: []string{"aws_instance", "baz"}},
			},
		},
	}

	got, err := GetResourceContent(getter, "aws_instance", inner, nil)
	if err != nil {
		t.Fatal(err)
	}

	want := &hclext.BodyContent{
		Blocks: hclext.Blocks{
			{Type: "resource", Labels: []string{"aws_instance", "foo"}},
			{Type: "resource", Labels: []string{"aws_instance", "baz"}},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}

	wantSchema := &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}, Body: inner},
		},
	}
	if diff := cmp.Diff(wantSchema, getter.schema); diff != "" {
		t.Fatal(diff)
	}

	if getter.opts.Hint.ResourceType != "aws_instance" {
		t.Fatalf(`expected hint to be "aws_instance", but got %q`, getter.opts.Hint.ResourceType)
	}
}

func TestGetProviderContent(t *testing.T) {
	inner := &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "region"}}}
	getter := &fakeGetter{
		content: &hclext.BodyContent{
			Blocks: hclext.Blocks{
				{Type: "provider", Labels: []string{"aws"}},
				{Type: "provider", Labels: []string{"google"}},
			},
		},
	}

	got, err := GetProviderContent(getter, "aws", inner, nil)
	if err != nil {
		t.Fatal(err)
	}

	want := &hclext.BodyContent{
		Blocks: hclext.Blocks{
			{Type: "provider", Labels: []string{"aws"}},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}

	wantSchema := &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{Type: "provider", LabelNames: []string{"name"}, Body: inner},
		},
	}
	if diff := cmp.Diff(wantSchema, getter.schema); diff != "" {
		t.Fatal(diff)
	}
}

func TestGetResourceContent_error(t *testing.T) {
	want := errors.New("unexpected")
	getter := &fakeGetter{err: want}

	if _, err := GetResourceContent(getter, "aws_instance", &hclext.BodySchema{}, nil); !errors.Is(err, want) {
		t.Fatalf("expected %v, but got %v", want, err)
	}
}
