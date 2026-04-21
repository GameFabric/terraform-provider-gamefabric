package validators

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Int32 = greaterOrEqual{}

type greaterOrEqual struct {
	exprs path.Expressions
}

// Description describes the validation in plain text formatting.
func (av greaterOrEqual) Description(_ context.Context) string {
	paths := make([]string, 0, len(av.exprs))
	for _, p := range av.exprs {
		paths = append(paths, p.String())
	}

	return "value must be greater than or equal to " + strings.Join(paths, " + ")
}

// MarkdownDescription describes the validation in Markdown formatting.
func (av greaterOrEqual) MarkdownDescription(ctx context.Context) string {
	return av.Description(ctx)
}

// ValidateInt32 performs the validation.
func (av greaterOrEqual) ValidateInt32(ctx context.Context, request validator.Int32Request, response *validator.Int32Response) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// Ensure input path expressions resolution against the current attribute
	expressions := request.PathExpression.MergeExpressions(av.exprs...)

	// Sum the value of all the attributes involved, but only if they are all known.
	var sumOfAttribs int32
	for _, expression := range expressions {
		matchedPaths, diags := request.Config.PathMatches(ctx, expression)
		response.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			// If the user specifies the same attribute this validator is applied to,
			// also as part of the input, skip it
			if mp.Equal(request.Path) {
				continue
			}

			// Get the value
			var matchedValue attr.Value
			diags := request.Config.GetAttribute(ctx, mp, &matchedValue)
			response.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			if matchedValue.IsUnknown() {
				return
			}

			if matchedValue.IsNull() {
				continue
			}

			// We know there is a value, convert it to the expected type
			var attribToSum types.Int32
			diags = tfsdk.ValueAs(ctx, matchedValue, &attribToSum)
			response.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			sumOfAttribs += attribToSum.ValueInt32()
		}
	}

	if request.ConfigValue.ValueInt32() < sumOfAttribs {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			av.Description(ctx),
			strconv.Itoa(int(request.ConfigValue.ValueInt32())),
		))
	}
}

func GreaterOrEqualTo(path ...path.Expression) validator.Int32 {
	return greaterOrEqual{path}
}

type lessOrEqual struct {
	exprs path.Expressions
}

// Description describes the validation in plain text formatting.
func (av lessOrEqual) Description(_ context.Context) string {
	paths := make([]string, 0, len(av.exprs))
	for _, p := range av.exprs {
		paths = append(paths, p.String())
	}

	return "value must be less than or equal to " + strings.Join(paths, " + ")
}

// MarkdownDescription describes the validation in Markdown formatting.
func (av lessOrEqual) MarkdownDescription(ctx context.Context) string {
	return av.Description(ctx)
}

// ValidateInt32 performs the validation.
func (av lessOrEqual) ValidateInt32(ctx context.Context, request validator.Int32Request, response *validator.Int32Response) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// Ensure input path expressions resolution against the current attribute
	expressions := request.PathExpression.MergeExpressions(av.exprs...)

	// Sum the value of all the attributes involved, but only if they are all known.
	var sumOfAttribs int32
	for _, expression := range expressions {
		matchedPaths, diags := request.Config.PathMatches(ctx, expression)
		response.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			// If the user specifies the same attribute this validator is applied to,
			// also as part of the input, skip it
			if mp.Equal(request.Path) {
				continue
			}

			// Get the value
			var matchedValue attr.Value
			diags := request.Config.GetAttribute(ctx, mp, &matchedValue)
			response.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			if matchedValue.IsUnknown() {
				return
			}

			if matchedValue.IsNull() {
				continue
			}

			// We know there is a value, convert it to the expected type
			var attribToSum types.Int32
			diags = tfsdk.ValueAs(ctx, matchedValue, &attribToSum)
			response.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			sumOfAttribs += attribToSum.ValueInt32()
		}
	}

	if request.ConfigValue.ValueInt32() > sumOfAttribs {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			av.Description(ctx),
			strconv.Itoa(int(request.ConfigValue.ValueInt32())),
		))
	}
}

func LessOrEqualTo(path ...path.Expression) validator.Int32 {
	return lessOrEqual{path}
}
