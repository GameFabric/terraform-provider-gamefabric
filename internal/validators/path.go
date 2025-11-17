package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type foundInValidator struct {
	Message         string
	PathExpressions []path.Expression
}

// FoundIn creates a new validator that checks if a string value exists
// in any of the specified path expressions.
//
// Parameters:
//   - msg: The error message to display if the value is not found
//   - paths: One or more path expressions where the value should be found
func FoundIn(msg string, paths ...path.Expression) validator.String {
	return foundInValidator{
		Message:         msg,
		PathExpressions: paths,
	}
}

func (v foundInValidator) Description(context.Context) string {
	return "Validates that the value is found in the specified path expressions."
}

func (v foundInValidator) MarkdownDescription(context.Context) string {
	return "Validates that the value is found in the specified path expressions."
}

func (v foundInValidator) ValidateString(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
	if !conv.IsKnown(req.ConfigValue) {
		return
	}

	if len(v.PathExpressions) == 0 {
		return
	}

	summary := make([]string, 0, len(v.PathExpressions))
	for _, pathExpr := range v.PathExpressions {
		summary = append(summary, pathExpr.String())

		matchedPaths, diags := req.Config.PathMatches(ctx, pathExpr)
		res.Diagnostics.Append(diags...)

		if diags.HasError() {
			return
		}

		for _, matchedPath := range matchedPaths {
			var val attr.Value
			diags = req.Config.GetAttribute(ctx, matchedPath, &val)
			res.Diagnostics.Append(diags...)

			if diags.HasError() {
				return
			}

			if val.IsUnknown() {
				return
			}

			if val.IsNull() {
				continue
			}

			str, ok := val.(types.String)
			if !ok {
				res.Diagnostics.AddError(
					"Invalid Type",
					fmt.Sprintf("Expected value at path %q to be of type String, got %T", matchedPath.String(), val),
				)
				return
			}

			if req.ConfigValue.ValueString() == str.ValueString() {
				return
			}
		}
	}

	res.Diagnostics.AddError(
		v.Message,
		fmt.Sprintf("%q was not found in any of the provided paths: %s", req.ConfigValue.ValueString(), strings.Join(summary, ", ")),
	)
}
