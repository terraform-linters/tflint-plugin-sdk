package helper

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func Test_GetResourceContent(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Resource string
		Schema   *hclext.BodySchema
		Expected *hclext.BodyContent
	}{
		{
			Name: "attribute",
			Src: `
resource "aws_instance" "foo" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}

resource "aws_s3_bucket" "bar" {
  bucket = "my-tf-test-bucket"
  acl    = "private"
}`,
			Resource: "aws_instance",
			Schema: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
			},
			Expected: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "foo"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{
								"instance_type": {
									Name: "instance_type",
									Expr: &hclsyntax.TemplateExpr{
										Parts: []hclsyntax.Expression{
											&hclsyntax.LiteralValueExpr{
												SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 20}, End: hcl.Pos{Line: 4, Column: 28}},
											},
										},
										SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 19}, End: hcl.Pos{Line: 4, Column: 29}},
									},
									Range:     hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 3}, End: hcl.Pos{Line: 4, Column: 29}},
									NameRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 3}, End: hcl.Pos{Line: 4, Column: 16}},
								},
							},
							Blocks: hclext.Blocks{},
						},
						DefRange:  hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 30}},
						TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 9}},
						LabelRanges: []hcl.Range{
							{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 10}, End: hcl.Pos{Line: 2, Column: 24}},
							{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 25}, End: hcl.Pos{Line: 2, Column: 30}},
						},
					},
				},
			},
		},
		{
			Name: "block",
			Src: `
resource "aws_instance" "foo" {
  ami = "ami-123456"
  ebs_block_device {
    volume_size = 16
  }
}

resource "aws_s3_bucket" "bar" {
  bucket = "my-tf-test-bucket"
  acl    = "private"
}`,
			Resource: "aws_instance",
			Schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "ebs_block_device", Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "volume_size"}}}},
				},
			},
			Expected: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "foo"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{
											"volume_size": {
												Name: "volume_size",
												Expr: &hclsyntax.LiteralValueExpr{
													SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 5, Column: 19}, End: hcl.Pos{Line: 5, Column: 21}},
												},
												Range:     hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 5, Column: 5}, End: hcl.Pos{Line: 5, Column: 21}},
												NameRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 5, Column: 5}, End: hcl.Pos{Line: 5, Column: 16}},
											},
										},
										Blocks: hclext.Blocks{},
									},
									DefRange:  hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 3}, End: hcl.Pos{Line: 4, Column: 19}},
									TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 3}, End: hcl.Pos{Line: 4, Column: 19}},
								},
							},
						},
						DefRange:  hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 30}},
						TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 9}},
						LabelRanges: []hcl.Range{
							{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 10}, End: hcl.Pos{Line: 2, Column: 24}},
							{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 25}, End: hcl.Pos{Line: 2, Column: 30}},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := TestRunner(t, map[string]string{"main.tf": tc.Src})

			got, err := runner.GetResourceContent(tc.Resource, tc.Schema, nil)
			if err != nil {
				t.Error(err)
			} else {
				opts := cmp.Options{
					cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val"),
					cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
				}
				if diff := cmp.Diff(tc.Expected, got, opts...); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func Test_GetModuleContent(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Schema   *hclext.BodySchema
		Expected *hclext.BodyContent
	}{
		{
			Name: "backend",
			Src: `
terraform {
	backend "s3" {
	bucket = "mybucket"
	key    = "path/to/my/key"
	region = "us-east-1"
	}
}`,
			Schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type: "terraform",
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type:       "backend",
									LabelNames: []string{"name"},
									Body: &hclext.BodySchema{
										Attributes: []hclext.AttributeSchema{{Name: "bucket"}},
									},
								},
							},
						},
					},
				},
			},
			Expected: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type: "terraform",
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type:   "backend",
									Labels: []string{"s3"},
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{
											"bucket": &hclext.Attribute{
												Name: "bucket",
												Expr: &hclsyntax.TemplateExpr{
													Parts: []hclsyntax.Expression{
														&hclsyntax.LiteralValueExpr{
															SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 12}, End: hcl.Pos{Line: 4, Column: 20}},
														},
													},
													SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 11}, End: hcl.Pos{Line: 4, Column: 21}},
												},
												Range:     hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 2}, End: hcl.Pos{Line: 4, Column: 21}},
												NameRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 2}, End: hcl.Pos{Line: 4, Column: 8}},
											},
										},
										Blocks: hclext.Blocks{},
									},
									DefRange:  hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 2}, End: hcl.Pos{Line: 3, Column: 14}},
									TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 2}, End: hcl.Pos{Line: 3, Column: 9}},
									LabelRanges: []hcl.Range{
										{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 14}},
									},
								},
							},
						},
						DefRange:  hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 10}},
						TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 10}},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := TestRunner(t, map[string]string{"main.tf": tc.Src})

			got, err := runner.GetModuleContent(tc.Schema, nil)
			if err != nil {
				t.Error(err)
			} else {
				opts := cmp.Options{
					cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val"),
					cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
				}
				if diff := cmp.Diff(tc.Expected, got, opts...); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func Test_EvaluateExpr(t *testing.T) {
	tests := []struct {
		Name string
		Src  string
		Want string
	}{
		{
			Name: "string literal",
			Src: `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`,
			Want: "t2.micro",
		},
		{
			Name: "string interpolation",
			Src: `
variable "instance_type" {
	type = string
  default = "t2.micro"
}

resource "aws_instance" "foo" {
  instance_type = var.instance_type
}`,
			Want: "t2.micro",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner := TestRunner(t, map[string]string{"main.tf": test.Src})

			resources, err := runner.GetResourceContent("aws_instance", &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
			}, nil)
			if err != nil {
				t.Fatal(err)
			}

			for _, resource := range resources.Blocks {
				var instanceType string
				if err := runner.EvaluateExpr(resource.Body.Attributes["instance_type"].Expr, &instanceType, nil); err != nil {
					t.Fatal(err)
				}

				if instanceType != test.Want {
					t.Fatalf(`"%s" is expected, but got "%s"`, test.Want, instanceType)
				}
			}
		})
	}
}

type dummyRule struct {
	tflint.DefaultRule
}

func (r *dummyRule) Name() string              { return "dummy_rule" }
func (r *dummyRule) Enabled() bool             { return true }
func (r *dummyRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *dummyRule) Check(tflint.Runner) error { return nil }

func Test_EmitIssue(t *testing.T) {
	src := `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`

	runner := TestRunner(t, map[string]string{"main.tf": src})

	resources, err := runner.GetResourceContent("aws_instance", &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, resource := range resources.Blocks {
		if err := runner.EmitIssue(&dummyRule{}, "issue found", resource.Body.Attributes["instance_type"].Expr.Range()); err != nil {
			t.Fatal(err)
		}
	}

	expected := Issues{
		{
			Rule:    &dummyRule{},
			Message: "issue found",
			Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 19}, End: hcl.Pos{Line: 3, Column: 29}},
		},
	}

	opt := cmpopts.IgnoreFields(hcl.Pos{}, "Byte")
	if diff := cmp.Diff(expected, runner.Issues, opt); diff != "" {
		t.Fatal(diff)
	}
}

func Test_EnsureNoError(t *testing.T) {
	runner := TestRunner(t, map[string]string{})

	var run bool
	err := runner.EnsureNoError(nil, func() error {
		run = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if !run {
		t.Fatal("Expected to exec the passed proc, but doesn't")
	}
}
