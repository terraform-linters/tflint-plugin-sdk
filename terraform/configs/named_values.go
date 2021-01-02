package configs

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// Variable is an alternative representation of configs.Variable.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/named_values.go#L21-L34
type Variable struct {
	Name        string
	Description string
	Default     cty.Value
	Type        cty.Type
	ParsingMode VariableParsingMode
	Validations []*VariableValidation
	Sensitive   bool

	DescriptionSet bool
	SensitiveSet   bool

	DeclRange hcl.Range
}

// VariableParsingMode is an alternative representation of configs.VariableParsingMode.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/named_values.go#L234-L242
type VariableParsingMode rune

// VariableParseLiteral is a variable parsing mode that just takes the given
// string directly as a cty.String value.
const VariableParseLiteral VariableParsingMode = 'L'

// VariableParseHCL is a variable parsing mode that attempts to parse the given
// string as an HCL expression and returns the result.
const VariableParseHCL VariableParsingMode = 'H'

// VariableValidation is an alternative representation of configs.VariableValidation.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/named_values.go#L281-L297
type VariableValidation struct {
	// Condition is an expression that refers to the variable being tested
	// and contains no other references. The expression must return true
	// to indicate that the value is valid or false to indicate that it is
	// invalid. If the expression produces an error, that's considered a bug
	// in the module defining the validation rule, not an error in the caller.
	Condition hcl.Expression

	// ErrorMessage is one or more full sentences, which would need to be in
	// English for consistency with the rest of the error message output but
	// can in practice be in any language as long as it ends with a period.
	// The message should describe what is required for the condition to return
	// true in a way that would make sense to a caller of the module.
	ErrorMessage string

	DeclRange hcl.Range
}

// Local is an alternative representation of configs.Local.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/named_values.go#L501-L506
type Local struct {
	Name string
	Expr hcl.Expression

	DeclRange hcl.Range
}

// Output is an alternative representation of configs.Output.
// https://github.com/hashicorp/terraform/blob/v0.14.3/configs/named_values.go#L430-L441
type Output struct {
	Name        string
	Description string
	Expr        hcl.Expression
	// DependsOn   []hcl.Traversal
	Sensitive bool

	DescriptionSet bool
	SensitiveSet   bool

	DeclRange hcl.Range
}
