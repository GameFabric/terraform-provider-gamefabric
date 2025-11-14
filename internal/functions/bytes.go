package functions

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ function.Function = &BytesFunction{}

func NewBytesFunction() function.Function {
	return &BytesFunction{
		name: "bytes",
	}
}

type BytesFunction struct {
	name string
}

func (f BytesFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.name
}

func (f BytesFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Returns the number of bytes as a number.",
		Description: "Takes either a unitless number of bytes or a string with a unit suffix, and converts it to a unitless number of bytes",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:               "quantity",
				AllowNullValue:     false,
				AllowUnknownValues: false,
				Description:        "The number of bytes, either as a number or as a string with a unit suffix",
				Validators:         []function.DynamicParameterValidator{quantityTypeDynamicParameterValidator{}},
			},
		},
		Return: function.Int64Return{},
	}
}

func (f BytesFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))
	if resp.Error != nil {
		return
	}

	underlyingValue := input.UnderlyingValue()

	if numVal, ok := underlyingValue.(basetypes.NumberValue); ok {
		intVal, _ := numVal.ValueBigFloat().Int64()
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.Int64Value(intVal)))
		return
	}

	if strVal, ok := underlyingValue.(basetypes.StringValue); ok {
		str := strVal.ValueString()
		quantity, err := resource.ParseQuantity(str)
		if err != nil {
			resp.Error = function.NewFuncError(fmt.Sprintf("Unable to parse quantity: %s", err))
			return
		}

		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.Int64Value(quantity.Value())))
		return
	}

	resp.Error = function.NewFuncError("Unable to parse input as number or string")
}
