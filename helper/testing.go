package helper

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// TestRunner returns a pseudo Runner for testing
func TestRunner(t *testing.T, files map[string]string) *Runner {
	runner := &Runner{Files: map[string]*hcl.File{}}
	parser := hclparse.NewParser()

	for name, src := range files {
		file, diags := parser.ParseHCL([]byte(src), name)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		runner.Files[name] = file
	}

	return runner
}

// AssertIssues is an assertion helper for comparing issues
func AssertIssues(t *testing.T, expected Issues, actual Issues) {
	opts := []cmp.Option{
		// Byte field will be ignored because it's not important in tests such as positions
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		cmpopts.IgnoreFields(Issue{}, "Rule"),
	}
	if !cmp.Equal(expected, actual, opts...) {
		t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(expected, actual, opts...))
	}
}

// AssertIssuesWithoutRange is an assertion helper for comparing issues
func AssertIssuesWithoutRange(t *testing.T, expected Issues, actual Issues) {
	opts := []cmp.Option{
		cmpopts.IgnoreFields(Issue{}, "Range"),
		cmpopts.IgnoreFields(Issue{}, "Rule"),
	}
	if !cmp.Equal(expected, actual, opts...) {
		t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(expected, actual, opts...))
	}
}
