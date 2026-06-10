package runner

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/logger"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// EvalExprFunc evaluates an expression into the target pointer. The target is
// always a pointer here; EvaluateExpr resolves callback functions before
// calling it.
type EvalExprFunc func(expr hcl.Expression, target any, opts *tflint.EvaluateExprOption) error

var errRefTy = reflect.TypeOf((*error)(nil)).Elem()

// EvaluateExpr dispatches a tflint.Runner.EvaluateExpr target, which must be
// a pointer or a callback function, to the passed evaluator. Errors that mean
// the value cannot be represented as a Go value (unknown, null, marked, or
// unevaluable) skip the callback instead of being returned.
func EvaluateExpr(expr hcl.Expression, target any, opts *tflint.EvaluateExprOption, eval EvalExprFunc) error {
	rval := reflect.ValueOf(target)
	rty := rval.Type()

	var callback bool
	switch rty.Kind() {
	case reflect.Func:
		// Callback must meet the following requirements:
		//   - It must be a function
		//   - It must take an argument
		//   - It must return an error
		if !(rty.NumIn() == 1 && rty.NumOut() == 1 && rty.Out(0).Implements(errRefTy)) {
			panic(`callback must be of type "func (v T) error"`)
		}
		callback = true
		target = reflect.New(rty.In(0)).Interface()

	case reflect.Pointer:
		// ok
	default:
		panic("target value is not a pointer or function")
	}

	err := eval(expr, target, opts)
	if !callback {
		// error should be handled in the caller
		return err
	}

	if err != nil {
		// If it cannot be represented as a Go value, exit without invoking the callback rather than returning an error.
		if errors.Is(err, tflint.ErrUnknownValue) ||
			errors.Is(err, tflint.ErrNullValue) ||
			errors.Is(err, tflint.ErrSensitive) ||
			errors.Is(err, tflint.ErrEphemeral) ||
			errors.Is(err, tflint.ErrUnevaluable) {
			return nil
		}
		return err
	}

	rerr := rval.Call([]reflect.Value{reflect.ValueOf(target).Elem()})
	if rerr[0].IsNil() {
		return nil
	}
	return rerr[0].Interface().(error)
}

// WantType returns the type the evaluated value should be converted into for
// the given target pointer. The opts.WantType option takes precedence over
// inference. Unsupported target types panic, per the documented
// tflint.Runner.EvaluateExpr contract.
func WantType(target any, opts *tflint.EvaluateExprOption) cty.Type {
	if opts != nil && opts.WantType != nil {
		return *opts.WantType
	}

	switch target.(type) {
	case *string:
		return cty.String
	case *int:
		return cty.Number
	case *bool:
		return cty.Bool
	case *[]string:
		return cty.List(cty.String)
	case *[]int:
		return cty.List(cty.Number)
	case *[]bool:
		return cty.List(cty.Bool)
	case *map[string]string:
		return cty.Map(cty.String)
	case *map[string]int:
		return cty.Map(cty.Number)
	case *map[string]bool:
		return cty.Map(cty.Bool)
	case *cty.Value:
		return cty.DynamicPseudoType
	default:
		panic(fmt.Sprintf("unsupported target type: %T", target))
	}
}

// DecodeValue decodes an evaluated value into the target pointer. Unless ty
// is cty.DynamicPseudoType, values that cannot be represented as a Go value
// return sentinel errors: tflint.ErrUnknownValue, tflint.ErrNullValue,
// tflint.ErrSensitive, and tflint.ErrEphemeral. The range only appears in
// debug logs.
func DecodeValue(val cty.Value, ty cty.Type, rng hcl.Range, target any) error {
	if ty == cty.DynamicPseudoType {
		return gocty.FromCtyValue(val, target)
	}

	// Returns an error if the value cannot be decoded to a Go value (e.g. unknown, null, marked).
	// This allows the caller to handle the value by the errors package.
	err := cty.Walk(val, func(path cty.Path, v cty.Value) (bool, error) {
		if !v.IsKnown() {
			logger.Debug(fmt.Sprintf("unknown value found in %s", rng))
			return false, tflint.ErrUnknownValue
		}
		if v.IsNull() {
			logger.Debug(fmt.Sprintf("null value found in %s", rng))
			return false, tflint.ErrNullValue
		}
		if v.HasMark(marks.Sensitive) {
			logger.Debug(fmt.Sprintf("sensitive value found in %s", rng))
			return false, tflint.ErrSensitive
		}
		if v.HasMark(marks.Ephemeral) {
			logger.Debug(fmt.Sprintf("ephemeral value found in %s", rng))
			return false, tflint.ErrEphemeral
		}
		return true, nil
	})
	if err != nil {
		return err
	}

	return gocty.FromCtyValue(val, target)
}

// EnsureNoError runs proc only if err is nil, filtering generally-ignorable
// evaluation errors. It deliberately does not filter tflint.ErrEphemeral,
// preserving the historical behavior of the deprecated
// tflint.Runner.EnsureNoError.
func EnsureNoError(err error, proc func() error) error {
	if err == nil {
		return proc()
	}

	if errors.Is(err, tflint.ErrUnevaluable) || errors.Is(err, tflint.ErrNullValue) || errors.Is(err, tflint.ErrUnknownValue) || errors.Is(err, tflint.ErrSensitive) {
		return nil
	}
	return err
}
