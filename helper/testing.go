package helper

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TestRunner returns a mock Runner for testing.
// You can pass the map of file names and their contents in the second argument.
func TestRunner(t *testing.T, files map[string]string) *Runner {
	t.Helper()

	runner := newLocalRunner(map[string]*hcl.File{}, Issues{})
	parser := hclparse.NewParser()

	for name, src := range files {
		var file *hcl.File
		var diags hcl.Diagnostics
		if strings.HasSuffix(name, ".json") {
			file, diags = parser.ParseJSON([]byte(src), name)
		} else {
			file, diags = parser.ParseHCL([]byte(src), name)
		}
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		if name == ".tflint.hcl" {
			var config Config
			if diags := gohcl.DecodeBody(file.Body, nil, &config); diags.HasErrors() {
				t.Fatal(diags)
			}
			runner.config = config
		} else {
			runner.addLocalFile(name, file)
		}
	}

	if err := runner.initFromFiles(); err != nil {
		panic(fmt.Sprintf("Failed to initialize runner: %s", err))
	}
	return runner
}

// AssertIssues is an assertion helper for comparing issues.
func AssertIssues(t *testing.T, want Issues, got Issues) {
	t.Helper()

	opts := []cmp.Option{
		// Byte field will be ignored because it's not important in tests such as positions
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		// Issues will be sorted and output in the end, so ignore the order.
		ignoreIssuesOrder(),
		ruleComparer(),
	}
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Fatalf("Expected issues are not matched:\n %s\n", diff)
	}
}

// AssertIssuesWithoutRange is an assertion helper for comparing issues except for range.
func AssertIssuesWithoutRange(t *testing.T, want Issues, got Issues) {
	t.Helper()

	opts := []cmp.Option{
		cmpopts.IgnoreFields(Issue{}, "Range"),
		ignoreIssuesOrder(),
		ruleComparer(),
	}
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Fatalf("Expected issues are not matched:\n %s\n", diff)
	}
}

// AssertChanges is an assertion helper for comparing autofix changes.
func AssertChanges(t *testing.T, want map[string]string, got map[string][]byte) {
	t.Helper()

	sources := make(map[string]string)
	for name, src := range got {
		sources[name] = string(src)
	}
	if diff := cmp.Diff(want, sources); diff != "" {
		t.Fatalf("Expected changes are not matched:\n %s\n", diff)
	}
}

// ruleComparer returns a Comparer func that checks that two rule interfaces
// have the same underlying type. It does not compare struct fields.
func ruleComparer() cmp.Option {
	return cmp.Comparer(func(x, y tflint.Rule) bool {
		return reflect.TypeOf(x) == reflect.TypeOf(y)
	})
}

func ignoreIssuesOrder() cmp.Option {
	return cmpopts.SortSlices(func(i, j *Issue) bool {
		if i.Range.Filename != j.Range.Filename {
			return i.Range.Filename < j.Range.Filename
		}
		if i.Range.Start.Line != j.Range.Start.Line {
			return i.Range.Start.Line < j.Range.Start.Line
		}
		if i.Range.Start.Column != j.Range.Start.Column {
			return i.Range.Start.Column < j.Range.Start.Column
		}
		if i.Range.End.Line != j.Range.End.Line {
			return i.Range.End.Line > j.Range.End.Line
		}
		if i.Range.End.Column != j.Range.End.Column {
			return i.Range.End.Column > j.Range.End.Column
		}
		return i.Message < j.Message
	})
}
