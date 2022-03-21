package tflint

import (
	"errors"
)

// List of errors returned by TFLint.
// It's possible to get this error from a plugin, but the error handling is hidden
// inside the plugin system, so you usually don't have to worry about it.
var (
	// ErrUnknownValue is an error when an unknown value is referenced
	ErrUnknownValue = errors.New("")
	// ErrNullValue is an error when null value is referenced
	ErrNullValue = errors.New("")
	// ErrUnevaluable is an error when a received expression has unevaluable references.
	ErrUnevaluable = errors.New("")
)
