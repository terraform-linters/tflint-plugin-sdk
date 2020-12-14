package client

import (
	"errors"
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

type mockServer struct {
	Listener *net.TCPListener
}

func (*mockServer) Attributes(req *AttributesRequest, resp *AttributesResponse) error {
	*resp = AttributesResponse{Attributes: []*Attribute{
		{
			Name: req.AttributeName,
			Expr: []byte("1"),
			ExprRange: hcl.Range{
				Filename: "example.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 2},
			},
			Range: hcl.Range{
				Start: hcl.Pos{Line: 1, Column: 1},
				End:   hcl.Pos{Line: 2, Column: 2},
			},
		},
	}, Err: nil}
	return nil
}

func (*mockServer) Blocks(req *BlocksRequest, resp *BlocksResponse) error {
	*resp = BlocksResponse{Blocks: []*Block{
		{
			Type:      "resource",
			Labels:    []string{"aws_instance", "foo"},
			Body:      []byte(`instance_type = "t2.micro"`),
			BodyRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 3}, End: hcl.Pos{Line: 2, Column: 28}},
			DefRange:  hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 1}, End: hcl.Pos{Line: 1, Column: 29}},
			TypeRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 1}, End: hcl.Pos{Line: 1, Column: 8}},
			LabelRanges: []hcl.Range{
				{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 10}, End: hcl.Pos{Line: 3, Column: 23}},
				{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 25}, End: hcl.Pos{Line: 3, Column: 29}},
			},
		},
	}, Err: nil}
	return nil
}

func (*mockServer) Resources(req *ResourcesRequest, resp *ResourcesResponse) error {
	*resp = ResourcesResponse{Resources: []*Resource{
		{
			Mode:              addrs.ManagedResourceMode,
			Name:              "example",
			Type:              "resource",
			Config:            []byte(`instance_type = "t2.micro"`),
			ConfigRange:       hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 3}, End: hcl.Pos{Line: 2, Column: 28}},
			Count:             nil,
			ForEach:           nil,
			ProviderConfigRef: nil,
			Provider:          addrs.Provider{Type: "aws", Namespace: "hashicorp", Hostname: "registry.terraform.io"},
			Managed: &ManagedResource{
				Connection:             nil,
				Provisioners:           []*Provisioner{},
				CreateBeforeDestroy:    false,
				PreventDestroy:         false,
				IgnoreAllChanges:       false,
				CreateBeforeDestroySet: false,
				PreventDestroySet:      false,
			},
			DeclRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 1}, End: hcl.Pos{Line: 1, Column: 29}},
			TypeRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 1}, End: hcl.Pos{Line: 1, Column: 8}},
		},
	}, Err: nil}
	return nil
}

func (*mockServer) Backend(req *BackendRequest, resp *BackendResponse) error {
	*resp = BackendResponse{Backend: &Backend{
		Type:        "example",
		Config:      []byte(`storage = "cloud"`),
		ConfigRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 3, Column: 5}, End: hcl.Pos{Line: 3, Column: 22}},
		DeclRange:   hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 3}, End: hcl.Pos{Line: 2, Column: 22}},
		TypeRange:   hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 11}, End: hcl.Pos{Line: 2, Column: 19}},
	}, Err: nil}
	return nil
}

func (*mockServer) EvalExpr(req *EvalExprRequest, resp *EvalExprResponse) error {
	*resp = EvalExprResponse{Val: cty.StringVal("1"), Err: nil}
	return nil
}

func (s *mockServer) EmitIssue(req *EmitIssueRequest, resp *interface{}) error {
	return nil
}

func startMockServer(t *testing.T) (*Client, *mockServer) {
	addy, err := net.ResolveTCPAddr("tcp", "0.0.0.0:42586")
	if err != nil {
		t.Fatal(err)
	}
	inbound, err := net.ListenTCP("tcp", addy)
	if err != nil {
		t.Fatal(err)
	}

	server := &mockServer{Listener: inbound}
	rpc.RegisterName("Plugin", server)
	go rpc.Accept(inbound)

	conn, err := net.Dial("tcp", "0.0.0.0:42586")
	if err != nil {
		t.Fatal(err)
	}
	return NewClient(conn), server
}

func Test_WalkResourceAttributes(t *testing.T) {
	client, server := startMockServer(t)
	defer server.Listener.Close()

	walked := []*hcl.Attribute{}
	walker := func(attribute *hcl.Attribute) error {
		walked = append(walked, attribute)
		return nil
	}

	if err := client.WalkResourceAttributes("foo", "bar", walker); err != nil {
		t.Fatal(err)
	}

	expr, diags := hclsyntax.ParseExpression([]byte("1"), "example.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	expected := []*hcl.Attribute{
		{
			Name: "bar",
			Expr: expr,
			Range: hcl.Range{
				Start: hcl.Pos{Line: 1, Column: 1},
				End:   hcl.Pos{Line: 2, Column: 2},
			},
		},
	}

	opt := cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val")
	if !cmp.Equal(expected, walked, opt) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, walked, opt))
	}
}

func Test_Backend(t *testing.T) {
	client, server := startMockServer(t)
	defer server.Listener.Close()

	expected := &configs.Backend{
		Type: "example",
		Config: &hclsyntax.Body{
			Attributes: hclsyntax.Attributes{
				"storage": {
					Name: "storage",
					Expr: &hclsyntax.TemplateExpr{
						Parts: []hclsyntax.Expression{
							&hclsyntax.LiteralValueExpr{
								SrcRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 3, Column: 16}, End: hcl.Pos{Line: 3, Column: 21}},
							},
						},
						SrcRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 3, Column: 15}, End: hcl.Pos{Line: 3, Column: 22}},
					},
					SrcRange:    hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 3, Column: 5}, End: hcl.Pos{Line: 3, Column: 22}},
					NameRange:   hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 3, Column: 5}, End: hcl.Pos{Line: 3, Column: 12}},
					EqualsRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 3, Column: 13}, End: hcl.Pos{Line: 3, Column: 14}},
				},
			},
			Blocks:   hclsyntax.Blocks{},
			SrcRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 3, Column: 5}, End: hcl.Pos{Line: 3, Column: 22}},
			EndRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 3, Column: 22}, End: hcl.Pos{Line: 3, Column: 22}},
		},
		DeclRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 3}, End: hcl.Pos{Line: 2, Column: 22}},
		TypeRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 11}, End: hcl.Pos{Line: 2, Column: 19}},
	}

	backend, err := client.Backend()
	if err != nil {
		t.Fatal(err)
	}

	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(hclsyntax.Body{}),
		cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val"),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
	}
	if !cmp.Equal(expected, backend, opts...) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, backend, opts...))
	}
}

func Test_WalkResourceBlocks(t *testing.T) {
	client, server := startMockServer(t)
	defer server.Listener.Close()

	walked := []*hcl.Block{}
	walker := func(block *hcl.Block) error {
		walked = append(walked, block)
		return nil
	}

	if err := client.WalkResourceBlocks("foo", "bar", walker); err != nil {
		t.Fatal(err)
	}

	expected := []*hcl.Block{
		{
			Type:   "resource",
			Labels: []string{"aws_instance", "foo"},
			Body: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"instance_type": {
						Name: "instance_type",
						Expr: &hclsyntax.TemplateExpr{
							Parts: []hclsyntax.Expression{
								&hclsyntax.LiteralValueExpr{
									SrcRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 20}, End: hcl.Pos{Line: 2, Column: 28}},
								},
							},
							SrcRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 19}, End: hcl.Pos{Line: 2, Column: 29}},
						},
						SrcRange:    hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 3}, End: hcl.Pos{Line: 2, Column: 29}},
						NameRange:   hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 3}, End: hcl.Pos{Line: 2, Column: 16}},
						EqualsRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 17}, End: hcl.Pos{Line: 2, Column: 18}},
					},
				},
				Blocks:   hclsyntax.Blocks{},
				SrcRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 3}, End: hcl.Pos{Line: 2, Column: 29}},
				EndRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 29}, End: hcl.Pos{Line: 2, Column: 29}},
			},
			DefRange:  hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 1}, End: hcl.Pos{Line: 1, Column: 29}},
			TypeRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 1}, End: hcl.Pos{Line: 1, Column: 8}},
			LabelRanges: []hcl.Range{
				{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 10}, End: hcl.Pos{Line: 3, Column: 23}},
				{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 25}, End: hcl.Pos{Line: 3, Column: 29}},
			},
		},
	}

	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(hclsyntax.Body{}),
		cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val"),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
	}
	if !cmp.Equal(expected, walked, opts...) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, walked, opts...))
	}
}

func Test_WalkResources(t *testing.T) {
	client, server := startMockServer(t)
	defer server.Listener.Close()

	walked := []*configs.Resource{}
	walker := func(block *configs.Resource) error {
		walked = append(walked, block)
		return nil
	}

	if err := client.WalkResources("example", walker); err != nil {
		t.Fatal(err)
	}

	expected := []*configs.Resource{
		{
			Mode: addrs.ManagedResourceMode,
			Name: "example",
			Type: "resource",
			Config: &hclsyntax.Body{
				Attributes: hclsyntax.Attributes{
					"instance_type": {
						Name: "instance_type",
						Expr: &hclsyntax.TemplateExpr{
							Parts: []hclsyntax.Expression{
								&hclsyntax.LiteralValueExpr{
									SrcRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 20}, End: hcl.Pos{Line: 2, Column: 28}},
								},
							},
							SrcRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 19}, End: hcl.Pos{Line: 2, Column: 29}},
						},
						SrcRange:    hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 3}, End: hcl.Pos{Line: 2, Column: 29}},
						NameRange:   hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 3}, End: hcl.Pos{Line: 2, Column: 16}},
						EqualsRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 17}, End: hcl.Pos{Line: 2, Column: 18}},
					},
				},
				Blocks:   hclsyntax.Blocks{},
				SrcRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 3}, End: hcl.Pos{Line: 2, Column: 29}},
				EndRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 2, Column: 29}, End: hcl.Pos{Line: 2, Column: 29}},
			},
			Count:             nil,
			ForEach:           nil,
			ProviderConfigRef: nil,
			Provider:          addrs.Provider{Type: "aws", Namespace: "hashicorp", Hostname: "registry.terraform.io"},
			Managed: &configs.ManagedResource{
				Connection:             nil,
				Provisioners:           []*configs.Provisioner{},
				CreateBeforeDestroy:    false,
				PreventDestroy:         false,
				IgnoreAllChanges:       false,
				CreateBeforeDestroySet: false,
				PreventDestroySet:      false,
			},
			DeclRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 1}, End: hcl.Pos{Line: 1, Column: 29}},
			TypeRange: hcl.Range{Filename: "example.tf", Start: hcl.Pos{Line: 1, Column: 1}, End: hcl.Pos{Line: 1, Column: 8}},
		},
	}

	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(hclsyntax.Body{}),
		cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val"),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
	}
	if !cmp.Equal(expected, walked, opts...) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, walked, opts...))
	}
}

func Test_EvaluateExpr(t *testing.T) {
	client, server := startMockServer(t)
	defer server.Listener.Close()

	file, err := ioutil.TempFile("", "tflint-test-evaluateExpr-*.tf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	if _, err := file.Write([]byte("1")); err != nil {
		t.Fatal(err)
	}

	expr, diags := hclsyntax.ParseExpression([]byte("1"), file.Name(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	var ret string
	if err := client.EvaluateExpr(expr, &ret, nil); err != nil {
		t.Fatal(err)
	}

	if ret != "1" {
		t.Fatalf("Expected: 1, but got %s", ret)
	}
}

type testRule struct{}

func (*testRule) Name() string              { return "test" }
func (*testRule) Enabled() bool             { return true }
func (*testRule) Severity() string          { return "Error" }
func (*testRule) Link() string              { return "" }
func (*testRule) Check(tflint.Runner) error { return nil }

func Test_EmitIssueOnExpr(t *testing.T) {
	client, server := startMockServer(t)
	defer server.Listener.Close()

	file, err := ioutil.TempFile("", "tflint-test-evaluateExpr-*.tf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	if _, err := file.Write([]byte("1")); err != nil {
		t.Fatal(err)
	}

	expr, diags := hclsyntax.ParseExpression([]byte("1"), file.Name(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	if err := client.EmitIssueOnExpr(&testRule{}, file.Name(), expr); err != nil {
		t.Fatal(err)
	}
}

func Test_EmitIssue(t *testing.T) {
	client, server := startMockServer(t)
	defer server.Listener.Close()

	file, err := ioutil.TempFile("", "tflint-test-evaluateExpr-*.tf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	if _, err := file.Write([]byte("1")); err != nil {
		t.Fatal(err)
	}

	expr, diags := hclsyntax.ParseExpression([]byte("1"), file.Name(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	if err := client.EmitIssue(&testRule{}, file.Name(), expr.Range()); err != nil {
		t.Fatal(err)
	}
}

func Test_EnsureNoError(t *testing.T) {
	cases := []struct {
		Name      string
		Error     error
		ErrorText string
	}{
		{
			Name:      "no error",
			Error:     nil,
			ErrorText: "function called",
		},
		{
			Name:      "native error",
			Error:     errors.New("Error occurred"),
			ErrorText: "Error occurred",
		},
		{
			Name: "warning error",
			Error: tflint.Error{
				Code:    tflint.UnknownValueError,
				Level:   tflint.WarningLevel,
				Message: "Warning error",
			},
		},
		{
			Name: "app error",
			Error: tflint.Error{
				Code:    tflint.TypeMismatchError,
				Level:   tflint.ErrorLevel,
				Message: "App error",
			},
			ErrorText: "App error",
		},
	}

	client, _ := startMockServer(t)

	for _, tc := range cases {
		err := client.EnsureNoError(tc.Error, func() error {
			return errors.New("function called")
		})
		if err == nil {
			if tc.ErrorText != "" {
				t.Fatalf("Failed `%s` test: expected error is not occurred `%s`", tc.Name, tc.ErrorText)
			}
		} else if err.Error() != tc.ErrorText {
			t.Fatalf("Failed `%s` test: expected error is %s, but get %s", tc.Name, tc.ErrorText, err)
		}
	}
}
