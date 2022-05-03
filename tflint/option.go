package tflint

import "github.com/zclconf/go-cty/cty"

// ModuleCtxType represents target module.
//go:generate stringer -type=ModuleCtxType
type ModuleCtxType int32

const (
	// SelfModuleCtxType targets the current module. The default is this behavior.
	SelfModuleCtxType ModuleCtxType = iota
	// RootModuleCtxType targets the root module. This is useful when you want to refer to a provider config.
	RootModuleCtxType
)

// GetModuleContentOption is an option that controls the behavior when getting a module content.
type GetModuleContentOption struct {
	// Specify the module to be acquired.
	ModuleCtx ModuleCtxType
	// Whether it includes resources that are not created, for example because count is 0 or unknown.
	IncludeNotCreated bool
	// Hint is info for optimizing a query. This is an advanced option and it is not intended to be used directly from plugins.
	Hint GetModuleContentHint
}

// GetModuleContentHint is info for optimizing a query. This is an advanced option and it is not intended to be used directly from plugins.
type GetModuleContentHint struct {
	ResourceType string
}

// EvaluateExprOption is an option that controls the behavior when evaluating an expression.
type EvaluateExprOption struct {
	// Specify what type of value is expected.
	WantType *cty.Type
	// Set the scope of the module to evaluate.
	ModuleCtx ModuleCtxType
}
