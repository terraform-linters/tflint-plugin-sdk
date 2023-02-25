package tflint

import (
	"errors"
)

// List of errors returned by TFLint.
// It's possible to get this error from a plugin, but the error handling is hidden
// inside the plugin system, so you usually don't have to worry about it.
var (
	// ErrUnknownValue is an error when an unknown value is referenced
	ErrUnknownValue = errors.New("unknown value found")
	// ErrNullValue is an error when null value is referenced
	ErrNullValue = errors.New("null value found")
	// ErrUnevaluable is an error when a received expression has unevaluable references.
	// Deprecated: This error is no longer returned since TFLint v0.41.
	ErrUnevaluable = errors.New("")
	// ErrSensitive is an error when a received expression contains a sensitive value.
	ErrSensitive = errors.New("")
)
