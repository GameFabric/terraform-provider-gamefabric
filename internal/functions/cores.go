package functions

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ function.Function = &CoresFunction{}

func NewCoresFunction() function.Function {
	return &CoresFunction{
		name: "cores",
	}
}

type CoresFunction struct {
	name string
}

func (f CoresFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.name
}

func (f CoresFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Returns the number of cores as a number.",
		Description: "Takes either a (fractional) unitless number of cores or millicores (<number>m), and converts it to a unitless number of cores",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:               "cores",
				AllowNullValue:     false,
				AllowUnknownValues: false,
				Description:        "The number of cores, either as a number or as a string with 'm' suffix for millicores",
				Validators:         []function.DynamicParameterValidator{quantityTypeDynamicParameterValidator{}},
			},
		},
		Return: function.NumberReturn{},
	}
}

func (f CoresFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))
	if resp.Error != nil {
		return
	}

	underlyingValue := input.UnderlyingValue()

	if numVal, ok := underlyingValue.(basetypes.NumberValue); ok {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, numVal))
		return
	}

	if strVal, ok := underlyingValue.(basetypes.StringValue); ok {
		str := strVal.ValueString()
		quantity, err := resource.ParseQuantity(str)
		if err != nil {
			resp.Error = function.NewFuncError(fmt.Sprintf("Unable to parse quantity: %s", err))
			return
		}

		// don't really care for speed here right now, if this raises concerns, use AsApproximateFloat64()
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.NumberValue(big.NewFloat(quantity.AsFloat64Slow()))))
		return
	}

	resp.Error = function.NewFuncError("Unable to parse input as number or string")
}
