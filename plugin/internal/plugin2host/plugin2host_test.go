package plugin2host

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/internal"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/internal/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"google.golang.org/grpc"
)

func startTestGRPCServer(t *testing.T, runner Server) *GRPCClient {
	conn, _ := plugin.TestGRPCConn(t, func(server *grpc.Server) {
		proto.RegisterRunnerServer(server, &GRPCServer{Impl: runner})
	})

	return &GRPCClient{
		Client:     proto.NewRunnerClient(conn),
		Fixer:      internal.NewFixer(runner.GetFiles(tflint.RootModuleCtxType)),
		FixEnabled: false,
	}
}

var _ Server = &mockServer{}

type mockServer struct {
	impl mockServerImpl
}

type mockServerImpl struct {
	getOriginalwd        func() string
	getModulePath        func() []string
	getModuleContent     func(*hclext.BodySchema, tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics)
	getFile              func(string) (*hcl.File, error)
	getFiles             func() map[string][]byte
	getRuleConfigContent func(string, *hclext.BodySchema) (*hclext.BodyContent, map[string][]byte, error)
	evaluateExpr         func(hcl.Expression, tflint.EvaluateExprOption) (cty.Value, error)
	emitIssue            func(tflint.Rule, string, hcl.Range, bool) (bool, error)
	applyChanges         func(map[string][]byte) error
}

func newMockServer(impl mockServerImpl) *mockServer {
	return &mockServer{impl: impl}
}

func (s *mockServer) GetOriginalwd() string {
	if s.impl.getOriginalwd != nil {
		return s.impl.getOriginalwd()
	}
	return ""
}

func (s *mockServer) GetModulePath() []string {
	if s.impl.getModulePath != nil {
		return s.impl.getModulePath()
	}
	return []string{}
}

func (s *mockServer) GetModuleContent(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
	if s.impl.getModuleContent != nil {
		return s.impl.getModuleContent(schema, opts)
	}
	return &hclext.BodyContent{}, hcl.Diagnostics{}
}

func (s *mockServer) GetFile(filename string) (*hcl.File, error) {
	if s.impl.getFile != nil {
		return s.impl.getFile(filename)
	}
	return nil, nil
}

func (s *mockServer) GetFiles(tflint.ModuleCtxType) map[string][]byte {
	if s.impl.getFiles != nil {
		return s.impl.getFiles()
	}
	return map[string][]byte{}
}

func (s *mockServer) GetRuleConfigContent(name string, schema *hclext.BodySchema) (*hclext.BodyContent, map[string][]byte, error) {
	if s.impl.getRuleConfigContent != nil {
		return s.impl.getRuleConfigContent(name, schema)
	}
	return &hclext.BodyContent{}, map[string][]byte{}, nil
}

func (s *mockServer) EvaluateExpr(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
	if s.impl.evaluateExpr != nil {
		return s.impl.evaluateExpr(expr, opts)
	}
	return cty.Value{}, nil
}

func (s *mockServer) EmitIssue(rule tflint.Rule, message string, location hcl.Range, fixable bool) (bool, error) {
	if s.impl.emitIssue != nil {
		return s.impl.emitIssue(rule, message, location, fixable)
	}
	return true, nil
}

func (s *mockServer) ApplyChanges(sources map[string][]byte) error {
	if s.impl.applyChanges != nil {
		return s.impl.applyChanges(sources)
	}
	return nil
}

// @see https://github.com/google/go-cmp/issues/40
var allowAllUnexported = cmp.Exporter(func(reflect.Type) bool { return true })

func TestGetOriginalwd(t *testing.T) {
	tests := []struct {
		Name       string
		ServerImpl func() string
		Want       string
	}{
		{
			Name: "get the original working directory",
			ServerImpl: func() string {
				return "/work"
			},
			Want: "/work",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{getOriginalwd: test.ServerImpl}))

			got, err := client.GetOriginalwd()
			if err != nil {
				t.Fatalf("failed to call GetOriginalwd: %s", err)
			}
			if diff := cmp.Diff(got, test.Want); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestGetModulePath(t *testing.T) {
	tests := []struct {
		Name       string
		ServerImpl func() []string
		Want       addrs.Module
	}{
		{
			Name: "get root module path",
			ServerImpl: func() []string {
				return []string{}
			},
			Want: nil,
		},
		{
			Name: "get child module path",
			ServerImpl: func() []string {
				return []string{"child1", "child2"}
			},
			Want: []string{"child1", "child2"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{getModulePath: test.ServerImpl}))

			got, err := client.GetModulePath()
			if err != nil {
				t.Fatalf("failed to call GetModulePath: %s", err)
			}
			if diff := cmp.Diff(got, test.Want); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestGetResourceContent(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	// default getFileImpl function
	files := map[string][]byte{}
	fileExists := func() map[string][]byte {
		return files
	}

	// test util functions
	hclFile := func(filename string, code string) *hcl.File {
		file, diags := hclsyntax.ParseConfig([]byte(code), filename, hcl.InitialPos)
		if diags.HasErrors() {
			panic(diags)
		}
		files[filename] = file.Bytes
		return file
	}
	jsonFile := func(filename string, code string) *hcl.File {
		file, diags := json.Parse([]byte(code), filename)
		if diags.HasErrors() {
			panic(diags)
		}
		files[filename] = file.Bytes
		return file
	}

	tests := []struct {
		Name       string
		Args       func() (string, *hclext.BodySchema, *tflint.GetModuleContentOption)
		ServerImpl func(*hclext.BodySchema, tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics)
		Want       func(string, *hclext.BodySchema, *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics)
		ErrCheck   func(error) bool
	}{
		{
			Name: "get HCL content",
			Args: func() (string, *hclext.BodySchema, *tflint.GetModuleContentOption) {
				return "aws_instance", &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
				}, nil
			},
			ServerImpl: func(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				file := hclFile("test.tf", `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}

resource "aws_s3_bucket" "bar" {
	bucket = "test"
}`)
				return hclext.PartialContent(file.Body, schema)
			},
			Want: func(resource string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				// Removed "aws_s3_bucket" resource
				file := hclFile("test.tf", `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`)
				return hclext.Content(file.Body, &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
							},
						},
					},
				})
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "get JSON content",
			Args: func() (string, *hclext.BodySchema, *tflint.GetModuleContentOption) {
				return "aws_instance", &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
				}, nil
			},
			ServerImpl: func(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				file := jsonFile("test.tf.json", `
{
  "resource": {
    "aws_instance": {
      "foo": {
        "instance_type": "t2.micro"
      }
    },
	"aws_s3_bucket": {
      "bar": {
        "bucket": "test"
	  }
	}
  }
}`)
				return hclext.PartialContent(file.Body, schema)
			},
			Want: func(resource string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				// Removed "aws_s3_bucket" resource
				file := jsonFile("test.tf.json", `
{
  "resource": {
    "aws_instance": {
      "foo": {
        "instance_type": "t2.micro"
      }
    }
  }
}`)
				return hclext.Content(file.Body, &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
							},
						},
					},
				})
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "get content with options",
			Args: func() (string, *hclext.BodySchema, *tflint.GetModuleContentOption) {
				return "aws_instance", &hclext.BodySchema{}, &tflint.GetModuleContentOption{
					ModuleCtx: tflint.RootModuleCtxType,
				}
			},
			ServerImpl: func(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				if opts.ModuleCtx != tflint.RootModuleCtxType {
					return &hclext.BodyContent{}, hcl.Diagnostics{
						&hcl.Diagnostic{Severity: hcl.DiagError, Summary: "unexpected moduleCtx options"},
					}
				}
				if opts.Hint.ResourceType != "aws_instance" {
					return &hclext.BodyContent{}, hcl.Diagnostics{
						&hcl.Diagnostic{Severity: hcl.DiagError, Summary: "unexpected hint options"},
					}
				}
				return &hclext.BodyContent{}, hcl.Diagnostics{}
			},
			Want: func(resource string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				return &hclext.BodyContent{
					Attributes: hclext.Attributes{},
					Blocks:     hclext.Blocks{},
				}, hcl.Diagnostics{}
			},
			ErrCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{getModuleContent: test.ServerImpl, getFiles: fileExists}))

			got, err := client.GetResourceContent(test.Args())
			if test.ErrCheck(err) {
				t.Fatalf("failed to call GetResourceContent: %s", err)
			}
			want, diags := test.Want(test.Args())
			if diags.HasErrors() {
				t.Fatalf("failed to get want: %d diagsnotics", len(diags))
				for _, diag := range diags {
					t.Logf("  - %s", diag.Error())
				}
			}

			opts := cmp.Options{
				cmp.Comparer(func(x, y cty.Value) bool {
					return x.GoString() == y.GoString()
				}),
				cmpopts.EquateEmpty(),
				allowAllUnexported,
			}
			if diff := cmp.Diff(got, want, opts); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestGetProviderContent(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	// default getFileImpl function
	files := map[string][]byte{}
	fileExists := func() map[string][]byte {
		return files
	}

	// test util functions
	hclFile := func(filename string, code string) *hcl.File {
		file, diags := hclsyntax.ParseConfig([]byte(code), filename, hcl.InitialPos)
		if diags.HasErrors() {
			panic(diags)
		}
		files[filename] = file.Bytes
		return file
	}

	tests := []struct {
		Name       string
		Args       func() (string, *hclext.BodySchema, *tflint.GetModuleContentOption)
		ServerImpl func(*hclext.BodySchema, tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics)
		Want       func(string, *hclext.BodySchema, *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics)
		ErrCheck   func(error) bool
	}{
		{
			Name: "get HCL content",
			Args: func() (string, *hclext.BodySchema, *tflint.GetModuleContentOption) {
				return "aws", &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "region"}},
				}, nil
			},
			ServerImpl: func(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				file := hclFile("test.tf", `
provider "aws" {
  region = "us-east-1"
}

provider "google" {
  region = "us-central1"
}`)
				return hclext.PartialContent(file.Body, schema)
			},
			Want: func(resource string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				// Removed "google" provider
				file := hclFile("test.tf", `
provider "aws" {
  region = "us-east-1"
}`)
				return hclext.Content(file.Body, &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "provider",
							LabelNames: []string{"name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "region"}},
							},
						},
					},
				})
			},
			ErrCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{getModuleContent: test.ServerImpl, getFiles: fileExists}))

			got, err := client.GetProviderContent(test.Args())
			if test.ErrCheck(err) {
				t.Fatalf("failed to call GetProviderContent: %s", err)
			}
			want, diags := test.Want(test.Args())
			if diags.HasErrors() {
				t.Fatalf("failed to get want: %d diagsnotics", len(diags))
				for _, diag := range diags {
					t.Logf("  - %s", diag.Error())
				}
			}

			opts := cmp.Options{
				cmp.Comparer(func(x, y cty.Value) bool {
					return x.GoString() == y.GoString()
				}),
				cmpopts.EquateEmpty(),
				allowAllUnexported,
			}
			if diff := cmp.Diff(got, want, opts); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestGetModuleContent(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	// default getFileImpl function
	files := map[string][]byte{}
	fileExists := func() map[string][]byte {
		return files
	}

	// test util functions
	hclFile := func(filename string, code string) *hcl.File {
		file, diags := hclsyntax.ParseConfig([]byte(code), filename, hcl.InitialPos)
		if diags.HasErrors() {
			panic(diags)
		}
		files[filename] = file.Bytes
		return file
	}
	jsonFile := func(filename string, code string) *hcl.File {
		file, diags := json.Parse([]byte(code), filename)
		if diags.HasErrors() {
			panic(diags)
		}
		files[filename] = file.Bytes
		return file
	}

	tests := []struct {
		Name       string
		Args       func() (*hclext.BodySchema, *tflint.GetModuleContentOption)
		ServerImpl func(*hclext.BodySchema, tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics)
		Want       func(*hclext.BodySchema, *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics)
		ErrCheck   func(error) bool
	}{
		{
			Name: "get HCL content",
			Args: func() (*hclext.BodySchema, *tflint.GetModuleContentOption) {
				return &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
							},
						},
					},
				}, nil
			},
			ServerImpl: func(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				file := hclFile("test.tf", `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`)
				return hclext.Content(file.Body, schema)
			},
			Want: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				file := hclFile("test.tf", `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`)
				return hclext.Content(file.Body, schema)
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "get JSON content",
			Args: func() (*hclext.BodySchema, *tflint.GetModuleContentOption) {
				return &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
							},
						},
					},
				}, nil
			},
			ServerImpl: func(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				file := jsonFile("test.tf.json", `
{
  "resource": {
    "aws_instance": {
      "foo": {
        "instance_type": "t2.micro"
      }
    }
  }
}`)
				return hclext.Content(file.Body, schema)
			},
			Want: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				file := jsonFile("test.tf.json", `
{
  "resource": {
    "aws_instance": {
      "foo": {
        "instance_type": "t2.micro"
      }
    }
  }
}`)
				return hclext.Content(file.Body, schema)
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "get content as just attributes",
			Args: func() (*hclext.BodySchema, *tflint.GetModuleContentOption) {
				return &hclext.BodySchema{Mode: hclext.SchemaJustAttributesMode}, nil
			},
			ServerImpl: func(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				file := hclFile("test.tf", `
instance_type = "t2.micro"
volume_size = 10`)
				return hclext.Content(file.Body, schema)
			},
			Want: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				file := hclFile("test.tf", `
instance_type = "t2.micro"
volume_size = 10`)
				return hclext.Content(file.Body, schema)
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "get content with bound expr",
			Args: func() (*hclext.BodySchema, *tflint.GetModuleContentOption) {
				return &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "value"}},
				}, nil
			},
			ServerImpl: func(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				file := hclFile("test.tf", "value = each.key")
				attrs, diags := file.Body.JustAttributes()
				if diags.HasErrors() {
					return nil, diags
				}
				attr := attrs["value"]

				return &hclext.BodyContent{
					Attributes: hclext.Attributes{
						"value": {
							Name:      attr.Name,
							Expr:      hclext.BindValue(cty.StringVal("bound value"), attr.Expr),
							Range:     attr.Range,
							NameRange: attr.NameRange,
						},
					},
					Blocks: hclext.Blocks{},
				}, nil
			},
			Want: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				file := hclFile("test.tf", "value = each.key")
				attrs, diags := file.Body.JustAttributes()
				if diags.HasErrors() {
					return nil, diags
				}
				attr := attrs["value"]

				return &hclext.BodyContent{
					Attributes: hclext.Attributes{
						"value": {
							Name:      attr.Name,
							Expr:      hclext.BindValue(cty.StringVal("bound value"), attr.Expr),
							Range:     attr.Range,
							NameRange: attr.NameRange,
						},
					},
					Blocks: hclext.Blocks{},
				}, nil
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "get content with options",
			Args: func() (*hclext.BodySchema, *tflint.GetModuleContentOption) {
				return &hclext.BodySchema{}, &tflint.GetModuleContentOption{
					ModuleCtx:  tflint.RootModuleCtxType,
					ExpandMode: tflint.ExpandModeNone,
					Hint:       tflint.GetModuleContentHint{ResourceType: "aws_instance"},
				}
			},
			ServerImpl: func(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				if opts.ModuleCtx != tflint.RootModuleCtxType {
					return &hclext.BodyContent{}, hcl.Diagnostics{
						&hcl.Diagnostic{Severity: hcl.DiagError, Summary: "unexpected moduleCtx options"},
					}
				}
				if opts.ExpandMode != tflint.ExpandModeNone {
					return &hclext.BodyContent{}, hcl.Diagnostics{
						&hcl.Diagnostic{Severity: hcl.DiagError, Summary: "unexpected expand mode options"},
					}
				}
				if opts.Hint.ResourceType != "aws_instance" {
					return &hclext.BodyContent{}, hcl.Diagnostics{
						&hcl.Diagnostic{Severity: hcl.DiagError, Summary: "unexpected hint options"},
					}
				}
				return &hclext.BodyContent{}, hcl.Diagnostics{}
			},
			Want: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				return &hclext.BodyContent{
					Attributes: hclext.Attributes{},
					Blocks:     hclext.Blocks{},
				}, hcl.Diagnostics{}
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "schema is null",
			Args: func() (*hclext.BodySchema, *tflint.GetModuleContentOption) {
				return nil, nil
			},
			Want: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				return &hclext.BodyContent{
					Attributes: hclext.Attributes{},
					Blocks:     hclext.Blocks{},
				}, hcl.Diagnostics{}
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "server returns an error",
			Args: func() (*hclext.BodySchema, *tflint.GetModuleContentOption) {
				return &hclext.BodySchema{}, nil
			},
			ServerImpl: func(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				return &hclext.BodyContent{}, hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "unexpected error",
						Detail:   "unexpected error occurred",
						Subject:  &hcl.Range{Filename: "test.tf", Start: hcl.Pos{Line: 1, Column: 1}, End: hcl.Pos{Line: 1, Column: 5}},
					},
				}
			},
			Want: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				return nil, hcl.Diagnostics{}
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "test.tf:1,1-5: unexpected error; unexpected error occurred"
			},
		},
		{
			Name: "response body is empty",
			Args: func() (*hclext.BodySchema, *tflint.GetModuleContentOption) {
				return &hclext.BodySchema{}, nil
			},
			ServerImpl: func(schema *hclext.BodySchema, opts tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				return nil, hcl.Diagnostics{}
			},
			Want: func(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
				return nil, hcl.Diagnostics{}
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "response body is empty"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{getModuleContent: test.ServerImpl, getFiles: fileExists}))

			got, err := client.GetModuleContent(test.Args())
			if test.ErrCheck(err) {
				t.Fatalf("failed to call GetModuleContent: %s", err)
			}
			want, diags := test.Want(test.Args())
			if diags.HasErrors() {
				t.Fatalf("failed to get want: %d diagsnotics", len(diags))
				for _, diag := range diags {
					t.Logf("  - %s", diag.Error())
				}
			}

			opts := cmp.Options{
				cmp.Comparer(func(x, y cty.Value) bool {
					return x.GoString() == y.GoString()
				}),
				allowAllUnexported,
			}
			if diff := cmp.Diff(got, want, opts); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestGetFile(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	// test util functions
	hclFile := func(filename string, code string) (*hcl.File, error) {
		file, diags := hclsyntax.ParseConfig([]byte(code), filename, hcl.InitialPos)
		if diags.HasErrors() {
			return nil, diags
		}
		return file, nil
	}
	jsonFile := func(filename string, code string) (*hcl.File, error) {
		file, diags := json.Parse([]byte(code), filename)
		if diags.HasErrors() {
			return nil, diags
		}
		return file, nil
	}

	tests := []struct {
		Name       string
		Arg        string
		ServerImpl func(string) (*hcl.File, error)
		Want       string
		ErrCheck   func(error) bool
	}{
		{
			Name: "HCL file exists",
			Arg:  "test.tf",
			ServerImpl: func(filename string) (*hcl.File, error) {
				if filename != "test.tf" {
					return nil, nil
				}
				return hclFile(filename, `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`)
			},
			Want: `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`,
			ErrCheck: neverHappend,
		},
		{
			Name: "JSON file exists",
			Arg:  "test.tf.json",
			ServerImpl: func(filename string) (*hcl.File, error) {
				if filename != "test.tf.json" {
					return nil, nil
				}
				return jsonFile(filename, `
{
  "resource": {
    "aws_instance": {
      "foo": {
        "instance_type": "t2.micro"
      }
    }
  }
}`)
			},
			Want: `
{
  "resource": {
    "aws_instance": {
      "foo": {
        "instance_type": "t2.micro"
      }
    }
  }
}`,
			ErrCheck: neverHappend,
		},
		{
			Name: "file not found",
			Arg:  "test.tf",
			ServerImpl: func(filename string) (*hcl.File, error) {
				return nil, nil
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "file not found"
			},
		},
		{
			Name: "server returns an error",
			Arg:  "test.tf",
			ServerImpl: func(filename string) (*hcl.File, error) {
				if filename != "test.tf" {
					return nil, nil
				}
				return nil, errors.New("unexpected error")
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
		},
		{
			Name: "filename is empty",
			Arg:  "",
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "name should not be empty"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{getFile: test.ServerImpl}))

			file, err := client.GetFile(test.Arg)
			if test.ErrCheck(err) {
				t.Fatalf("failed to call GetFile: %s", err)
			}

			var got string
			if file != nil {
				got = string(file.Bytes)
			}

			if got != test.Want {
				t.Errorf("got: %s", got)
			}
		})
	}
}

func TestGetFiles(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	// test util functions
	hclFile := func(filename string, code string) *hcl.File {
		file, diags := hclsyntax.ParseConfig([]byte(code), filename, hcl.InitialPos)
		if diags.HasErrors() {
			panic(diags)
		}
		return file
	}
	jsonFile := func(filename string, code string) *hcl.File {
		file, diags := json.Parse([]byte(code), filename)
		if diags.HasErrors() {
			panic(diags)
		}
		return file
	}

	tests := []struct {
		Name       string
		ServerImpl func() map[string][]byte
		Want       map[string]*hcl.File
		ErrCheck   func(error) bool
	}{
		{
			Name: "HCL files",
			ServerImpl: func() map[string][]byte {
				return map[string][]byte{
					"test1.tf": []byte(`
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`),
					"test2.tf": []byte(`
resource "aws_s3_bucket" "bar" {
	bucket = "baz"
}`),
				}
			},
			Want: map[string]*hcl.File{
				"test1.tf": hclFile("test1.tf", `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`),
				"test2.tf": hclFile("test2.tf", `
resource "aws_s3_bucket" "bar" {
	bucket = "baz"
}`),
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "JSON files",
			ServerImpl: func() map[string][]byte {
				return map[string][]byte{
					"test1.tf.json": []byte(`
{
  "resource": {
    "aws_instance": {
      "foo": {
        "instance_type": "t2.micro"
      }
    }
  }
}`),
					"test2.tf.json": []byte(`
{
  "resource": {
    "aws_s3_bucket": {
      "bar": {
        "bucket": "baz"
      }
    }
  }
}`),
				}
			},
			Want: map[string]*hcl.File{
				"test1.tf.json": jsonFile("test1.tf.json", `
{
  "resource": {
    "aws_instance": {
      "foo": {
        "instance_type": "t2.micro"
      }
    }
  }
}`),
				"test2.tf.json": jsonFile("test2.tf.json", `
{
  "resource": {
    "aws_s3_bucket": {
      "bar": {
        "bucket": "baz"
      }
    }
  }
}`),
			},
			ErrCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{getFiles: test.ServerImpl}))

			files, err := client.GetFiles()
			if test.ErrCheck(err) {
				t.Fatalf("failed to call GetFiles: %s", err)
			}

			opts := cmp.Options{
				cmp.Comparer(func(x, y cty.Value) bool {
					return x.GoString() == y.GoString()
				}),
				cmp.AllowUnexported(hclsyntax.Body{}),
				cmpopts.IgnoreFields(hcl.File{}, "Nav"),
				allowAllUnexported,
			}
			if diff := cmp.Diff(files, test.Want, opts); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestWalkExpressions(t *testing.T) {
	tests := []struct {
		name   string
		files  map[string][]byte
		walked []hcl.Range
	}{
		{
			name: "resource",
			files: map[string][]byte{
				"resource.tf": []byte(`
resource "null_resource" "test" {
  key = "foo"
}`),
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			name: "data source",
			files: map[string][]byte{
				"data.tf": []byte(`
data "null_dataresource" "test" {
  key = "foo"
}`),
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			name: "module call",
			files: map[string][]byte{
				"module.tf": []byte(`
module "m" {
  source = "./module"
  key    = "foo"
}`),
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
			files: map[string][]byte{
				"provider.tf": []byte(`
provider "p" {
  key = "foo"
}`),
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			name: "locals",
			files: map[string][]byte{
				"locals.tf": []byte(`
locals {
  key = "foo"
}`),
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			name: "output",
			files: map[string][]byte{
				"output.tf": []byte(`
output "o" {
  value = "foo"
}`),
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 11}, End: hcl.Pos{Line: 3, Column: 16}},
				{Start: hcl.Pos{Line: 3, Column: 12}, End: hcl.Pos{Line: 3, Column: 15}},
			},
		},
		{
			name: "resource with block",
			files: map[string][]byte{
				"resource.tf": []byte(`
resource "null_resource" "test" {
  key = "foo"

  lifecycle {
    ignore_changes = [key]
  }
}`),
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
			files: map[string][]byte{
				"resource.tf.json": []byte(`
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
}`),
			},
			walked: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 15}, End: hcl.Pos{Line: 15, Column: 4}},
			},
		},
		{
			name: "multiple files",
			files: map[string][]byte{
				"main.tf": []byte(`
provider "aws" {
  region = "us-east-1"

  assume_role {
    role_arn = "arn:aws:iam::123412341234:role/ExampleRole"
  }
}`),
				"main_override.tf": []byte(`
provider "aws" {
  region = "us-east-1"

  assume_role {
    role_arn = null
  }
}`),
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
			files: map[string][]byte{
				"data.tf": []byte(`
data "terraform_remote_state" "remote_state" {
  backend = "remote"

  config = {
    organization = "Organization"
    workspaces = {
      name = "${var.environment}"
    }
  }
}`),
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
			getFilesImpl := func() map[string][]byte { return test.files }
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{getFiles: getFilesImpl}))

			walked := []hcl.Range{}
			diags := client.WalkExpressions(tflint.ExprWalkFunc(func(expr hcl.Expression) hcl.Diagnostics {
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

func TestDecodeRuleConfig(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	// test struct for decoding
	type ruleConfig struct {
		Name string `hclext:"name"`
	}

	tests := []struct {
		Name       string
		RuleName   string
		Target     interface{}
		ServerImpl func(string, *hclext.BodySchema) (*hclext.BodyContent, map[string][]byte, error)
		Want       interface{}
		ErrCheck   func(error) bool
	}{
		{
			Name:     "decode to struct",
			RuleName: "test_rule",
			Target:   &ruleConfig{},
			ServerImpl: func(name string, schema *hclext.BodySchema) (*hclext.BodyContent, map[string][]byte, error) {
				if name != "test_rule" {
					return &hclext.BodyContent{}, map[string][]byte{}, errors.New("unexpected file name")
				}

				sources := map[string][]byte{
					".tflint.hcl": []byte(`
rule "test_rule" {
	name = "foo"
}`),
				}

				file, diags := hclsyntax.ParseConfig(sources[".tflint.hcl"], ".tflint.hcl", hcl.InitialPos)
				if diags.HasErrors() {
					return &hclext.BodyContent{}, sources, diags
				}

				content, diags := file.Body.Content(&hcl.BodySchema{
					Blocks: []hcl.BlockHeaderSchema{{Type: "rule", LabelNames: []string{"name"}}},
				})
				if diags.HasErrors() {
					return &hclext.BodyContent{}, sources, diags
				}

				body, diags := hclext.Content(content.Blocks[0].Body, schema)
				if diags.HasErrors() {
					return &hclext.BodyContent{}, sources, diags
				}
				return body, sources, nil
			},
			Want:     &ruleConfig{Name: "foo"},
			ErrCheck: neverHappend,
		},
		{
			Name:     "server returns an error",
			RuleName: "test_rule",
			Target:   &ruleConfig{},
			ServerImpl: func(name string, schema *hclext.BodySchema) (*hclext.BodyContent, map[string][]byte, error) {
				return nil, map[string][]byte{}, errors.New("unexpected error")
			},
			Want: &ruleConfig{},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
		},
		{
			Name:     "response body is empty",
			RuleName: "test_rule",
			Target:   &ruleConfig{},
			ServerImpl: func(name string, schema *hclext.BodySchema) (*hclext.BodyContent, map[string][]byte, error) {
				return nil, map[string][]byte{}, nil
			},
			Want: &ruleConfig{},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "response body is empty"
			},
		},
		{
			Name:     "config not found",
			RuleName: "not_found",
			Target:   &ruleConfig{},
			ServerImpl: func(name string, schema *hclext.BodySchema) (*hclext.BodyContent, map[string][]byte, error) {
				return &hclext.BodyContent{}, nil, nil
			},
			Want:     &ruleConfig{},
			ErrCheck: neverHappend,
		},
		{
			Name:     "config not found with non-empty config",
			RuleName: "not_found",
			Target:   &ruleConfig{},
			ServerImpl: func(name string, schema *hclext.BodySchema) (*hclext.BodyContent, map[string][]byte, error) {
				return &hclext.BodyContent{Attributes: hclext.Attributes{"foo": &hclext.Attribute{}}}, nil, nil
			},
			Want: &ruleConfig{},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "config file not found"
			},
		},
		{
			Name:     "name is empty",
			RuleName: "",
			Target:   &ruleConfig{},
			Want:     &ruleConfig{},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "name should not be empty"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{getRuleConfigContent: test.ServerImpl}))

			err := client.DecodeRuleConfig(test.RuleName, test.Target)
			if test.ErrCheck(err) {
				t.Fatalf("failed to call DecodeRuleConfig: %s", err)
			}

			if diff := cmp.Diff(test.Target, test.Want); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestEvaluateExpr(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	// default getFileImpl function
	fileIdx := 1
	files := map[string]*hcl.File{}
	fileExists := func(filename string) (*hcl.File, error) {
		return files[filename], nil
	}

	// test util functions
	hclExpr := func(expr string) hcl.Expression {
		filename := fmt.Sprintf("test%d.tf", fileIdx)
		file, diags := hclsyntax.ParseConfig([]byte(fmt.Sprintf(`expr = %s`, expr)), filename, hcl.InitialPos)
		if diags.HasErrors() {
			panic(diags)
		}
		attributes, diags := file.Body.JustAttributes()
		if diags.HasErrors() {
			panic(diags)
		}
		files[filename] = file
		fileIdx = fileIdx + 1
		return attributes["expr"].Expr
	}
	jsonExpr := func(expr string) hcl.Expression {
		filename := fmt.Sprintf("test%d.tf.json", fileIdx)
		file, diags := json.Parse([]byte(fmt.Sprintf(`{"expr": %s}`, expr)), filename)
		if diags.HasErrors() {
			panic(diags)
		}
		attributes, diags := file.Body.JustAttributes()
		if diags.HasErrors() {
			panic(diags)
		}
		files[filename] = file
		fileIdx = fileIdx + 1
		return attributes["expr"].Expr
	}
	evalExpr := func(expr hcl.Expression, ctx *hcl.EvalContext) (cty.Value, error) {
		val, diags := expr.Value(ctx)
		if diags.HasErrors() {
			return cty.Value{}, diags
		}
		return val, nil
	}

	// test struct for decoding from cty.Value
	type Object struct {
		Name    string `cty:"name"`
		Enabled bool   `cty:"enabled"`
	}
	objectTy := cty.Object(map[string]cty.Type{"name": cty.String, "enabled": cty.Bool})

	tests := []struct {
		Name        string
		Expr        hcl.Expression
		TargetType  reflect.Type
		Option      *tflint.EvaluateExprOption
		ServerImpl  func(hcl.Expression, tflint.EvaluateExprOption) (cty.Value, error)
		GetFileImpl func(string) (*hcl.File, error)
		Want        interface{}
		ErrCheck    func(err error) bool
	}{
		{
			Name:       "literal",
			Expr:       hclExpr(`"foo"`),
			TargetType: reflect.TypeOf(""),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				if *opts.WantType != cty.String {
					return cty.Value{}, errors.New("wantType should be string")
				}
				if opts.ModuleCtx != tflint.SelfModuleCtxType {
					return cty.Value{}, errors.New("moduleCtx should be self")
				}
				return evalExpr(expr, nil)
			},
			Want:        "foo",
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "string variable",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(""),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("bar"),
						}),
					},
				})
			},
			Want:        "bar",
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "number variable",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(0),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				if *opts.WantType != cty.Number {
					return cty.Value{}, errors.New("wantType should be number")
				}
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.NumberIntVal(7),
						}),
					},
				})
			},
			Want:        7,
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "bool variable",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(true),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				if *opts.WantType != cty.Bool {
					return cty.Value{}, errors.New("wantType should be bool")
				}
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.BoolVal(true),
						}),
					},
				})
			},
			Want:        true,
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "string list variable",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf([]string{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				if *opts.WantType != cty.List(cty.String) {
					return cty.Value{}, errors.New("wantType should be string list")
				}
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}),
						}),
					},
				})
			},
			Want:        []string{"foo", "bar"},
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "number list variable",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf([]int{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				if *opts.WantType != cty.List(cty.Number) {
					return cty.Value{}, errors.New("wantType should be number list")
				}
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.ListVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2)}),
						}),
					},
				})
			},
			Want:        []int{1, 2},
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "bool list variable",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf([]bool{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				if *opts.WantType != cty.List(cty.Bool) {
					return cty.Value{}, errors.New("wantType should be bool list")
				}
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.ListVal([]cty.Value{cty.BoolVal(true), cty.BoolVal(false)}),
						}),
					},
				})
			},
			Want:        []bool{true, false},
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "string map variable",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(map[string]string{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				if *opts.WantType != cty.Map(cty.String) {
					return cty.Value{}, errors.New("wantType should be string map")
				}
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("bar"), "baz": cty.StringVal("qux")}),
						}),
					},
				})
			},
			Want:        map[string]string{"foo": "bar", "baz": "qux"},
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "number map variable",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(map[string]int{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				if *opts.WantType != cty.Map(cty.Number) {
					return cty.Value{}, errors.New("wantType should be number map")
				}
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.MapVal(map[string]cty.Value{"foo": cty.NumberIntVal(1), "bar": cty.NumberIntVal(2)}),
						}),
					},
				})
			},
			Want:        map[string]int{"foo": 1, "bar": 2},
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "bool map variable",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(map[string]bool{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				if *opts.WantType != cty.Map(cty.Bool) {
					return cty.Value{}, errors.New("wantType should be bool map")
				}
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.MapVal(map[string]cty.Value{"foo": cty.BoolVal(true), "bar": cty.BoolVal(false)}),
						}),
					},
				})
			},
			Want:        map[string]bool{"foo": true, "bar": false},
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "object variable",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(cty.Value{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				if *opts.WantType != cty.DynamicPseudoType {
					return cty.Value{}, errors.New("wantType should be pseudo type")
				}
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.ObjectVal(map[string]cty.Value{
								"foo": cty.NumberIntVal(1),
								"bar": cty.StringVal("baz"),
								"qux": cty.UnknownVal(cty.String),
							}),
						}),
					},
				})
			},
			Want: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.NumberIntVal(1),
				"bar": cty.StringVal("baz"),
				"qux": cty.UnknownVal(cty.String),
			}),
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "object variable to struct",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(Object{}),
			Option:     &tflint.EvaluateExprOption{WantType: &objectTy},
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.ObjectVal(map[string]cty.Value{
								"name":    cty.StringVal("bar"),
								"enabled": cty.BoolVal(true),
							}),
						}),
					},
				})
			},
			Want:        Object{Name: "bar", Enabled: true},
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "JSON expr",
			Expr:       jsonExpr(`"${var.foo}"`),
			TargetType: reflect.TypeOf(""),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("bar"),
						}),
					},
				})
			},
			Want:        "bar",
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "JSON object",
			Expr:       jsonExpr(`{"foo": "bar"}`),
			TargetType: reflect.TypeOf(map[string]string{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, nil)
			},
			Want:        map[string]string{"foo": "bar"},
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "bound expr",
			Expr:       hclext.BindValue(cty.StringVal("bound value"), hclExpr(`var.foo`)),
			TargetType: reflect.TypeOf(""),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{})
			},
			Want:        "bound value",
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "eval with moduleCtx option",
			Expr:       hclExpr(`1`),
			TargetType: reflect.TypeOf(0),
			Option:     &tflint.EvaluateExprOption{ModuleCtx: tflint.RootModuleCtxType},
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				if opts.ModuleCtx != tflint.RootModuleCtxType {
					return cty.Value{}, errors.New("moduleCtx should be root")
				}
				return evalExpr(expr, nil)
			},
			Want:        1,
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "getFile returns no file",
			Expr:       hclExpr(`1`),
			TargetType: reflect.TypeOf(0),
			Want:       0,
			GetFileImpl: func(string) (*hcl.File, error) {
				return nil, nil
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "file not found"
			},
		},
		{
			Name:       "getFile returns an error",
			Expr:       hclExpr(`1`),
			TargetType: reflect.TypeOf(0),
			Want:       0,
			GetFileImpl: func(string) (*hcl.File, error) {
				return nil, errors.New("unexpected error")
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
		},
		{
			Name:       "server returns an unexpected error",
			Expr:       hclExpr(`1`),
			TargetType: reflect.TypeOf(0),
			ServerImpl: func(hcl.Expression, tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.Value{}, errors.New("unexpected error")
			},
			Want:        0,
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
		},
		{
			Name:       "server returns an unknown error",
			Expr:       hclExpr(`1`),
			TargetType: reflect.TypeOf(0),
			ServerImpl: func(hcl.Expression, tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.Value{}, fmt.Errorf("unknown%w", tflint.ErrUnknownValue)
			},
			Want:        0,
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return !errors.Is(err, tflint.ErrUnknownValue)
			},
		},
		{
			Name:       "server returns a null value error",
			Expr:       hclExpr(`1`),
			TargetType: reflect.TypeOf(0),
			ServerImpl: func(hcl.Expression, tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.Value{}, fmt.Errorf("null value%w", tflint.ErrNullValue)
			},
			Want:        0,
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return !errors.Is(err, tflint.ErrNullValue)
			},
		},
		{
			Name:       "server returns a unevaluable error",
			Expr:       hclExpr(`1`),
			TargetType: reflect.TypeOf(0),
			ServerImpl: func(hcl.Expression, tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.Value{}, fmt.Errorf("unevaluable%w", tflint.ErrUnevaluable)
			},
			Want:        0,
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return !errors.Is(err, tflint.ErrUnevaluable)
			},
		},
		{
			Name:       "server returns a sensitive error",
			Expr:       hclExpr(`1`),
			TargetType: reflect.TypeOf(0),
			ServerImpl: func(hcl.Expression, tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.Value{}, fmt.Errorf("sensitive%w", tflint.ErrSensitive)
			},
			Want:        0,
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return !errors.Is(err, tflint.ErrSensitive)
			},
		},
		{
			Name:       "unknown value",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(""),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.UnknownVal(cty.String),
						}),
					},
				})
			},
			Want:        "",
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return !errors.Is(err, tflint.ErrUnknownValue)
			},
		},
		{
			Name:       "unknown value as cty.Value",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(cty.Value{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.UnknownVal(cty.String),
						}),
					},
				})
			},
			Want:        cty.UnknownVal(cty.String),
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "unknown value in object",
			Expr:       hclExpr(`{ value = var.foo }`),
			TargetType: reflect.TypeOf(map[string]string{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.UnknownVal(cty.String),
						}),
					},
				})
			},
			Want:        (map[string]string)(nil),
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return !errors.Is(err, tflint.ErrUnknownValue)
			},
		},
		{
			Name:       "null",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(""),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.NullVal(cty.String),
						}),
					},
				})
			},
			Want:        "",
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return !errors.Is(err, tflint.ErrNullValue)
			},
		},
		{
			Name:       "null as cty.Value",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(cty.Value{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.NullVal(cty.String),
						}),
					},
				})
			},
			Want:        cty.NullVal(cty.String),
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "null value in object",
			Expr:       hclExpr(`{ value = var.foo }`),
			TargetType: reflect.TypeOf(map[string]string{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.NullVal(cty.String),
						}),
					},
				})
			},
			Want:        (map[string]string)(nil),
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return !errors.Is(err, tflint.ErrNullValue)
			},
		},
		{
			Name:       "sensitive",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(""),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("bar").Mark(marks.Sensitive),
						}),
					},
				})
			},
			Want:        "",
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return !errors.Is(err, tflint.ErrSensitive)
			},
		},
		{
			Name:       "sensitive as cty.Value",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(cty.Value{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("bar").Mark(marks.Sensitive),
						}),
					},
				})
			},
			Want:        cty.StringVal("bar").Mark(marks.Sensitive),
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "sensitive in object",
			Expr:       hclExpr(`{ value = var.foo }`),
			TargetType: reflect.TypeOf(map[string]string{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("bar").Mark(marks.Sensitive),
						}),
					},
				})
			},
			Want:        (map[string]string)(nil),
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return !errors.Is(err, tflint.ErrSensitive)
			},
		},
		{
			Name:       "sensitive object as cty.Value",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(cty.Value{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.ObjectVal(map[string]cty.Value{
								"bar": cty.StringVal("barval").Mark(marks.Sensitive),
								"baz": cty.ListVal([]cty.Value{cty.NumberIntVal(1).Mark(marks.Sensitive)}),
								"qux": cty.TupleVal([]cty.Value{cty.StringVal("quxval").Mark(marks.Sensitive)}),
								"quux": cty.MapVal(map[string]cty.Value{
									"foo": cty.StringVal("fooval").Mark(marks.Sensitive),
								}),
								"corge": cty.ObjectVal(map[string]cty.Value{
									"bar": cty.NumberIntVal(2).Mark(marks.Sensitive),
								}),
							}),
						}),
					},
				})
			},
			Want: cty.ObjectVal(map[string]cty.Value{
				"bar": cty.StringVal("barval").Mark(marks.Sensitive),
				"baz": cty.ListVal([]cty.Value{cty.NumberIntVal(1).Mark(marks.Sensitive)}),
				"qux": cty.TupleVal([]cty.Value{cty.StringVal("quxval").Mark(marks.Sensitive)}),
				"quux": cty.MapVal(map[string]cty.Value{
					"foo": cty.StringVal("fooval").Mark(marks.Sensitive),
				}),
				"corge": cty.ObjectVal(map[string]cty.Value{
					"bar": cty.NumberIntVal(2).Mark(marks.Sensitive),
				}),
			}),
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
		{
			Name:       "ephemeral",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(""),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("bar").Mark(marks.Ephemeral),
						}),
					},
				})
			},
			Want:        "",
			GetFileImpl: fileExists,
			ErrCheck: func(err error) bool {
				return !errors.Is(err, tflint.ErrEphemeral)
			},
		},
		{
			Name:       "ephemeral as cty.Value",
			Expr:       hclExpr(`var.foo`),
			TargetType: reflect.TypeOf(cty.Value{}),
			ServerImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return evalExpr(expr, &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("bar").Mark(marks.Ephemeral),
						}),
					},
				})
			},
			Want:        cty.StringVal("bar").Mark(marks.Ephemeral),
			GetFileImpl: fileExists,
			ErrCheck:    neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			target := reflect.New(test.TargetType)

			client := startTestGRPCServer(t, newMockServer(mockServerImpl{evaluateExpr: test.ServerImpl, getFile: test.GetFileImpl}))

			err := client.EvaluateExpr(test.Expr, target.Interface(), test.Option)
			if test.ErrCheck(err) {
				t.Fatalf("failed to call EvaluateExpr: %s", err)
			}

			got := target.Elem().Interface()

			opts := cmp.Options{
				cmp.Comparer(func(x, y cty.Value) bool {
					return x.GoString() == y.GoString()
				}),
			}
			if diff := cmp.Diff(got, test.Want, opts); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestEvaluateExpr_callback(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	// default getFileImpl function
	fileIdx := 1
	files := map[string]*hcl.File{}
	fileExists := func(filename string) (*hcl.File, error) {
		return files[filename], nil
	}

	// test util functions
	hclExpr := func(expr string) hcl.Expression {
		filename := fmt.Sprintf("test%d.tf", fileIdx)
		file, diags := hclsyntax.ParseConfig([]byte(fmt.Sprintf(`expr = %s`, expr)), filename, hcl.InitialPos)
		if diags.HasErrors() {
			panic(diags)
		}
		attributes, diags := file.Body.JustAttributes()
		if diags.HasErrors() {
			panic(diags)
		}
		files[filename] = file
		fileIdx = fileIdx + 1
		return attributes["expr"].Expr
	}

	// test struct for decoding from cty.Value
	type Object struct {
		Name    string `cty:"name"`
		Enabled bool   `cty:"enabled"`
	}
	objectTy := cty.Object(map[string]cty.Type{"name": cty.String, "enabled": cty.Bool})

	tests := []struct {
		name        string
		expr        hcl.Expression
		target      any
		option      *tflint.EvaluateExprOption
		serverImpl  func(hcl.Expression, tflint.EvaluateExprOption) (cty.Value, error)
		getFileImpl func(string) (*hcl.File, error)
		errCheck    func(err error) bool
	}{
		{
			name: "callback with string",
			expr: hclExpr(`"foo"`),
			target: func(val string) error {
				return fmt.Errorf("value is %s", val)
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.StringVal("foo"), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "value is foo"
			},
		},
		{
			name: "callback with int",
			expr: hclExpr(`1`),
			target: func(val int) error {
				return fmt.Errorf("value is %d", val)
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.NumberIntVal(1), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "value is 1"
			},
		},
		{
			name: "callback with bool",
			expr: hclExpr(`true`),
			target: func(val bool) error {
				return fmt.Errorf("value is %t", val)
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.BoolVal(true), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "value is true"
			},
		},
		{
			name: "callback with []string",
			expr: hclExpr(`["foo", "bar"]`),
			target: func(val []string) error {
				return fmt.Errorf("value is %#v", val)
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `value is []string{"foo", "bar"}`
			},
		},
		{
			name: "callback with []int",
			expr: hclExpr(`[1, 2]`),
			target: func(val []int) error {
				return fmt.Errorf("value is %#v", val)
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.ListVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2)}), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `value is []int{1, 2}`
			},
		},
		{
			name: "callback with []bool",
			expr: hclExpr(`[true, false]`),
			target: func(val []bool) error {
				return fmt.Errorf("value is %#v", val)
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.ListVal([]cty.Value{cty.BoolVal(true), cty.BoolVal(false)}), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `value is []bool{true, false}`
			},
		},
		{
			name: "callback with map[string]string",
			expr: hclExpr(`{ "foo" = "bar", "baz" = "qux" }`),
			target: func(val map[string]string) error {
				return fmt.Errorf("foo is %s, baz is %s", val["foo"], val["baz"])
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.MapVal(map[string]cty.Value{
					"foo": cty.StringVal("bar"),
					"baz": cty.StringVal("qux"),
				}), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `foo is bar, baz is qux`
			},
		},
		{
			name: "callback with map[string]int",
			expr: hclExpr(`{ "foo" = "bar", "baz" = "qux" }`),
			target: func(val map[string]int) error {
				return fmt.Errorf("foo is %d, baz is %d", val["foo"], val["baz"])
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.MapVal(map[string]cty.Value{
					"foo": cty.NumberIntVal(1),
					"baz": cty.NumberIntVal(2),
				}), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `foo is 1, baz is 2`
			},
		},
		{
			name: "callback with map[string]bool",
			expr: hclExpr(`{ "foo" = true, "baz" = false }`),
			target: func(val map[string]bool) error {
				return fmt.Errorf("foo is %t, baz is %t", val["foo"], val["baz"])
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.MapVal(map[string]cty.Value{
					"foo": cty.BoolVal(true),
					"baz": cty.BoolVal(false),
				}), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `foo is true, baz is false`
			},
		},
		{
			name: "callback with cty.Value",
			expr: hclExpr(`var.foo`),
			target: func(val cty.Value) error {
				return fmt.Errorf("value is %s", val.GoString())
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.ObjectVal(map[string]cty.Value{
					"foo": cty.NumberIntVal(1),
					"bar": cty.StringVal("baz"),
					"qux": cty.UnknownVal(cty.String),
				}), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `value is cty.ObjectVal(map[string]cty.Value{"bar":cty.StringVal("baz"), "foo":cty.NumberIntVal(1), "qux":cty.UnknownVal(cty.String)})`
			},
		},
		{
			name: "callback with struct",
			expr: hclExpr(`var.foo`),
			target: func(val Object) error {
				return fmt.Errorf("value is %#v", val)
			},
			option: &tflint.EvaluateExprOption{WantType: &objectTy},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.ObjectVal(map[string]cty.Value{
					"name":    cty.StringVal("bar"),
					"enabled": cty.BoolVal(true),
				}), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `value is plugin2host.Object{Name:"bar", Enabled:true}`
			},
		},
		{
			name: "callback with unknown value as Go value",
			expr: hclExpr(`var.foo`),
			target: func(val string) error {
				return fmt.Errorf("value is %s", val)
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.UnknownVal(cty.String), nil
			},
			getFileImpl: fileExists,
			errCheck:    neverHappend,
		},
		{
			name: "callback with unknown value as cty.Value",
			expr: hclExpr(`var.foo`),
			target: func(val cty.Value) error {
				return fmt.Errorf("value is %s", val.GoString())
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.UnknownVal(cty.String), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `value is cty.UnknownVal(cty.String)`
			},
		},
		{
			name: "callback with null as Go value",
			expr: hclExpr(`var.foo`),
			target: func(val string) error {
				return fmt.Errorf("value is %s", val)
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.NullVal(cty.String), nil
			},
			getFileImpl: fileExists,
			errCheck:    neverHappend,
		},
		{
			name: "callback with null as cty.Value",
			expr: hclExpr(`var.foo`),
			target: func(val cty.Value) error {
				return fmt.Errorf("value is %s", val.GoString())
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.NullVal(cty.String), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `value is cty.NullVal(cty.String)`
			},
		},
		{
			name: "callback with sensitive value as Go value",
			expr: hclExpr(`var.foo`),
			target: func(val string) error {
				return fmt.Errorf("value is %s", val)
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.StringVal("foo").Mark(marks.Sensitive), nil
			},
			getFileImpl: fileExists,
			errCheck:    neverHappend,
		},
		{
			name: "callback with sensitive value as cty.Value",
			expr: hclExpr(`var.foo`),
			target: func(val cty.Value) error {
				return fmt.Errorf("value is %s", val.GoString())
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.StringVal("foo").Mark(marks.Sensitive), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `value is cty.StringVal("foo").Mark(marks.Sensitive)`
			},
		},
		{
			name: "callback with ephemeral value as Go value",
			expr: hclExpr(`var.foo`),
			target: func(val string) error {
				return fmt.Errorf("value is %s", val)
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.StringVal("foo").Mark(marks.Ephemeral), nil
			},
			getFileImpl: fileExists,
			errCheck:    neverHappend,
		},
		{
			name: "callback with ephemeral value as cty.Value",
			expr: hclExpr(`var.foo`),
			target: func(val cty.Value) error {
				return fmt.Errorf("value is %s", val.GoString())
			},
			serverImpl: func(expr hcl.Expression, opts tflint.EvaluateExprOption) (cty.Value, error) {
				return cty.StringVal("foo").Mark(marks.Ephemeral), nil
			},
			getFileImpl: fileExists,
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `value is cty.StringVal("foo").Mark(marks.Ephemeral)`
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{evaluateExpr: test.serverImpl, getFile: test.getFileImpl}))

			err := client.EvaluateExpr(test.expr, test.target, test.option)
			if test.errCheck(err) {
				t.Fatalf("failed to call EvaluateExpr: %s", err)
			}
		})
	}
}

// test rule for TestEmitIssue
type Rule struct {
	tflint.DefaultRule
}

func (*Rule) Name() string                     { return "test_rule" }
func (*Rule) Enabled() bool                    { return true }
func (*Rule) Severity() tflint.Severity        { return tflint.ERROR }
func (*Rule) Link() string                     { return "https://example.com" }
func (*Rule) Check(runner tflint.Runner) error { return nil }

func TestEmitIssue(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		Name       string
		Args       func() (tflint.Rule, string, hcl.Range)
		ServerImpl func(tflint.Rule, string, hcl.Range, bool) (bool, error)
		ErrCheck   func(error) bool
	}{
		{
			Name: "emit issue",
			Args: func() (tflint.Rule, string, hcl.Range) {
				return &Rule{}, "this is test", hcl.Range{Filename: "test.tf", Start: hcl.Pos{Line: 2, Column: 2}, End: hcl.Pos{Line: 2, Column: 10}}
			},
			ServerImpl: func(rule tflint.Rule, message string, location hcl.Range, fixable bool) (bool, error) {
				if rule.Name() != "test_rule" {
					return false, fmt.Errorf("rule.Name() should be test_rule, but %s", rule.Name())
				}
				if rule.Enabled() != true {
					return false, fmt.Errorf("rule.Enabled() should be true, but %t", rule.Enabled())
				}
				if rule.Severity() != tflint.ERROR {
					return false, fmt.Errorf("rule.Severity() should be ERROR, but %s", rule.Severity())
				}
				if rule.Link() != "https://example.com" {
					return false, fmt.Errorf("rule.Link() should be https://example.com, but %s", rule.Link())
				}
				if message != "this is test" {
					return false, fmt.Errorf("message should be `this is test`, but %s", message)
				}
				want := hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 2},
					End:      hcl.Pos{Line: 2, Column: 10},
				}
				if diff := cmp.Diff(location, want); diff != "" {
					return false, fmt.Errorf("diff: %s", diff)
				}
				return true, nil
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "server returns an error",
			Args: func() (tflint.Rule, string, hcl.Range) {
				return &Rule{}, "this is test", hcl.Range{Filename: "test.tf", Start: hcl.Pos{Line: 2, Column: 2}, End: hcl.Pos{Line: 2, Column: 10}}
			},
			ServerImpl: func(tflint.Rule, string, hcl.Range, bool) (bool, error) {
				return false, errors.New("unexpected error")
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{emitIssue: test.ServerImpl}))

			err := client.EmitIssue(test.Args())
			if test.ErrCheck(err) {
				t.Fatalf("failed to call EmitIssue: %s", err)
			}
		})
	}
}

func TestEmitIssueWithFix(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }
	getFiles := func() map[string][]byte {
		return map[string][]byte{
			"test.tf": []byte(`foo = "bar"`),
		}
	}

	tests := []struct {
		Name       string
		Args       func() (tflint.Rule, string, hcl.Range, func(tflint.Fixer) error)
		ServerImpl func(tflint.Rule, string, hcl.Range, bool) (bool, error)
		ModulePath []string
		DisableFix bool
		ErrCheck   func(error) bool
		Changes    map[string]string
	}{
		{
			Name: "emit issue",
			Args: func() (tflint.Rule, string, hcl.Range, func(tflint.Fixer) error) {
				return &Rule{},
					"this is test",
					hcl.Range{Filename: "test.tf", Start: hcl.Pos{Line: 2, Column: 2}, End: hcl.Pos{Line: 2, Column: 10}},
					func(f tflint.Fixer) error {
						return f.ReplaceText(
							hcl.Range{Filename: "test.tf", Start: hcl.Pos{Byte: 7}, End: hcl.Pos{Byte: 10}},
							"baz",
						)
					}
			},
			ServerImpl: func(rule tflint.Rule, message string, location hcl.Range, fixable bool) (bool, error) {
				if rule.Name() != "test_rule" {
					return false, fmt.Errorf("rule.Name() should be test_rule, but %s", rule.Name())
				}
				if rule.Enabled() != true {
					return false, fmt.Errorf("rule.Enabled() should be true, but %t", rule.Enabled())
				}
				if rule.Severity() != tflint.ERROR {
					return false, fmt.Errorf("rule.Severity() should be ERROR, but %s", rule.Severity())
				}
				if rule.Link() != "https://example.com" {
					return false, fmt.Errorf("rule.Link() should be https://example.com, but %s", rule.Link())
				}
				if message != "this is test" {
					return false, fmt.Errorf("message should be `this is test`, but %s", message)
				}
				want := hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 2},
					End:      hcl.Pos{Line: 2, Column: 10},
				}
				if diff := cmp.Diff(location, want); diff != "" {
					return false, fmt.Errorf("diff: %s", diff)
				}
				if fixable != true {
					return false, fmt.Errorf("fixable should be true, but %t", fixable)
				}
				return true, nil
			},
			ErrCheck: neverHappend,
			Changes: map[string]string{
				"test.tf": `foo = "baz"`,
			},
		},
		{
			Name: "child modules",
			Args: func() (tflint.Rule, string, hcl.Range, func(tflint.Fixer) error) {
				return &Rule{},
					"this is test",
					hcl.Range{Filename: "test.tf", Start: hcl.Pos{Line: 2, Column: 2}, End: hcl.Pos{Line: 2, Column: 10}},
					func(f tflint.Fixer) error {
						return f.ReplaceText(
							hcl.Range{Filename: "test.tf", Start: hcl.Pos{Byte: 7}, End: hcl.Pos{Byte: 10}},
							"baz",
						)
					}
			},
			ServerImpl: func(rule tflint.Rule, message string, location hcl.Range, fixable bool) (bool, error) {
				if fixable != false {
					return false, fmt.Errorf("fixable should be false, but %t", fixable)
				}
				return true, nil
			},
			ModulePath: []string{"module", "child"},
			ErrCheck:   neverHappend,
			Changes:    map[string]string{},
		},
		{
			Name: "autofix is not supported",
			Args: func() (tflint.Rule, string, hcl.Range, func(tflint.Fixer) error) {
				return &Rule{},
					"this is test",
					hcl.Range{Filename: "test.tf", Start: hcl.Pos{Line: 2, Column: 2}, End: hcl.Pos{Line: 2, Column: 10}},
					func(f tflint.Fixer) error {
						if err := f.ReplaceText(
							hcl.Range{Filename: "test.tf", Start: hcl.Pos{Byte: 7}, End: hcl.Pos{Byte: 10}},
							"baz",
						); err != nil {
							return err
						}
						return tflint.ErrFixNotSupported
					}
			},
			ServerImpl: func(rule tflint.Rule, message string, location hcl.Range, fixable bool) (bool, error) {
				if fixable != false {
					return false, fmt.Errorf("fixable should be false, but %t", fixable)
				}
				return true, nil
			},
			ErrCheck: neverHappend,
			Changes:  map[string]string{},
		},
		{
			Name: "fix is disbaled",
			Args: func() (tflint.Rule, string, hcl.Range, func(tflint.Fixer) error) {
				return &Rule{},
					"this is test",
					hcl.Range{Filename: "test.tf", Start: hcl.Pos{Line: 2, Column: 2}, End: hcl.Pos{Line: 2, Column: 10}},
					func(f tflint.Fixer) error {
						return f.ReplaceText(
							hcl.Range{Filename: "test.tf", Start: hcl.Pos{Byte: 7}, End: hcl.Pos{Byte: 10}},
							"baz",
						)
					}
			},
			ServerImpl: func(rule tflint.Rule, message string, location hcl.Range, fixable bool) (bool, error) {
				if fixable != true {
					return false, fmt.Errorf("fixable should be true, but %t", fixable)
				}
				return true, nil
			},
			DisableFix: true,
			ErrCheck:   neverHappend,
			Changes:    map[string]string{},
		},
		{
			Name: "fix is not applied",
			Args: func() (tflint.Rule, string, hcl.Range, func(tflint.Fixer) error) {
				return &Rule{},
					"this is test",
					hcl.Range{Filename: "test.tf", Start: hcl.Pos{Line: 2, Column: 2}, End: hcl.Pos{Line: 2, Column: 10}},
					func(f tflint.Fixer) error {
						return f.ReplaceText(
							hcl.Range{Filename: "test.tf", Start: hcl.Pos{Byte: 7}, End: hcl.Pos{Byte: 10}},
							"baz",
						)
					}
			},
			ServerImpl: func(rule tflint.Rule, message string, location hcl.Range, fixable bool) (bool, error) {
				if fixable != true {
					return false, fmt.Errorf("fixable should be true, but %t", fixable)
				}
				return false, nil
			},
			ErrCheck: neverHappend,
			Changes:  map[string]string{},
		},
		{
			Name: "fix raises an error",
			Args: func() (tflint.Rule, string, hcl.Range, func(tflint.Fixer) error) {
				return &Rule{},
					"this is test",
					hcl.Range{Filename: "test.tf", Start: hcl.Pos{Line: 2, Column: 2}, End: hcl.Pos{Line: 2, Column: 10}},
					func(f tflint.Fixer) error {
						if err := f.ReplaceText(
							hcl.Range{Filename: "test.tf", Start: hcl.Pos{Byte: 7}, End: hcl.Pos{Byte: 10}},
							"baz",
						); err != nil {
							return err
						}
						return errors.New("unexpected error")
					}
			},
			ServerImpl: func(rule tflint.Rule, message string, location hcl.Range, fixable bool) (bool, error) {
				if fixable != true {
					return false, fmt.Errorf("fixable should be true, but %t", fixable)
				}
				return true, nil
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
			Changes: map[string]string{
				"test.tf": `foo = "baz"`,
			},
		},
		{
			Name: "server returns an error",
			Args: func() (tflint.Rule, string, hcl.Range, func(tflint.Fixer) error) {
				return &Rule{},
					"this is test",
					hcl.Range{Filename: "test.tf", Start: hcl.Pos{Line: 2, Column: 2}, End: hcl.Pos{Line: 2, Column: 10}},
					func(f tflint.Fixer) error {
						return f.ReplaceText(
							hcl.Range{Filename: "test.tf", Start: hcl.Pos{Byte: 7}, End: hcl.Pos{Byte: 10}},
							"baz",
						)
					}
			},
			ServerImpl: func(tflint.Rule, string, hcl.Range, bool) (bool, error) {
				return false, errors.New("unexpected error")
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
			Changes: map[string]string{
				"test.tf": `foo = "baz"`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCServer(
				t,
				newMockServer(mockServerImpl{
					getFiles:      getFiles,
					getModulePath: func() []string { return test.ModulePath },
					emitIssue:     test.ServerImpl,
				}),
			)
			client.FixEnabled = !test.DisableFix

			err := client.EmitIssueWithFix(test.Args())
			if test.ErrCheck(err) {
				t.Fatalf("failed to call EmitIssueWithFix: %s", err)
			}

			got := map[string]string{}
			for name, content := range client.Fixer.Changes() {
				got[name] = string(content)
			}
			if diff := cmp.Diff(got, test.Changes); diff != "" {
				t.Fatalf("diff: %s", diff)
			}
		})
	}
}

func TestApplyChanges(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }
	getFiles := func() map[string][]byte {
		return map[string][]byte{
			"test.tf": []byte(`foo = "bar"`),
		}
	}

	tests := []struct {
		name       string
		fix        func(tflint.Fixer) error
		serverImpl func(map[string][]byte) error
		errCheck   func(error) bool
	}{
		{
			name: "apply changes",
			fix: func(f tflint.Fixer) error {
				return f.ReplaceText(
					hcl.Range{Filename: "test.tf", Start: hcl.Pos{Byte: 7}, End: hcl.Pos{Byte: 10}},
					"baz",
				)
			},
			serverImpl: func(files map[string][]byte) error {
				if diff := cmp.Diff(files, map[string][]byte{
					"test.tf": []byte(`foo = "baz"`),
				}); diff != "" {
					return fmt.Errorf("diff: %s", diff)
				}
				return nil
			},
			errCheck: neverHappend,
		},
		{
			name: "server returns an error",
			fix: func(f tflint.Fixer) error {
				return f.ReplaceText(
					hcl.Range{Filename: "test.tf", Start: hcl.Pos{Byte: 7}, End: hcl.Pos{Byte: 10}},
					"baz",
				)
			},
			serverImpl: func(files map[string][]byte) error {
				return errors.New("unexpected error")
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{getFiles: getFiles, applyChanges: test.serverImpl}))
			client.FixEnabled = true

			if err := test.fix(client.Fixer); err != nil {
				t.Fatalf("failed to call fix: %s", err)
			}

			err := client.ApplyChanges()
			if test.errCheck(err) {
				t.Fatalf("failed to call ApplyChanges: %s", err)
			}

			if err == nil && client.Fixer.HasChanges() {
				t.Fatal("fixer should have no changes")
			}
		})
	}
}

func TestEnsureNoError(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		Name     string
		Err      error
		Proc     func() error
		ErrCheck func(error) bool
	}{
		{
			Name: "no errors",
			Err:  nil,
			Proc: func() error {
				return errors.New("should be called")
			},
			ErrCheck: func(err error) bool {
				// should be passed result of proc()
				return err == nil || err.Error() != "should be called"
			},
		},
		{
			Name: "ErrUnevaluable",
			Err:  fmt.Errorf("unevaluable%w", tflint.ErrUnevaluable),
			Proc: func() error {
				return errors.New("should not be called")
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "ErrNullValue",
			Err:  fmt.Errorf("null value%w", tflint.ErrNullValue),
			Proc: func() error {
				return errors.New("should not be called")
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "ErrUnknownValue",
			Err:  fmt.Errorf("unknown value%w", tflint.ErrUnknownValue),
			Proc: func() error {
				return errors.New("should not be called")
			},
			ErrCheck: neverHappend,
		},
		{
			Name: "unexpected error",
			Err:  errors.New("unexpected error"),
			Proc: func() error {
				return errors.New("should not be called")
			},
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unexpected error"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			client := startTestGRPCServer(t, newMockServer(mockServerImpl{}))

			err := client.EnsureNoError(test.Err, test.Proc)
			if test.ErrCheck(err) {
				t.Fatalf("failed to call EnsureNoError: %s", err)
			}
		})
	}
}
