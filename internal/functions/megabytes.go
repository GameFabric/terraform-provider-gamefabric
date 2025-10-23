package functions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ function.Function = &MegaBytesFunction{}

func NewMegaBytesFunction() function.Function {
	return &MegaBytesFunction{
		name: "megabytes",
	}
}

type MegaBytesFunction struct {
	name string
}

func (f MegaBytesFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.name
}

func (f MegaBytesFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Returns the number of bytes in MiB.",
		Description: "Takes either a unitless number of bytes or a string with a unit suffix, and converts it to a quantity in MiB",
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

func (f MegaBytesFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	res, err := convertToMemoryQuantity(ctx, req, resource.DecimalSI, resource.Mega)

	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, err)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.StringValue(res)))
}
