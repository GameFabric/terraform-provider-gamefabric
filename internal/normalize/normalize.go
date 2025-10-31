package normalize

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// State provides access to the current state or plan attributes.
type State interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}

// Model normalizes the given model in the context of the current state.
func Model(ctx context.Context, model any, state State) diag.Diagnostics {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		// This is always a developer error.
		panic(fmt.Errorf("expected T to be a struct or pointer to a struct, got %T", model))
	}

	return structType(ctx, v, v.Type(), state, path.Empty())
}

var attrValueType = reflect.TypeOf((*attr.Value)(nil)).Elem()

func normTypes(ctx context.Context, v reflect.Value, t reflect.Type, state State, p path.Path) diag.Diagnostics {
	if t.Implements(attrValueType) {
		// This is an attr.Value, treat as primitive.
		return primitiveType(ctx, v, state, p)
	}

	switch t.Kind() {
	case reflect.Ptr:
		return ptrType(ctx, v, t, state, p)
	case reflect.Struct:
		return structType(ctx, v, t, state, p)
	case reflect.Map:
		return mapType(ctx, v, t, state, p)
	case reflect.Slice:
		return sliceType(ctx, v, t, state, p)
	default:
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unsupported Type",
				fmt.Sprintf("The type %s is not supported for default normalization.", t.String()),
			),
		}
	}
}

func ptrType(ctx context.Context, v reflect.Value, t reflect.Type, state State, p path.Path) diag.Diagnostics {
	var tfVal attr.Value
	diags := state.GetAttribute(ctx, p, &tfVal)
	if diags.HasError() {
		return diags
	}

	switch {
	case v.IsNil() && !tfVal.IsNull():
		// Set the pointer to a new value as Terraform has a non-null value.
		v.Set(reflect.New(t.Elem()))
		return normTypes(ctx, v.Elem(), t.Elem(), state, p)
	case !v.IsNil() && v.IsZero() && tfVal.IsNull():
		// Set the pointer to nil as Terraform has a null value.
		v.Set(reflect.Zero(t))
		return nil
	case v.IsNil() && tfVal.IsNull():
		// Both are nil, nothing to do.
		return nil
	}
	return normTypes(ctx, v.Elem(), t.Elem(), state, p)
}

func structType(ctx context.Context, v reflect.Value, t reflect.Type, state State, p path.Path) diag.Diagnostics {
	// A struct type cannot be normalized as there is no other terraform value to set.
	// Just check the fields.
	var diags diag.Diagnostics
	for i := range t.NumField() {
		fld := v.Field(i)
		fldType := t.Field(i)

		tagName := fldType.Tag.Get("tfsdk")
		if tagName == "" || tagName == "-" {
			continue
		}

		diags.Append(normTypes(ctx, fld, fldType.Type, state, p.AtName(tagName))...)
	}
	return diags
}

func mapType(ctx context.Context, v reflect.Value, t reflect.Type, state State, p path.Path) diag.Diagnostics {
	if v.IsNil() || v.Len() == 0 {
		var tfVal attr.Value
		diags := state.GetAttribute(ctx, p, &tfVal)
		if diags.HasError() {
			return diags
		}

		switch {
		case tfVal.IsNull() && !v.IsNil():
			v.Set(reflect.Zero(v.Type()))
		case !tfVal.IsNull() && v.IsNil():
			v.Set(reflect.MakeMap(v.Type()))
		}
		return nil
	}

	var diags diag.Diagnostics
	for _, key := range v.MapKeys() {
		diags.Append(normTypes(ctx, v.MapIndex(key), t.Elem(), state, p.AtMapKey(key.String()))...)
	}
	return diags
}

func sliceType(ctx context.Context, v reflect.Value, t reflect.Type, state State, p path.Path) diag.Diagnostics {
	if v.IsNil() || v.Len() == 0 {
		var tfVal attr.Value
		diags := state.GetAttribute(ctx, p, &tfVal)
		if diags.HasError() {
			return diags
		}

		switch {
		case tfVal.IsNull() && !v.IsNil():
			v.Set(reflect.Zero(v.Type()))
		case !tfVal.IsNull() && v.IsNil():
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		}
		return nil
	}

	var diags diag.Diagnostics
	for i := range v.Len() {
		diags.Append(normTypes(ctx, v.Index(i), t.Elem(), state, p.AtListIndex(i))...)
	}
	return diags
}

func primitiveType(ctx context.Context, v reflect.Value, state State, p path.Path) diag.Diagnostics {
	var tfVal attr.Value
	diags := state.GetAttribute(ctx, p, &tfVal)
	if diags.HasError() {
		return diags
	}

	val := v.Interface().(attr.Value)
	if (tfVal.IsNull() && !val.IsNull() && isZeroValue(ctx, val)) || (!tfVal.IsNull() && val.IsNull()) {
		v.Set(reflect.ValueOf(tfVal))
	}
	return nil
}

func isZeroValue(ctx context.Context, v attr.Value) bool {
	switch v.Type(ctx) {
	case types.StringType:
		return v.(types.String).ValueString() == ""
	case types.Int64Type:
		return v.(types.Int64).ValueInt64() == 0
	case types.Int32Type:
		return v.(types.Int32).ValueInt32() == 0
	case types.Float64Type:
		return v.(types.Float64).ValueFloat64() == 0.0
	case types.Float32Type:
		return v.(types.Float32).ValueFloat32() == 0.0
	case types.BoolType:
		return v.(types.Bool).ValueBool() == false
	default:
		return false
	}
}
