package functions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ function.Function = &TeraBytesFunction{}

func NewTeraBytesFunction() function.Function {
	return &TeraBytesFunction{
		name: "terabytes",
	}
}

type TeraBytesFunction struct {
	name string
}

func (f TeraBytesFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.name
}

func (f TeraBytesFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Returns the number of bytes in TB.",
		Description: "Takes either a unitless number of bytes or a string with a unit suffix, and converts it to a quantity in TB",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:               "quantity",
				AllowNullValue:     false,
				AllowUnknownValues: false,
				Description:        "The quantity, either as a number or as a string with a unit suffix",
				Validators:         []function.DynamicParameterValidator{quantityTypeDynamicParameterValidator{}},
			},
		},
		Return: function.StringReturn{},
	}
}

func (f TeraBytesFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	res, err := convertToMemoryQuantity(ctx, req, resource.DecimalSI, resource.Tera)

	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, err)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.StringValue(res)))
}
