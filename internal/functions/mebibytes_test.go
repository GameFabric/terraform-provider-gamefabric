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

func TestMebiBytesFunctionRun(t *testing.T) {
	testCases := map[string]struct {
		request  function.RunRequest
		expected function.RunResponse
	}{
		// we've forbidden null/unknown inputs in validation, so no need to test those here
		"32-bytes": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.NumberValue(big.NewFloat(32)))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("1Mi")), // rounded up
			},
		},
		"2-megabytes-in-bytes": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.NumberValue(big.NewFloat(2000000)))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("2Mi")), // rounded up, as 2,000,000 bytes is ~1.907 MiB
			},
		},
		"1-gibibyte-in-bytes": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.NumberValue(big.NewFloat(1073741824)))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("1024Mi")),
			},
		},
		"32-gigabytes-in-bytes": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.NumberValue(big.NewFloat(32000000000)))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("30518Mi")), // rounded up, 30517,578125
			},
		},
		"1Gi": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.StringValue("1Gi"))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("1024Mi")),
			},
		},
		"1234Mi": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.StringValue("1234Mi"))}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("1234Mi")),
			},
		},
		// Note: we could also test for quantities like "100M" here, which parses correctly using resource.ParseQuantity,
		// but our validation forbids them before entering the function run.
		// "string-invalid": {
		// 	request: function.RunRequest{
		// 		Arguments: function.NewArgumentsData([]attr.Value{types.DynamicValue(types.StringValue("abc"))}),
		// 	},
		// 	expected: function.RunResponse{
		// 		Result: function.NewResultData(types.StringUnknown()),
		// 		// error message after the colon comes from https://github.com/kubernetes/apimachinery/blob/e79daceaa31bba8cd82e1b5b1619494bcaccf0e5/pkg/api/resource/quantity.go#L155
		// 		Error: function.NewFuncError("Unable to parse quantity: quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'"),
		// 	},
		// },
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase // a shame that we still have to do this...

		t.Run(name, func(t *testing.T) {
			got := function.RunResponse{
				Result: function.NewResultData(types.StringUnknown()),
			}

			MebiBytesFunction{}.Run(context.Background(), testCase.request, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
