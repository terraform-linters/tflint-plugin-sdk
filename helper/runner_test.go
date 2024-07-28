package helper

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
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

func Test_GetModuleContent_json(t *testing.T) {
	files := map[string]string{
		"main.tf.json": `{"variable": {"foo": {"type": "string"}}}`,
	}

	runner := TestRunner(t, files)

	schema := &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "variable",
				Body: &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "type",
							LabelNames: []string{"name"},
							Body:       &hclext.BodySchema{},
						},
					},
				},
			},
		},
	}
	got, err := runner.GetModuleContent(schema, nil)
	if err != nil {
		t.Error(err)
	} else {
		if len(got.Blocks) != 1 {
			t.Errorf("got %d blocks, but 1 block is expected", len(got.Blocks))
		}
	}
}

func TestWalkExpressions(t *testing.T) {
	tests := []struct {
		name   string
		files  map[string]string
		walked []hcl.Range
	}{
		{
			name: "resource",
			files: map[string]string{
				"resource.tf": `
resource "null_resource" "test" {
  key = "foo"
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			name: "data source",
			files: map[string]string{
				"data.tf": `
data "null_dataresource" "test" {
  key = "foo"
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			name: "module call",
			files: map[string]string{
				"module.tf": `
module "m" {
  source = "./module"
  key    = "foo"
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 12}, End: hcl.Pos{Line: 3, Column: 22}},
				{Start: hcl.Pos{Line: 3, Column: 13}, End: hcl.Pos{Line: 3, Column: 21}},
				{Start: hcl.Pos{Line: 4, Column: 12}, End: hcl.Pos{Line: 4, Column: 17}},
				{Start: hcl.Pos{Line: 4, Column: 13}, End: hcl.Pos{Line: 4, Column: 16}},
			},
		},
		{
			name: "provider config",
			files: map[string]string{
				"provider.tf": `
provider "p" {
  key = "foo"
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			name: "locals",
			files: map[string]string{
				"locals.tf": `
locals {
  key = "foo"
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			name: "output",
			files: map[string]string{
				"output.tf": `
output "o" {
  value = "foo"
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 11}, End: hcl.Pos{Line: 3, Column: 16}},
				{Start: hcl.Pos{Line: 3, Column: 12}, End: hcl.Pos{Line: 3, Column: 15}},
			},
		},
		{
			name: "resource with block",
			files: map[string]string{
				"resource.tf": `
resource "null_resource" "test" {
  key = "foo"

  lifecycle {
    ignore_changes = [key]
  }
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
				{Start: hcl.Pos{Line: 6, Column: 22}, End: hcl.Pos{Line: 6, Column: 27}},
				{Start: hcl.Pos{Line: 6, Column: 23}, End: hcl.Pos{Line: 6, Column: 26}},
			},
		},
		{
			name: "resource json",
			files: map[string]string{
				"resource.tf.json": `
{
  "resource": {
    "null_resource": {
      "test": {
        "key": "foo",
        "nested": {
          "key": "foo"
        },
        "list": [{
          "key": "foo"
        }]
      }
    }
  }
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 15}, End: hcl.Pos{Line: 15, Column: 4}},
			},
		},
		{
			name: "multiple files",
			files: map[string]string{
				"main.tf": `
provider "aws" {
  region = "us-east-1"

  assume_role {
    role_arn = "arn:aws:iam::123412341234:role/ExampleRole"
  }
}`,
				"main_override.tf": `
provider "aws" {
  region = "us-east-1"

  assume_role {
    role_arn = null
  }
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 12}, End: hcl.Pos{Line: 3, Column: 23}, Filename: "main.tf"},
				{Start: hcl.Pos{Line: 3, Column: 13}, End: hcl.Pos{Line: 3, Column: 22}, Filename: "main.tf"},
				{Start: hcl.Pos{Line: 6, Column: 16}, End: hcl.Pos{Line: 6, Column: 60}, Filename: "main.tf"},
				{Start: hcl.Pos{Line: 6, Column: 17}, End: hcl.Pos{Line: 6, Column: 59}, Filename: "main.tf"},
				{Start: hcl.Pos{Line: 3, Column: 12}, End: hcl.Pos{Line: 3, Column: 23}, Filename: "main_override.tf"},
				{Start: hcl.Pos{Line: 3, Column: 13}, End: hcl.Pos{Line: 3, Column: 22}, Filename: "main_override.tf"},
				{Start: hcl.Pos{Line: 6, Column: 16}, End: hcl.Pos{Line: 6, Column: 20}, Filename: "main_override.tf"},
			},
		},
		{
			name: "nested attributes",
			files: map[string]string{
				"data.tf": `
data "terraform_remote_state" "remote_state" {
  backend = "remote"

  config = {
    organization = "Organization"
    workspaces = {
      name = "${var.environment}"
    }
  }
}`,
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 13}, End: hcl.Pos{Line: 3, Column: 21}},
				{Start: hcl.Pos{Line: 3, Column: 14}, End: hcl.Pos{Line: 3, Column: 20}},
				{Start: hcl.Pos{Line: 5, Column: 12}, End: hcl.Pos{Line: 10, Column: 4}},
				{Start: hcl.Pos{Line: 6, Column: 5}, End: hcl.Pos{Line: 6, Column: 17}},
				{Start: hcl.Pos{Line: 6, Column: 20}, End: hcl.Pos{Line: 6, Column: 34}},
				{Start: hcl.Pos{Line: 6, Column: 21}, End: hcl.Pos{Line: 6, Column: 33}},
				{Start: hcl.Pos{Line: 7, Column: 5}, End: hcl.Pos{Line: 7, Column: 15}},
				{Start: hcl.Pos{Line: 7, Column: 18}, End: hcl.Pos{Line: 9, Column: 6}},
				{Start: hcl.Pos{Line: 8, Column: 7}, End: hcl.Pos{Line: 8, Column: 11}},
				{Start: hcl.Pos{Line: 8, Column: 14}, End: hcl.Pos{Line: 8, Column: 34}},
				{Start: hcl.Pos{Line: 8, Column: 17}, End: hcl.Pos{Line: 8, Column: 32}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := TestRunner(t, test.files)

			walked := []hcl.Range{}
			diags := runner.WalkExpressions(tflint.ExprWalkFunc(func(expr hcl.Expression) hcl.Diagnostics {
				walked = append(walked, expr.Range())
				return nil
			}))
			if diags.HasErrors() {
				t.Fatal(diags)
			}
			opts := cmp.Options{
				cmpopts.IgnoreFields(hcl.Range{}, "Filename"),
				cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
				cmpopts.SortSlices(func(x, y hcl.Range) bool { return x.String() > y.String() }),
			}
			if diff := cmp.Diff(walked, test.walked, opts); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_DecodeRuleConfig(t *testing.T) {
	files := map[string]string{
		".tflint.hcl": `
rule "test" {
  enabled = true
  foo     = "bar"
}`,
	}

	runner := TestRunner(t, files)

	type ruleConfig struct {
		Foo string `hclext:"foo"`
	}
	target := &ruleConfig{}
	if err := runner.DecodeRuleConfig("test", target); err != nil {
		t.Fatal(err)
	}

	if target.Foo != "bar" {
		t.Errorf("target.Foo should be `bar`, but got `%s`", target.Foo)
	}
}

func Test_DecodeRuleConfig_config_not_found(t *testing.T) {
	runner := TestRunner(t, map[string]string{})

	type ruleConfig struct {
		Foo string `hclext:"foo"`
	}
	target := &ruleConfig{}
	if err := runner.DecodeRuleConfig("test", target); err != nil {
		t.Fatal(err)
	}

	if target.Foo != "" {
		t.Errorf("target.Foo should be empty, but got `%s`", target.Foo)
	}
}

func Test_EvaluateExpr_string(t *testing.T) {
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
				// raw value
				var instanceType string
				if err := runner.EvaluateExpr(resource.Body.Attributes["instance_type"].Expr, &instanceType, nil); err != nil {
					t.Fatal(err)
				}

				if instanceType != test.Want {
					t.Fatalf(`"%s" is expected, but got "%s"`, test.Want, instanceType)
				}

				// callback
				if err := runner.EvaluateExpr(resource.Body.Attributes["instance_type"].Expr, func(val string) error {
					if instanceType != test.Want {
						t.Fatalf(`"%s" is expected, but got "%s"`, test.Want, instanceType)
					}
					return nil
				}, nil); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func Test_EvaluateExpr_value(t *testing.T) {
	tests := []struct {
		Name string
		Src  string
		Want string
	}{
		{
			Name: "sensitive variable",
			Src: `
variable "instance_type" {
  type = string
  default = "secret"
  sensitive = true
}

resource "aws_instance" "foo" {
  instance_type = var.instance_type
}`,
			Want: `cty.StringVal("secret").Mark(marks.Sensitive)`,
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
				// raw value
				var instanceType cty.Value
				if err := runner.EvaluateExpr(resource.Body.Attributes["instance_type"].Expr, &instanceType, nil); err != nil {
					t.Fatal(err)
				}

				if instanceType.GoString() != test.Want {
					t.Fatalf(`"%s" is expected, but got "%s"`, test.Want, instanceType.GoString())
				}

				// callback
				if err := runner.EvaluateExpr(resource.Body.Attributes["instance_type"].Expr, func(val cty.Value) error {
					if instanceType.GoString() != test.Want {
						t.Fatalf(`"%s" is expected, but got "%s"`, test.Want, instanceType.GoString())
					}
					return nil
				}, nil); err != nil {
					t.Fatal(err)
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

func Test_EmitIssueWithFix(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		name     string
		src      string
		rng      hcl.Range
		fix      func(tflint.Fixer) error
		want     Issues
		fixed    string
		errCheck func(error) bool
	}{
		{
			name: "with fix",
			src: `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`,
			rng: hcl.Range{
				Filename: "main.tf",
				Start:    hcl.Pos{Line: 3, Column: 19, Byte: 51},
				End:      hcl.Pos{Line: 3, Column: 29, Byte: 61},
			},
			fix: func(fixer tflint.Fixer) error {
				return fixer.ReplaceText(
					hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 19, Byte: 51},
						End:      hcl.Pos{Line: 3, Column: 29, Byte: 61},
					},
					`"t3.micro"`,
				)
			},
			want: Issues{
				{
					Rule:    &dummyRule{},
					Message: "issue found",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 19}, End: hcl.Pos{Line: 3, Column: 29}},
				},
			},
			fixed: `
resource "aws_instance" "foo" {
  instance_type = "t3.micro"
}`,
			errCheck: neverHappend,
		},
		{
			name: "autofix is not supported",
			src: `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`,
			rng: hcl.Range{
				Filename: "main.tf",
				Start:    hcl.Pos{Line: 3, Column: 19, Byte: 51},
				End:      hcl.Pos{Line: 3, Column: 29, Byte: 61},
			},
			fix: func(fixer tflint.Fixer) error {
				if err := fixer.ReplaceText(
					hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 19, Byte: 51},
						End:      hcl.Pos{Line: 3, Column: 29, Byte: 61},
					},
					`"t3.micro"`,
				); err != nil {
					return err
				}
				return tflint.ErrFixNotSupported
			},
			want: Issues{
				{
					Rule:    &dummyRule{},
					Message: "issue found",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 19}, End: hcl.Pos{Line: 3, Column: 29}},
				},
			},
			errCheck: neverHappend,
		},
		{
			name: "other errors",
			src: `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`,
			rng: hcl.Range{
				Filename: "main.tf",
				Start:    hcl.Pos{Line: 3, Column: 19, Byte: 51},
				End:      hcl.Pos{Line: 3, Column: 29, Byte: 61},
			},
			fix: func(fixer tflint.Fixer) error {
				if err := fixer.ReplaceText(
					hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 19, Byte: 51},
						End:      hcl.Pos{Line: 3, Column: 29, Byte: 61},
					},
					`"t3.micro"`,
				); err != nil {
					return err
				}
				return errors.New("unexpected error")
			},
			want: Issues{},
			fixed: `
resource "aws_instance" "foo" {
  instance_type = "t3.micro"
}`,
			errCheck: func(err error) bool {
				return err == nil && err.Error() != "unexpected error"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := TestRunner(t, map[string]string{"main.tf": test.src})

			err := runner.EmitIssueWithFix(&dummyRule{}, "issue found", test.rng, test.fix)
			if test.errCheck(err) {
				t.Fatal(err)
			}

			opt := cmpopts.IgnoreFields(hcl.Pos{}, "Byte")
			if diff := cmp.Diff(test.want, runner.Issues, opt); diff != "" {
				t.Fatal(diff)
			}
			if diff := cmp.Diff(test.fixed, string(runner.Changes()["main.tf"]), opt); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestChanges(t *testing.T) {
	tests := []struct {
		name string
		src  string
		fix  func(tflint.Fixer) error
		want string
	}{
		{
			name: "changes",
			src: `
locals {
  foo = "bar"
}`,
			fix: func(fixer tflint.Fixer) error {
				return fixer.InsertTextBefore(
					hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Byte: 12},
						End:      hcl.Pos{Byte: 15},
					},
					"bar = \"baz\"\n",
				)
			},
			want: `
locals {
  bar = "baz"
  foo = "bar"
}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := TestRunner(t, map[string]string{"main.tf": test.src})

			if err := test.fix(runner.fixer); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.want, string(runner.Changes()["main.tf"])); diff != "" {
				t.Fatal(diff)
			}
		})
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
