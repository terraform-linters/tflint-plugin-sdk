package tflint

import (
	"encoding/gob"
	"errors"
	"net"
	"net/rpc"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type mockServer struct {
	Listener *net.TCPListener
}

func (*mockServer) Attributes(req *AttributesRequest, resp *AttributesResponse) error {
	expr, diags := hclsyntax.ParseExpression([]byte("1"), "example.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		*resp = AttributesResponse{Attributes: []*hcl.Attribute{}, Err: diags}
		return nil
	}

	*resp = AttributesResponse{Attributes: []*hcl.Attribute{
		{
			Name: req.AttributeName,
			Expr: expr,
			Range: hcl.Range{
				Start: hcl.Pos{Line: 1, Column: 1},
				End:   hcl.Pos{Line: 2, Column: 2},
			},
		},
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
	gob.Register(&hclsyntax.LiteralValueExpr{})

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

func Test_EvaluateExpr(t *testing.T) {
	client, server := startMockServer(t)
	defer server.Listener.Close()

	expr, diags := hclsyntax.ParseExpression([]byte("1"), "example.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	var ret string
	if err := client.EvaluateExpr(expr, &ret); err != nil {
		t.Fatal(err)
	}

	if ret != "1" {
		t.Fatalf("Expected: 1, but got %s", ret)
	}
}

type testRule struct{}

func (*testRule) Name() string       { return "test" }
func (*testRule) Enabled() bool      { return true }
func (*testRule) Severity() string   { return "Error" }
func (*testRule) Link() string       { return "" }
func (*testRule) Check(Runner) error { return nil }

func Test_EmitIssue(t *testing.T) {
	client, server := startMockServer(t)
	defer server.Listener.Close()

	if err := client.EmitIssue(&testRule{}, "test", hcl.Range{}, Metadata{}); err != nil {
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
			Error: Error{
				Code:    UnknownValueError,
				Level:   WarningLevel,
				Message: "Warning error",
			},
		},
		{
			Name: "app error",
			Error: Error{
				Code:    TypeMismatchError,
				Level:   ErrorLevel,
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
