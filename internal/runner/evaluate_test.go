package runner

import (
	"errors"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

func TestEvaluateExpr(t *testing.T) {
	for _, tc := range []struct {
		name       string
		evalErr    error
		want       error
		wantCalled bool
	}{
		{
			name:       "value is passed to the callback",
			evalErr:    nil,
			want:       nil,
			wantCalled: true,
		},
		{
			name:    "unknown value skips the callback",
			evalErr: tflint.ErrUnknownValue,
			want:    nil,
		},
		{
			name:    "null value skips the callback",
			evalErr: tflint.ErrNullValue,
			want:    nil,
		},
		{
			name:    "sensitive value skips the callback",
			evalErr: tflint.ErrSensitive,
			want:    nil,
		},
		{
			name:    "ephemeral value skips the callback",
			evalErr: tflint.ErrEphemeral,
			want:    nil,
		},
		{
			name:    "unevaluable value skips the callback",
			evalErr: tflint.ErrUnevaluable,
			want:    nil,
		},
		{
			name:    "other errors are returned",
			evalErr: errors.New("unexpected"),
			want:    errors.New("unexpected"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eval := func(expr hcl.Expression, target any, opts *tflint.EvaluateExprOption) error {
				*(target.(*string)) = "value"
				return tc.evalErr
			}

			var called bool
			err := EvaluateExpr(nil, func(val string) error {
				called = true
				if val != "value" {
					t.Fatalf(`expected "value", but got %q`, val)
				}
				return nil
			}, nil, eval)

			if tc.want == nil {
				if err != nil {
					t.Fatalf("expected no error, but got %v", err)
				}
			} else if err == nil || err.Error() != tc.want.Error() {
				t.Fatalf("expected %v, but got %v", tc.want, err)
			}
			if called != tc.wantCalled {
				t.Fatalf("expected callback invocation to be %t, but got %t", tc.wantCalled, called)
			}
		})
	}
}

func TestEvaluateExpr_pointer(t *testing.T) {
	for _, tc := range []struct {
		name    string
		evalErr error
	}{
		{
			name:    "value is assigned to the target",
			evalErr: nil,
		},
		{
			name:    "sentinel errors are returned to the caller",
			evalErr: tflint.ErrSensitive,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eval := func(expr hcl.Expression, target any, opts *tflint.EvaluateExprOption) error {
				*(target.(*string)) = "value"
				return tc.evalErr
			}

			var got string
			err := EvaluateExpr(nil, &got, nil, eval)
			if !errors.Is(err, tc.evalErr) {
				t.Fatalf("expected %v, but got %v", tc.evalErr, err)
			}
			if got != "value" {
				t.Fatalf(`expected "value", but got %q`, got)
			}
		})
	}
}

func TestEvaluateExpr_callbackError(t *testing.T) {
	eval := func(expr hcl.Expression, target any, opts *tflint.EvaluateExprOption) error {
		return nil
	}

	want := errors.New("callback error")
	err := EvaluateExpr(nil, func(val string) error {
		return want
	}, nil, eval)
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, but got %v", want, err)
	}
}

func TestEvaluateExpr_panics(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target any
		want   string
	}{
		{
			name:   "callback without argument",
			target: func() error { return nil },
			want:   `callback must be of type "func (v T) error"`,
		},
		{
			name:   "callback without error return",
			target: func(val string) string { return val },
			want:   `callback must be of type "func (v T) error"`,
		},
		{
			name:   "target is not a pointer or function",
			target: "value",
			want:   "target value is not a pointer or function",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				got := recover()
				if got != tc.want {
					t.Fatalf("expected panic %q, but got %v", tc.want, got)
				}
			}()

			eval := func(expr hcl.Expression, target any, opts *tflint.EvaluateExprOption) error {
				return nil
			}
			EvaluateExpr(nil, tc.target, nil, eval)
		})
	}
}

func TestWantType(t *testing.T) {
	wantType := cty.Set(cty.String)

	for _, tc := range []struct {
		name   string
		target any
		opts   *tflint.EvaluateExprOption
		want   cty.Type
	}{
		{
			name:   "string",
			target: new(string),
			want:   cty.String,
		},
		{
			name:   "int",
			target: new(int),
			want:   cty.Number,
		},
		{
			name:   "bool",
			target: new(bool),
			want:   cty.Bool,
		},
		{
			name:   "string list",
			target: &[]string{},
			want:   cty.List(cty.String),
		},
		{
			name:   "int list",
			target: &[]int{},
			want:   cty.List(cty.Number),
		},
		{
			name:   "bool list",
			target: &[]bool{},
			want:   cty.List(cty.Bool),
		},
		{
			name:   "string map",
			target: &map[string]string{},
			want:   cty.Map(cty.String),
		},
		{
			name:   "int map",
			target: &map[string]int{},
			want:   cty.Map(cty.Number),
		},
		{
			name:   "bool map",
			target: &map[string]bool{},
			want:   cty.Map(cty.Bool),
		},
		{
			name:   "cty value",
			target: new(cty.Value),
			want:   cty.DynamicPseudoType,
		},
		{
			name:   "want type option takes precedence",
			target: new(string),
			opts:   &tflint.EvaluateExprOption{WantType: &wantType},
			want:   cty.Set(cty.String),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := WantType(tc.target, tc.opts)
			if !got.Equals(tc.want) {
				t.Fatalf("expected %s, but got %s", tc.want.FriendlyName(), got.FriendlyName())
			}
		})
	}
}

func TestWantType_unsupported(t *testing.T) {
	defer func() {
		got := recover()
		if got != "unsupported target type: *float64" {
			t.Fatalf("expected panic, but got %v", got)
		}
	}()

	WantType(new(float64), nil)
}

func TestDecodeValue(t *testing.T) {
	for _, tc := range []struct {
		name string
		val  cty.Value
		want error
	}{
		{
			name: "known value",
			val:  cty.StringVal("value"),
			want: nil,
		},
		{
			name: "unknown value",
			val:  cty.UnknownVal(cty.String),
			want: tflint.ErrUnknownValue,
		},
		{
			name: "null value",
			val:  cty.NullVal(cty.String),
			want: tflint.ErrNullValue,
		},
		{
			name: "sensitive value",
			val:  cty.StringVal("value").Mark(marks.Sensitive),
			want: tflint.ErrSensitive,
		},
		{
			name: "ephemeral value",
			val:  cty.StringVal("value").Mark(marks.Ephemeral),
			want: tflint.ErrEphemeral,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			err := DecodeValue(tc.val, cty.String, hcl.Range{}, &got)
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, but got %v", tc.want, err)
			}
			if tc.want == nil && got != "value" {
				t.Fatalf(`expected "value", but got %q`, got)
			}
		})
	}
}

func TestDecodeValue_nested(t *testing.T) {
	var got []string
	err := DecodeValue(
		cty.ListVal([]cty.Value{cty.StringVal("value"), cty.UnknownVal(cty.String)}),
		cty.List(cty.String),
		hcl.Range{},
		&got,
	)
	if !errors.Is(err, tflint.ErrUnknownValue) {
		t.Fatalf("expected %v, but got %v", tflint.ErrUnknownValue, err)
	}
}

func TestDecodeValue_dynamic(t *testing.T) {
	for _, tc := range []struct {
		name string
		val  cty.Value
	}{
		{
			name: "sensitive value",
			val:  cty.StringVal("value").Mark(marks.Sensitive),
		},
		{
			name: "unknown value",
			val:  cty.UnknownVal(cty.String),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got cty.Value
			if err := DecodeValue(tc.val, cty.DynamicPseudoType, hcl.Range{}, &got); err != nil {
				t.Fatalf("expected no error, but got %v", err)
			}
			if !got.RawEquals(tc.val) {
				t.Fatalf("expected %s, but got %s", tc.val.GoString(), got.GoString())
			}
		})
	}
}

func TestEnsureNoError(t *testing.T) {
	for _, tc := range []struct {
		name    string
		err     error
		want    error
		wantRun bool
	}{
		{
			name:    "no error",
			err:     nil,
			want:    nil,
			wantRun: true,
		},
		{
			name: "unevaluable error",
			err:  tflint.ErrUnevaluable,
			want: nil,
		},
		{
			name: "null value error",
			err:  tflint.ErrNullValue,
			want: nil,
		},
		{
			name: "unknown value error",
			err:  tflint.ErrUnknownValue,
			want: nil,
		},
		{
			name: "sensitive error",
			err:  tflint.ErrSensitive,
			want: nil,
		},
		{
			name: "ephemeral error is not filtered",
			err:  tflint.ErrEphemeral,
			want: tflint.ErrEphemeral,
		},
		{
			name: "other error",
			err:  errors.New("unexpected"),
			want: errors.New("unexpected"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var run bool
			err := EnsureNoError(tc.err, func() error {
				run = true
				return nil
			})

			if tc.want == nil {
				if err != nil {
					t.Fatalf("expected no error, but got %v", err)
				}
			} else if err == nil || err.Error() != tc.want.Error() {
				t.Fatalf("expected %v, but got %v", tc.want, err)
			}
			if run != tc.wantRun {
				t.Fatalf("expected proc run to be %t, but got %t", tc.wantRun, run)
			}
		})
	}
}
