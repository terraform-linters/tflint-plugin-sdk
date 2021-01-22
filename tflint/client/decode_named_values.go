package client

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
	"github.com/zclconf/go-cty/cty/json"
	"github.com/zclconf/go-cty/cty/msgpack"
)

// Variable is an intermediate representation of configs.Variable.
type Variable struct {
	Name        string
	Description string
	Default     []byte
	Type        []byte
	ParsingMode configs.VariableParsingMode
	Validations []*VariableValidation
	Sensitive   bool

	DescriptionSet bool
	SensitiveSet   bool

	DeclRange hcl.Range
}

func decodeVariable(variable *Variable) (*configs.Variable, hcl.Diagnostics) {
	ret := make([]*configs.VariableValidation, len(variable.Validations))
	for i, v := range variable.Validations {
		validation, diags := decodeVariableValidation(v)
		if diags.HasErrors() {
			return nil, diags
		}
		ret[i] = validation
	}

	typeVal, err := json.UnmarshalType(variable.Type)
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "cannot unmarshal type for variable",
				Detail:   fmt.Sprint(err),
				Subject:  &variable.DeclRange,
			},
		}
	}
	defaultVal, err := msgpack.Unmarshal(variable.Default, typeVal)
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "cannot unmarshal variable default value",
				Detail:   fmt.Sprint(err),
				Subject:  &variable.DeclRange,
			},
		}
	}

	return &configs.Variable{
		Name:        variable.Name,
		Description: variable.Description,
		Default:     defaultVal,
		Type:        typeVal,
		ParsingMode: variable.ParsingMode,
		Validations: ret,
		Sensitive:   variable.Sensitive,

		DescriptionSet: variable.DescriptionSet,
		SensitiveSet:   variable.SensitiveSet,

		DeclRange: variable.DeclRange,
	}, nil
}

// VariableValidation is an intermediate representation of configs.VariableValidation.
type VariableValidation struct {
	Condition      []byte
	ConditionRange hcl.Range

	ErrorMessage string

	DeclRange hcl.Range
}

func decodeVariableValidation(validation *VariableValidation) (*configs.VariableValidation, hcl.Diagnostics) {
	expr, diags := parseExpression(validation.Condition, validation.ConditionRange.Filename, validation.ConditionRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &configs.VariableValidation{
		Condition:    expr,
		ErrorMessage: validation.ErrorMessage,
		DeclRange:    validation.DeclRange,
	}, nil
}

// Local is an intermediate representation of configs.Local.
type Local struct {
	Name      string
	Expr      []byte
	ExprRange hcl.Range

	DeclRange hcl.Range
}

func decodeLocal(local *Local) (*configs.Local, hcl.Diagnostics) {
	expr, diags := parseExpression(local.Expr, local.ExprRange.Filename, local.ExprRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &configs.Local{
		Name: local.Name,
		Expr: expr,

		DeclRange: local.DeclRange,
	}, nil
}

// Output is an intermediate representation of configs.Output.
type Output struct {
	Name        string
	Description string
	Expr        []byte
	ExprRange   hcl.Range
	// DependsOn   []hcl.Traversal
	Sensitive bool

	DescriptionSet bool
	SensitiveSet   bool

	DeclRange hcl.Range
}

func decodeOutput(output *Output) (*configs.Output, hcl.Diagnostics) {
	expr, diags := parseExpression(output.Expr, output.ExprRange.Filename, output.ExprRange.Start)
	if diags.HasErrors() {
		return nil, diags
	}

	return &configs.Output{
		Name:        output.Name,
		Description: output.Description,
		Expr:        expr,
		Sensitive:   output.Sensitive,

		DescriptionSet: output.DescriptionSet,
		SensitiveSet:   output.Sensitive,

		DeclRange: output.DeclRange,
	}, nil
}
