package functions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ function.Function = &GibiBytesFunction{}

func NewGibiBytesFunction() function.Function {
	return &GibiBytesFunction{
		name: "gibibytes",
	}
}

type GibiBytesFunction struct {
	name string
}

func (f GibiBytesFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.name
}

func (f GibiBytesFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Returns the number of bytes in GiB.",
		Description: "Takes either a unitless number of bytes or a string with a unit suffix, and converts it to a quantity in GiB",
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

func (f GibiBytesFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	res, err := convertToMemoryQuantity(ctx, req, resource.BinarySI, resource.Giga)

	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, err)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.StringValue(res)))
}
