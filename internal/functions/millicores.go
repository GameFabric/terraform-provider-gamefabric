package functions

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ function.Function = &MilliCoresFunction{}

func NewMilliCoresFunction() function.Function {
	return &CoresFunction{
		name: "millicores",
	}
}

type MilliCoresFunction struct {
	name string
}

func (f MilliCoresFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.name
}

func (f MilliCoresFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Returns the number of cores as millicores.",
		Description: "Takes either a (fractional) unitless number of cores or cores as millicores (<number>m), and converts it to millicores.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:               "cores",
				AllowNullValue:     false,
				AllowUnknownValues: false,
				Description:        "The number of cores, either as a number or as a string with 'm' suffix for millicores",
				Validators:         []function.DynamicParameterValidator{millicoresTypeDynamicParameterValidator{}},
			},
		},
		Return: function.StringReturn{},
	}
}

func (f MilliCoresFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))
	if resp.Error != nil {
		return
	}

	underlyingValue := input.UnderlyingValue()

	var quantity resource.Quantity
	switch underlyingValue := underlyingValue.(type) {
	case basetypes.NumberValue:
		// Note: this might overflow, as Number types are up to 512 bit floats
		val, _ := underlyingValue.ValueBigFloat().Float64()
		quantity = *resource.NewMilliQuantity(int64(val*1000), resource.DecimalSI)
	case basetypes.StringValue:
		str := underlyingValue.ValueString()
		var err error
		quantity, err = resource.ParseQuantity(str)
		if err != nil {
			resp.Error = function.NewFuncError(fmt.Sprintf("Unable to parse quantity: %s", err))
			return
		}
	default:
		resp.Error = function.NewFuncError("Unable to parse input as number or string")
		return
	}

	result := quantity.MilliValue()

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.StringValue(fmt.Sprintf("%dm", result))))
}

type millicoresTypeDynamicParameterValidator struct{}

func (v millicoresTypeDynamicParameterValidator) ValidateParameterDynamic(ctx context.Context, req function.DynamicParameterValidatorRequest, resp *function.DynamicParameterValidatorResponse) {
	if req.Value.IsNull() || req.Value.IsUnknown() {
		return
	}

	underlyingValue := req.Value.UnderlyingValue()

	if numVal, ok := underlyingValue.(basetypes.NumberValue); ok {
		if !numVal.IsNull() && !numVal.IsUnknown() {
			return // Valid number
		}
	}

	if strVal, ok := underlyingValue.(basetypes.StringValue); ok {
		if strVal.IsNull() || strVal.IsUnknown() {
			return
		}

		str := strVal.ValueString()
		if !strings.HasSuffix(str, "m") {
			resp.Error = function.NewArgumentFuncError(
				req.ArgumentPosition,
				fmt.Sprintf("Invalid string format: '%s' must end with 'm' for millicores", str),
			)
			return
		}

		numPart := strings.TrimSuffix(str, "m")
		if _, err := strconv.ParseFloat(numPart, 64); err != nil {
			resp.Error = function.NewArgumentFuncError(
				req.ArgumentPosition,
				fmt.Sprintf("Invalid millicore value: '%s' is not a valid number", numPart),
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
