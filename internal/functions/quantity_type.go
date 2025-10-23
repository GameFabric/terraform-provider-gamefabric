package functions

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"k8s.io/apimachinery/pkg/api/resource"
)

type quantityTypeDynamicParameterValidator struct{}

func (v quantityTypeDynamicParameterValidator) ValidateParameterDynamic(ctx context.Context, req function.DynamicParameterValidatorRequest, resp *function.DynamicParameterValidatorResponse) {
	if req.Value.IsNull() || req.Value.IsUnknown() {
		return
	}

	underlyingValue := req.Value.UnderlyingValue()

	if numVal, ok := underlyingValue.(basetypes.NumberValue); ok {
		if !numVal.IsNull() && !numVal.IsUnknown() {
			return
		}
	}

	if strVal, ok := underlyingValue.(basetypes.StringValue); ok {
		if strVal.IsNull() || strVal.IsUnknown() {
			return
		}

		str := strVal.ValueString()

		_, err := resource.ParseQuantity(str)
		if err != nil {
			resp.Error = function.NewArgumentFuncError(
				req.ArgumentPosition,
				fmt.Sprintf("Invalid quantity format: %s", err),
			)
			return
		}
		return
	}

	resp.Error = function.NewArgumentFuncError(
		req.ArgumentPosition,
		"Invalid type: value must be either a number or a valid quantity string",
	)
}

func convertToMemoryQuantity(ctx context.Context, req function.RunRequest, format resource.Format, scale resource.Scale) (string, *function.FuncError) {
	var input types.Dynamic

	err := req.Arguments.Get(ctx, &input)
	if err != nil {
		return "", err
	}
	underlyingValue := input.UnderlyingValue()

	var quantity resource.Quantity
	switch underlyingValue := underlyingValue.(type) {
	case basetypes.NumberValue:
		// Note: this might overflow, as Number types are up to 512 bit floats
		val, _ := underlyingValue.ValueBigFloat().Int64()
		quantity.Set(val)
		quantity.Format = resource.DecimalSI
	case basetypes.StringValue:
		str := underlyingValue.ValueString()
		var err error
		quantity, err = resource.ParseQuantity(str)
		if err != nil {
			return "", function.NewFuncError(fmt.Sprintf("Unable to parse quantity: %s", err))
		}
	default:
		return "", function.NewFuncError("Unable to parse input as number or string")
	}

	suffix := suffixer(format, scale)

	bytes := quantity.Value()
	if bytes == 0 {
		return fmt.Sprintf("0%s", suffix), nil
	}

	var mul int
	switch scale {
	case resource.Kilo:
		mul = 1
	case resource.Mega:
		mul = 2
	case resource.Giga:
		mul = 3
	case resource.Tera:
		mul = 4
	}
	var divisor float64
	switch format {
	case resource.BinarySI:
		divisor = 1024.0
	case resource.DecimalSI:
		divisor = 1000.0
	}
	for i := 0; i < mul; i++ {
		bytes = int64(math.Ceil(float64(bytes) / divisor))
	}
	if bytes == 0 {
		bytes = 1 // round up to at least 1 unit
	}

	return fmt.Sprintf("%d%s", bytes, suffix), nil
}

func suffixer(format resource.Format, scale resource.Scale) string {
	switch format {
	case resource.BinarySI:
		switch scale {
		case resource.Kilo:
			return "Ki"
		case resource.Mega:
			return "Mi"
		case resource.Giga:
			return "Gi"
		case resource.Tera:
			return "Ti"
		}
	case resource.DecimalSI:
		switch scale {
		case resource.Kilo:
			return "K"
		case resource.Mega:
			return "M"
		case resource.Giga:
			return "G"
		case resource.Tera:
			return "T"
		}
	}
	return ""
}
