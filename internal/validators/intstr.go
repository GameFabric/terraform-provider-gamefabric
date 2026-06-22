package validators

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.List = ThresholdsFormatValidator{}

// ThresholdsFormatValidator validates that all entries in a threshold list use
// the same format (all percentages or all absolute amounts), are individually
// valid, and are sorted in strictly ascending order.
type ThresholdsFormatValidator struct{}

func (v ThresholdsFormatValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v ThresholdsFormatValidator) MarkdownDescription(_ context.Context) string {
	return "All threshold entries must use the same format (all percentages or all absolute amounts), be valid, and be in strictly ascending order."
}

func (v ThresholdsFormatValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var thresholds []types.String
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &thresholds, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var globalPercent *bool
	prevVal := -1
	prevIdx := -1
	for i, t := range thresholds {
		if t.IsNull() || t.IsUnknown() {
			continue
		}

		percent := strings.HasSuffix(t.ValueString(), "%")
		if globalPercent == nil {
			globalPercent = &percent
		}

		if percent != *globalPercent {
			want := "absolute amount (e.g. \"50000\")"
			if *globalPercent {
				want = "percentage (e.g. \"50%\")"
			}
			resp.Diagnostics.AddAttributeError(
				req.Path.AtListIndex(i),
				"Inconsistent Threshold Format",
				"All thresholds must use the same format: either all percentages or all absolute amounts. "+
					"Expected a "+want+" to match the first threshold.",
			)
			continue
		}

		raw := t.ValueString()
		switch {
		case percent:
			resp.Diagnostics.Append(validatePercent(req.Path.AtListIndex(i), raw)...)
		default:
			resp.Diagnostics.Append(validateAmount(req.Path.AtListIndex(i), raw)...)
		}

		// Parse the numeric value for ascending-order validation.
		n, _ := strconv.Atoi(strings.TrimSuffix(raw, "%"))
		if prevIdx != -1 && n <= prevVal {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtListIndex(i),
				"Threshold Not in Ascending Order",
				fmt.Sprintf(
					"Thresholds must be in strictly ascending order; %s is not greater than %s.",
					raw, thresholds[prevIdx].ValueString(),
				),
			)
		}
		prevVal = n
		prevIdx = i
	}
}

func validatePercent(p path.Path, raw string) diag.Diagnostics {
	n, err := strconv.Atoi(strings.TrimSuffix(raw, "%"))
	if err != nil {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
			p, "Invalid Threshold Value",
			fmt.Sprintf("Percentage threshold must be a valid integer, got %q.", raw),
		)}
	}
	if n <= 0 {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
			p, "Invalid Threshold Value",
			"Percentage threshold must be greater than 0%.",
		)}
	}
	if n >= 100 {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
			p, "Invalid Threshold Value",
			"Percentage threshold must be less than 100%.",
		)}
	}
	return nil
}

func validateAmount(p path.Path, raw string) diag.Diagnostics {
	n, err := strconv.Atoi(raw)
	if err != nil {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
			p, "Invalid Threshold Value",
			fmt.Sprintf("Absolute threshold must be a valid integer, got %q.", raw),
		)}
	}
	if n <= 0 {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
			p, "Invalid Threshold Value",
			"Absolute threshold must be greater than 0.",
		)}
	}
	return nil
}
