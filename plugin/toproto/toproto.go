package toproto

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BodySchema converts schema.BodySchema to proto.BodySchema
func BodySchema(body *hclext.BodySchema) *proto.BodySchema {
	if body == nil {
		return &proto.BodySchema{}
	}

	attributes := make([]*proto.BodySchema_Attribute, len(body.Attributes))
	for idx, attr := range body.Attributes {
		attributes[idx] = &proto.BodySchema_Attribute{Name: attr.Name, Required: attr.Required}
	}

	blocks := make([]*proto.BodySchema_Block, len(body.Blocks))
	for idx, block := range body.Blocks {
		blocks[idx] = &proto.BodySchema_Block{
			Type:       block.Type,
			LabelNames: block.LabelNames,
			Body:       BodySchema(block.Body),
		}
	}

	return &proto.BodySchema{
		Attributes: attributes,
		Blocks:     blocks,
	}
}

// BodyContent converts schema.BodyContent to proto.BodyContent
func BodyContent(body *hclext.BodyContent, sources map[string][]byte) *proto.BodyContent {
	if body == nil {
		return &proto.BodyContent{}
	}

	attributes := map[string]*proto.BodyContent_Attribute{}
	for idx, attr := range body.Attributes {
		attributes[idx] = &proto.BodyContent_Attribute{
			Name:      attr.Name,
			Expr:      attr.Expr.Range().SliceBytes(sources[attr.Range.Filename]),
			Range:     Range(attr.Range),
			NameRange: Range(attr.NameRange),
			ExprRange: Range(attr.Expr.Range()),
		}
	}

	blocks := make([]*proto.BodyContent_Block, len(body.Blocks))
	for idx, block := range body.Blocks {
		labelRanges := make([]*proto.Range, len(block.LabelRanges))
		for idx, labelRange := range block.LabelRanges {
			labelRanges[idx] = Range(labelRange)
		}

		blocks[idx] = &proto.BodyContent_Block{
			Type:        block.Type,
			Labels:      block.Labels,
			Body:        BodyContent(block.Body, sources),
			DefRange:    Range(block.DefRange),
			TypeRange:   Range(block.TypeRange),
			LabelRanges: labelRanges,
		}
	}

	return &proto.BodyContent{
		Attributes: attributes,
		Blocks:     blocks,
	}
}

// Rule converts tflint.Rule to proto.EmitIssue_Rule
func Rule(rule tflint.Rule) *proto.EmitIssue_Rule {
	return &proto.EmitIssue_Rule{
		Name:     rule.Name(),
		Enabled:  rule.Enabled(),
		Severity: Severity(rule.Severity()),
		Link:     rule.Link(),
	}
}

// Severity converts severity to proto.EmitIssue_Severity
func Severity(severity string) proto.EmitIssue_Severity {
	switch severity {
	case tflint.ERROR:
		return proto.EmitIssue_SEVERITY_ERROR
	case tflint.WARNING:
		return proto.EmitIssue_SEVERITY_WARNING
	case tflint.NOTICE:
		return proto.EmitIssue_SEVERITY_NOTICE
	}

	return proto.EmitIssue_SEVERITY_ERROR
}

// Range converts hcl.Range to proto.Range
func Range(rng hcl.Range) *proto.Range {
	return &proto.Range{
		Filename: rng.Filename,
		Start:    Pos(rng.Start),
		End:      Pos(rng.End),
	}
}

// Pos converts hcl.Pos to proto.Range_Pos
func Pos(pos hcl.Pos) *proto.Range_Pos {
	return &proto.Range_Pos{
		Line:   int64(pos.Line),
		Column: int64(pos.Column),
		Byte:   int64(pos.Byte),
	}
}

// Config converts tflint.Config to proto.ApplyGlobalConfig_Config
func Config(config *tflint.Config) *proto.ApplyGlobalConfig_Config {
	rules := map[string]*proto.ApplyGlobalConfig_RuleConfig{}
	for name, rule := range config.Rules {
		rules[name] = &proto.ApplyGlobalConfig_RuleConfig{Name: rule.Name, Enabled: rule.Enabled}
	}
	return &proto.ApplyGlobalConfig_Config{Rules: rules, DisabledByDefault: config.DisabledByDefault}
}

// ModuleCtxType converts tflint.ModuleCtxType to proto.ModuleCtxType
func ModuleCtxType(ty tflint.ModuleCtxType) proto.ModuleCtxType {
	switch ty {
	case tflint.SelfModuleCtxType:
		return proto.ModuleCtxType_MODULE_CTX_TYPE_SELF
	case tflint.RootModuleCtxType:
		return proto.ModuleCtxType_MODULE_CTX_TYPE_ROOT
	default:
		panic(fmt.Sprintf("invalid ModuleCtxType: %s", ty.String()))
	}
}

// Error converts tflint.Error to gRPC error status with details
func Error(code codes.Code, raw error) error {
	appErr, ok := raw.(tflint.Error)
	if !ok {
		return status.Error(code, raw.Error())
	}

	var errCode proto.ErrorCode
	switch appErr.Code {
	case tflint.EvaluationError:
		errCode = proto.ErrorCode_ERROR_CODE_FAILED_TO_EVAL
	case tflint.UnknownValueError:
		errCode = proto.ErrorCode_ERROR_CODE_UNKNOWN_VALUE
	case tflint.NullValueError:
		errCode = proto.ErrorCode_ERROR_CODE_NULL_VALUE
	case tflint.TypeConversionError:
		errCode = proto.ErrorCode_ERROR_CODE_TYPE_CONVERSION
	case tflint.TypeMismatchError:
		errCode = proto.ErrorCode_ERROR_CODE_TYPE_MISMATCH
	case tflint.UnevaluableError:
		errCode = proto.ErrorCode_ERROR_CODE_UNEVALUABLE
	case tflint.UnexpectedAttributeError:
		errCode = proto.ErrorCode_ERROR_CODE_UNEXPECTED_ATTRIBUTE
	default:
		return status.Error(code, appErr.Error())
	}

	st := status.New(code, appErr.Error())
	dt, err := st.WithDetails(&proto.ErrorDetail{Code: errCode})
	if err != nil {
		return status.Error(codes.Unknown, fmt.Sprintf("Failed to add ErrorDetail: code=%d error=%s", code, appErr.Error()))
	}

	return dt.Err()
}
