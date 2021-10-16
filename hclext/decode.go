package hclext

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

// DecodeBody is a derivative of gohcl.DecodeBody the receives hclext.BodyContent instead of hcl.Body.
// Since hcl.Body and interfaces are hard to send over a wire protocol, it is needed to support BodyContent.
// This method differs from gohcl.DecodeBody in several ways:
//
// - Does not support decoding to map.
// - Does not support decoding to hcl.Body, hcl.Expression, etc.
// - Does not support `body` and `remain` tag.
func DecodeBody(body *BodyContent, ctx *hcl.EvalContext, val interface{}) hcl.Diagnostics {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("target value must be a pointer, not %s", rv.Type().String()))
	}

	return decodeBody(body, ctx, rv.Elem())
}

func decodeBody(body *BodyContent, ctx *hcl.EvalContext, val reflect.Value) hcl.Diagnostics {
	et := val.Type()
	switch et.Kind() {
	case reflect.Struct:
		return decodeBodyToStruct(body, ctx, val)
	default:
		panic(fmt.Sprintf("target value must be pointer to struct, not %s", et.String()))
	}
}

func decodeBodyToStruct(body *BodyContent, ctx *hcl.EvalContext, val reflect.Value) hcl.Diagnostics {
	var diags hcl.Diagnostics

	tags := getFieldTags(val.Type())

	for name, fieldIdx := range tags.Attributes {
		attr, exists := body.Attributes[name]
		if !exists {
			continue
		}
		diags = diags.Extend(gohcl.DecodeExpression(attr.Expr, ctx, val.Field(fieldIdx).Addr().Interface()))
	}

	blocksByType := body.Blocks.ByType()

	for typeName, fieldIdx := range tags.Blocks {
		blocks := blocksByType[typeName]
		field := val.Type().Field((fieldIdx))

		ty := field.Type
		isSlice := false
		isPtr := false
		if ty.Kind() == reflect.Slice {
			isSlice = true
			ty = ty.Elem()
		}
		if ty.Kind() == reflect.Ptr {
			isPtr = true
			ty = ty.Elem()
		}

		if len(blocks) > 1 && !isSlice {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Duplicate %s block", typeName),
				Detail: fmt.Sprintf(
					"Only one %s block is allowed. Another was defined at %s.",
					typeName, blocks[0].DefRange.String(),
				),
				Subject: &blocks[1].DefRange,
			})
			continue
		}

		if len(blocks) == 0 {
			if isSlice || isPtr {
				if val.Field(fieldIdx).IsNil() {
					val.Field(fieldIdx).Set(reflect.Zero(field.Type))
				}
			} else {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Missing %s block", typeName),
					Detail:   fmt.Sprintf("A %s block is required.", typeName),
				})
			}
			continue
		}

		switch {

		case isSlice:
			elemType := ty
			if isPtr {
				elemType = reflect.PtrTo(ty)
			}
			sli := val.Field(fieldIdx)
			if sli.IsNil() {
				sli = reflect.MakeSlice(reflect.SliceOf(elemType), len(blocks), len(blocks))
			}

			for i, block := range blocks {
				if isPtr {
					if i >= sli.Len() {
						sli = reflect.Append(sli, reflect.New(ty))
					}
					v := sli.Index(i)
					if v.IsNil() {
						v = reflect.New(ty)
					}
					diags = append(diags, decodeBlockToValue(block, ctx, v.Elem())...)
					sli.Index(i).Set(v)
				} else {
					if i >= sli.Len() {
						sli = reflect.Append(sli, reflect.Indirect(reflect.New(ty)))
					}
					diags = append(diags, decodeBlockToValue(block, ctx, sli.Index(i))...)
				}
			}

			if sli.Len() > len(blocks) {
				sli.SetLen(len(blocks))
			}

			val.Field(fieldIdx).Set(sli)

		default:
			block := blocks[0]
			if isPtr {
				v := val.Field(fieldIdx)
				if v.IsNil() {
					v = reflect.New(ty)
				}
				diags = append(diags, decodeBlockToValue(block, ctx, v.Elem())...)
				val.Field(fieldIdx).Set(v)
			} else {
				diags = append(diags, decodeBlockToValue(block, ctx, val.Field(fieldIdx))...)
			}

		}
	}

	return diags
}

func decodeBlockToValue(block *Block, ctx *hcl.EvalContext, v reflect.Value) hcl.Diagnostics {
	diags := decodeBody(block.Body, ctx, v)

	if len(block.Labels) > 0 {
		blockTags := getFieldTags(v.Type())
		for li, lv := range block.Labels {
			lfieldIdx := blockTags.Labels[li].FieldIndex
			v.Field(lfieldIdx).Set(reflect.ValueOf(lv))
		}
	}

	return diags
}
