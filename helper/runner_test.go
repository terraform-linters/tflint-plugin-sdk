package helper

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func Test_satisfyRunnerInterface(t *testing.T) {
	var runner tflint.Runner
	runner = TestRunner(t, map[string]string{})
	runner.EnsureNoError(nil, func() error { return nil })
}

func Test_WalkResourceAttributes(t *testing.T) {
	src := `
resource "aws_instance" "foo" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}

resource "aws_s3_bucket" "bar" {
  bucket = "my-tf-test-bucket"
  acl    = "private"
}`

	runner := TestRunner(t, map[string]string{"main.tf": src})

	walked := []*hcl.Attribute{}
	walker := func(attribute *hcl.Attribute) error {
		walked = append(walked, attribute)
		return nil
	}

	if err := runner.WalkResourceAttributes("aws_instance", "instance_type", walker); err != nil {
		t.Fatal(err)
	}

	expected := []*hcl.Attribute{
		{
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
	}

	opts := cmp.Options{
		cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val"),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
	}
	if !cmp.Equal(expected, walked, opts...) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, walked, opts...))
	}
}

func Test_WalkResourceBlocks(t *testing.T) {
	src := `
resource "aws_instance" "foo" {
  ami = "ami-123456"
  ebs_block_device {
    volume_size = 16
  }
}

resource "aws_s3_bucket" "bar" {
  bucket = "my-tf-test-bucket"
  acl    = "private"
}`

	runner := TestRunner(t, map[string]string{"main.tf": src})

	walked := []*hcl.Block{}
	walker := func(block *hcl.Block) error {
		walked = append(walked, block)
		return nil
	}

	if err := runner.WalkResourceBlocks("aws_instance", "ebs_block_device", walker); err != nil {
		t.Fatal(err)
	}

	expected := []*hcl.Block{
		{
			Type: "ebs_block_device",
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"volume_size": {
						Name: "volume_size",
						Expr: &hclsyntax.LiteralValueExpr{
							SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 5, Column: 19}, End: hcl.Pos{Line: 5, Column: 21}},
						},
						SrcRange:    hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 5, Column: 5}, End: hcl.Pos{Line: 5, Column: 21}},
						NameRange:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 5, Column: 5}, End: hcl.Pos{Line: 5, Column: 16}},
						EqualsRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 5, Column: 17}, End: hcl.Pos{Line: 5, Column: 18}},
					},
				},
				Blocks:   hclsyntax.Blocks{},
				SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 20}, End: hcl.Pos{Line: 6, Column: 4}},
				EndRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 6, Column: 4}, End: hcl.Pos{Line: 6, Column: 4}},
			},
			DefRange:  hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 3}, End: hcl.Pos{Line: 4, Column: 19}},
			TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 3}, End: hcl.Pos{Line: 4, Column: 19}},
		},
	}

	opts := cmp.Options{
		cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val"),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		cmpopts.IgnoreUnexported(hclsyntax.Body{}),
	}
	if !cmp.Equal(expected, walked, opts...) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, walked, opts...))
	}
}

func Test_WalkResources(t *testing.T) {
	src := `
resource "aws_instance" "foo" {
  provider = aws.west

  count = 1
  for_each = {
    foo = "bar"
  }

  instance_type = "t2.micro"
  
  connection {
    type = "ssh"
  }

  provisioner "local-exec" {
    command    = "chmod 600 ssh-key.pem"
    when       = destroy
    on_failure = continue

    connection {
      type = "ssh"
    }
  }

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = true
    ignore_changes        = all
  }
}

resource "aws_s3_bucket" "bar" {
  bucket = "my-tf-test-bucket"
  acl    = "private"
}`

	runner := TestRunner(t, map[string]string{"main.tf": src})

	walked := []*configs.Resource{}
	walker := func(resource *configs.Resource) error {
		walked = append(walked, resource)
		return nil
	}

	if err := runner.WalkResources("aws_instance", walker); err != nil {
		t.Fatal(err)
	}

	expected := []*configs.Resource{
		{
			Mode: addrs.ManagedResourceMode,
			Name: "foo",
			Type: "aws_instance",
			Config: parseBody(
				t,
				`provider = aws.west

  count = 1
  for_each = {
    foo = "bar"
  }

  instance_type = "t2.micro"

  connection {
    type = "ssh"
  }

  provisioner "local-exec" {
    command    = "chmod 600 ssh-key.pem"
    when       = destroy
    on_failure = continue

    connection {
      type = "ssh"
    }
  }

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = true
    ignore_changes        = all
  }`, "main.tf",
				hcl.Pos{Line: 3, Column: 3},
				hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 31}, End: hcl.Pos{Line: 31, Column: 2}},
				hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 31, Column: 2}, End: hcl.Pos{Line: 31, Column: 2}},
			),
			Count: parseExpression(t, `1`, "main.tf", hcl.Pos{Line: 5, Column: 11}),
			ForEach: parseExpression(t, `{
    foo = "bar"
  }`, "main.tf", hcl.Pos{Line: 6, Column: 14}),

			ProviderConfigRef: &configs.ProviderConfigRef{
				Name:       "aws",
				NameRange:  hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 14}, End: hcl.Pos{Line: 3, Column: 17}},
				Alias:      "west",
				AliasRange: &hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 17}, End: hcl.Pos{Line: 3, Column: 22}},
			},

			Managed: &configs.ManagedResource{
				Connection: &configs.Connection{
					Config: parseBody(
						t,
						`type = "ssh"`,
						"main.tf",
						hcl.Pos{Line: 13, Column: 5},
						hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 12, Column: 14}, End: hcl.Pos{Line: 14, Column: 4}},
						hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 14, Column: 4}, End: hcl.Pos{Line: 14, Column: 4}},
					),
					DeclRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 12, Column: 3}, End: hcl.Pos{Line: 12, Column: 13}},
				},
				Provisioners: []*configs.Provisioner{
					{
						Type: "local-exec",
						Config: parseBody(
							t,
							`command    = "chmod 600 ssh-key.pem"
    when       = destroy
    on_failure = continue

    connection {
      type = "ssh"
    }`,
							"main.tf",
							hcl.Pos{Line: 17, Column: 5},
							hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 16, Column: 28}, End: hcl.Pos{Line: 24, Column: 4}},
							hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 24, Column: 4}, End: hcl.Pos{Line: 24, Column: 4}},
						),
						Connection: &configs.Connection{
							Config: parseBody(
								t,
								`type = "ssh"`,
								"main.tf",
								hcl.Pos{Line: 22, Column: 7},
								hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 21, Column: 16}, End: hcl.Pos{Line: 23, Column: 6}},
								hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 23, Column: 6}, End: hcl.Pos{Line: 23, Column: 6}},
							),
							DeclRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 21, Column: 5}, End: hcl.Pos{Line: 21, Column: 15}},
						},
						When:      configs.ProvisionerWhenDestroy,
						OnFailure: configs.ProvisionerOnFailureContinue,
						DeclRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 16, Column: 3}, End: hcl.Pos{Line: 16, Column: 27}},
						TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 16, Column: 15}, End: hcl.Pos{Line: 16, Column: 27}},
					},
				},

				CreateBeforeDestroy:    true,
				PreventDestroy:         true,
				IgnoreAllChanges:       true,
				CreateBeforeDestroySet: true,
				PreventDestroySet:      true,
			},
			DeclRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 30}},
			TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 10}, End: hcl.Pos{Line: 2, Column: 24}},
		},
	}

	opts := cmp.Options{
		cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val"),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		cmpopts.IgnoreUnexported(hcl.TraverseRoot{}, hcl.TraverseAttr{}, hclsyntax.Body{}),
	}
	if !cmp.Equal(expected, walked, opts...) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, walked, opts...))
	}
}

func Test_WalkResourcesAll(t *testing.T) {
	src := `
resource "aws_instance" "foo" {
  provider = aws.west

  count = 1
  for_each = {
    foo = "bar"
  }

  instance_type = "t2.micro"
  
  connection {
    type = "ssh"
  }

  provisioner "local-exec" {
    command    = "chmod 600 ssh-key.pem"
    when       = destroy
    on_failure = continue

    connection {
      type = "ssh"
    }
  }

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = true
    ignore_changes        = all
  }
}

resource "aws_s3_bucket" "bar" {
  bucket = "my-tf-test-bucket"
  acl    = "private"
}`

	runner := TestRunner(t, map[string]string{"main.tf": src})

	walked := []*configs.Resource{}
	walker := func(resource *configs.Resource) error {
		walked = append(walked, resource)
		return nil
	}

	if err := runner.WalkResources("", walker); err != nil {
		t.Fatal(err)
	}

	expected := []*configs.Resource{
		{
			Mode: addrs.ManagedResourceMode,
			Name: "foo",
			Type: "aws_instance",
			Config: parseBody(
				t,
				`provider = aws.west

  count = 1
  for_each = {
    foo = "bar"
  }

  instance_type = "t2.micro"

  connection {
    type = "ssh"
  }

  provisioner "local-exec" {
    command    = "chmod 600 ssh-key.pem"
    when       = destroy
    on_failure = continue

    connection {
      type = "ssh"
    }
  }

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = true
    ignore_changes        = all
  }`, "main.tf",
				hcl.Pos{Line: 3, Column: 3},
				hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 31}, End: hcl.Pos{Line: 31, Column: 2}},
				hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 31, Column: 2}, End: hcl.Pos{Line: 31, Column: 2}},
			),
			Count: parseExpression(t, `1`, "main.tf", hcl.Pos{Line: 5, Column: 11}),
			ForEach: parseExpression(t, `{
    foo = "bar"
  }`, "main.tf", hcl.Pos{Line: 6, Column: 14}),

			ProviderConfigRef: &configs.ProviderConfigRef{
				Name:       "aws",
				NameRange:  hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 14}, End: hcl.Pos{Line: 3, Column: 17}},
				Alias:      "west",
				AliasRange: &hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 17}, End: hcl.Pos{Line: 3, Column: 22}},
			},

			Managed: &configs.ManagedResource{
				Connection: &configs.Connection{
					Config: parseBody(
						t,
						`type = "ssh"`,
						"main.tf",
						hcl.Pos{Line: 13, Column: 5},
						hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 12, Column: 14}, End: hcl.Pos{Line: 14, Column: 4}},
						hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 14, Column: 4}, End: hcl.Pos{Line: 14, Column: 4}},
					),
					DeclRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 12, Column: 3}, End: hcl.Pos{Line: 12, Column: 13}},
				},
				Provisioners: []*configs.Provisioner{
					{
						Type: "local-exec",
						Config: parseBody(
							t,
							`command    = "chmod 600 ssh-key.pem"
    when       = destroy
    on_failure = continue

    connection {
      type = "ssh"
    }`,
							"main.tf",
							hcl.Pos{Line: 17, Column: 5},
							hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 16, Column: 28}, End: hcl.Pos{Line: 24, Column: 4}},
							hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 24, Column: 4}, End: hcl.Pos{Line: 24, Column: 4}},
						),
						Connection: &configs.Connection{
							Config: parseBody(
								t,
								`type = "ssh"`,
								"main.tf",
								hcl.Pos{Line: 22, Column: 7},
								hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 21, Column: 16}, End: hcl.Pos{Line: 23, Column: 6}},
								hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 23, Column: 6}, End: hcl.Pos{Line: 23, Column: 6}},
							),
							DeclRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 21, Column: 5}, End: hcl.Pos{Line: 21, Column: 15}},
						},
						When:      configs.ProvisionerWhenDestroy,
						OnFailure: configs.ProvisionerOnFailureContinue,
						DeclRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 16, Column: 3}, End: hcl.Pos{Line: 16, Column: 27}},
						TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 16, Column: 15}, End: hcl.Pos{Line: 16, Column: 27}},
					},
				},

				CreateBeforeDestroy:    true,
				PreventDestroy:         true,
				IgnoreAllChanges:       true,
				CreateBeforeDestroySet: true,
				PreventDestroySet:      true,
			},
			DeclRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 30}},
			TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 10}, End: hcl.Pos{Line: 2, Column: 24}},
		},
		{
			Mode: addrs.ManagedResourceMode,
			Name: "bar",
			Type: "aws_s3_bucket",
			Config: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"acl": {
						Name: "acl",
						Expr: &hclsyntax.TemplateExpr{
							Parts: []hclsyntax.Expression{
								&hclsyntax.LiteralValueExpr{
									SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 35, Column: 13}, End: hcl.Pos{Line: 35, Column: 20}},
								},
							},
							SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 35, Column: 12}, End: hcl.Pos{Line: 35, Column: 21}},
						},
						SrcRange:    hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 35, Column: 3}, End: hcl.Pos{Line: 35, Column: 21}},
						NameRange:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 35, Column: 3}, End: hcl.Pos{Line: 35, Column: 6}},
						EqualsRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 35, Column: 10}, End: hcl.Pos{Line: 35, Column: 11}},
					},
					"bucket": {
						Name: "bucket",
						Expr: &hclsyntax.TemplateExpr{
							Parts: []hclsyntax.Expression{
								&hclsyntax.LiteralValueExpr{
									SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 34, Column: 13}, End: hcl.Pos{Line: 34, Column: 30}},
								},
							},
							SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 34, Column: 12}, End: hcl.Pos{Line: 34, Column: 31}},
						},
						SrcRange:    hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 34, Column: 3}, End: hcl.Pos{Line: 34, Column: 31}},
						NameRange:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 34, Column: 3}, End: hcl.Pos{Line: 34, Column: 9}},
						EqualsRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 34, Column: 10}, End: hcl.Pos{Line: 34, Column: 11}},
					},
				},
				Blocks:   hclsyntax.Blocks{},
				SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 33, Column: 32}, End: hcl.Pos{Line: 36, Column: 2}},
				EndRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 36, Column: 2}, End: hcl.Pos{Line: 36, Column: 2}},
			},
			Managed:   &configs.ManagedResource{},
			DeclRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 33, Column: 1}, End: hcl.Pos{Line: 33, Column: 31}},
			TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 33, Column: 10}, End: hcl.Pos{Line: 33, Column: 25}},
		},
	}

	opts := cmp.Options{
		cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val"),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		cmpopts.IgnoreUnexported(hcl.TraverseRoot{}, hcl.TraverseAttr{}, hclsyntax.Body{}),
	}
	if !cmp.Equal(expected, walked, opts...) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, walked, opts...))
	}
}

func parseBody(t *testing.T, src string, filename string, pos hcl.Pos, srcRange hcl.Range, endRange hcl.Range) hcl.Body {
	file, diags := hclsyntax.ParseConfig([]byte(src), filename, pos)
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	body := file.Body.(*hclsyntax.Body)
	body.SrcRange = srcRange
	body.EndRange = endRange

	return body
}

func parseExpression(t *testing.T, src string, filename string, pos hcl.Pos) hcl.Expression {
	expr, diags := hclsyntax.ParseExpression([]byte(src), filename, pos)
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	return expr
}

func Test_Backend(t *testing.T) {
	src := `
terraform {
  backend "s3" {
    bucket = "mybucket"
    key    = "path/to/my/key"
    region = "us-east-1"
  }
}`

	runner := TestRunner(t, map[string]string{"main.tf": src})

	backend, err := runner.Backend()
	if err != nil {
		t.Fatal(err)
	}

	expected := &configs.Backend{
		Type:      "s3",
		TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 11}, End: hcl.Pos{Line: 3, Column: 15}},
		Config: parseBody(
			t,
			`bucket = "mybucket"
    key    = "path/to/my/key"
    region = "us-east-1"`,
			"main.tf",
			hcl.Pos{Line: 4, Column: 5},
			hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 16}, End: hcl.Pos{Line: 7, Column: 4}},
			hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 7, Column: 4}, End: hcl.Pos{Line: 7, Column: 4}},
		),
		DeclRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 3}, End: hcl.Pos{Line: 3, Column: 15}},
	}

	opts := cmp.Options{
		cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val"),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		cmpopts.IgnoreUnexported(hclsyntax.Body{}),
	}
	if !cmp.Equal(expected, backend, opts...) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, backend, opts...))
	}
}

func Test_EvaluateExpr(t *testing.T) {
	src := `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`

	runner := TestRunner(t, map[string]string{"main.tf": src})

	err := runner.WalkResourceAttributes("aws_instance", "instance_type", func(attribute *hcl.Attribute) error {
		var instanceType string
		if err := runner.EvaluateExpr(attribute.Expr, &instanceType, nil); err != nil {
			t.Fatal(err)
		}

		if instanceType != "t2.micro" {
			t.Fatalf(`expected value is "t2.micro", but got "%s"`, instanceType)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

type dummyRule struct{}

func (r *dummyRule) Name() string              { return "dummy_rule" }
func (r *dummyRule) Enabled() bool             { return true }
func (r *dummyRule) Severity() string          { return tflint.ERROR }
func (r *dummyRule) Link() string              { return "" }
func (r *dummyRule) Check(tflint.Runner) error { return nil }

func Test_EmitIssueOnExpr(t *testing.T) {
	src := `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`

	runner := TestRunner(t, map[string]string{"main.tf": src})

	err := runner.WalkResourceAttributes("aws_instance", "instance_type", func(attribute *hcl.Attribute) error {
		if err := runner.EmitIssueOnExpr(&dummyRule{}, "issue found", attribute.Expr); err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := Issues{
		{
			Rule:    &dummyRule{},
			Message: "issue found",
			Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 19}, End: hcl.Pos{Line: 3, Column: 29}},
		},
	}

	opt := cmpopts.IgnoreFields(hcl.Pos{}, "Byte")
	if !cmp.Equal(expected, runner.Issues, opt) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, runner.Issues, opt))
	}
}

func Test_EmitIssue(t *testing.T) {
	src := `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`

	runner := TestRunner(t, map[string]string{"main.tf": src})

	err := runner.WalkResourceAttributes("aws_instance", "instance_type", func(attribute *hcl.Attribute) error {
		if err := runner.EmitIssue(&dummyRule{}, "issue found", attribute.Expr.Range()); err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := Issues{
		{
			Rule:    &dummyRule{},
			Message: "issue found",
			Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 19}, End: hcl.Pos{Line: 3, Column: 29}},
		},
	}

	opt := cmpopts.IgnoreFields(hcl.Pos{}, "Byte")
	if !cmp.Equal(expected, runner.Issues, opt) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, runner.Issues, opt))
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

func Test_Files(t *testing.T) {
	var sources = map[string]string{
		"main.tf": `
			resource "aws_instance" "foo" {
				instance_type = "t2.micro"
			}`,
		"outputs.tf": `
			output "dummy" {
				value = "test"
			}`,
		"providers.tf": `
			provider "aws" {
				region = "us-east-1"
			}`,
	}

	runner := TestRunner(t, sources)

	files, err := runner.Files()
	if err != nil {
		t.Fatalf("The response has an unexpected error: %s", err)
	}

	if !cmp.Equal(len(sources), len(files)) {
		t.Fatalf("Sources and Files differ: %s", cmp.Diff(sources, files))
	}
}
