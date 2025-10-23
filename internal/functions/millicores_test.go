package functions

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMilliCoresFunctionRun(t *testing.T) {
	testCases := map[string]struct {
		request  function.RunRequest
		expected function.RunResponse
	}{
		// we've forbidden null/unknown inputs in validation, so no need to test those here
		"number-integer": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.NumberValue(big.NewFloat(2)))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("2000m")),
			},
		},
		"number-integer-large": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.NumberValue(big.NewFloat(8942)))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("8942000m")),
			},
		},
		"number-fractional": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.NumberValue(big.NewFloat(75.25)))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("75250m")),
			},
		},
		"number-fractional-large": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.NumberValue(big.NewFloat(535.5)))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("535500m")),
			},
		},
		"string-millicores": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.StringValue("500m"))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("500m")),
			},
		},
		"string-cores": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.StringValue("50"))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("50000m")),
			},
		},
		// Note: we could also test for quantities like "100M" here, which parses correctly using resource.ParseQuantity,
		// but our validation forbids them before entering the function run.
		"string-invalid": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.StringValue("abc"))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringUnknown()),
				// error message after the colon comes from https://github.com/kubernetes/apimachinery/blob/e79daceaa31bba8cd82e1b5b1619494bcaccf0e5/pkg/api/resource/quantity.go#L155
				Error: function.NewFuncError("Unable to parse quantity: quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'"),
			},
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase // a shame that we still have to do this...

		t.Run(name, func(t *testing.T) {
			got := function.RunResponse{
				Result: function.NewResultData(types.StringUnknown()),
			}

			MilliCoresFunction{}.Run(context.Background(), testCase.request, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
