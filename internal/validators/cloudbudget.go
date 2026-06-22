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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// IntervalValidator validates the interval block of a cloud budget resource.
// It enforces that start and step use the same format (both percentages or both
// absolute amounts), that step is positive, and that a percentage step does not
// exceed 50%.
var _ validator.Object = IntervalValidator{}

type IntervalValidator struct{}

func (v IntervalValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v IntervalValidator) MarkdownDescription(_ context.Context) string {
	return "start and step must use the same format (both percentages or both absolute amounts), step must be positive, and a percentage step must not exceed 50%."
}

type intervalObject struct {
	Start types.String `tfsdk:"start"`
	Step  types.String `tfsdk:"step"`
}

func (v IntervalValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var interval intervalObject
	resp.Diagnostics.Append(req.ConfigValue.As(ctx, &interval, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	startKnown := !interval.Start.IsNull() && !interval.Start.IsUnknown()
	stepKnown := !interval.Step.IsNull() && !interval.Step.IsUnknown()

	startPercent := startKnown && strings.HasSuffix(interval.Start.ValueString(), "%")
	stepPercent := stepKnown && strings.HasSuffix(interval.Step.ValueString(), "%")

	if startKnown && stepKnown && startPercent != stepPercent {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Inconsistent Interval Format",
			"start and step must use the same format: either both percentages (e.g. \"50%\") or both absolute amounts (e.g. \"50000\").",
		)
		return
	}

	if stepKnown {
		resp.Diagnostics.Append(validateIntervalStep(interval.Step, req.Path.AtName("step"), stepPercent)...)
	}
}

func validateIntervalStep(step types.String, p path.Path, percent bool) diag.Diagnostics {
	raw := step.ValueString()

	n, err := strconv.Atoi(strings.TrimSuffix(raw, "%"))
	if err != nil {
		// Non-integer values are left for the store validator to report.
		return nil
	}

	if n <= 0 {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
			p, "Invalid Interval Step",
			"Interval step must be greater than 0.",
		)}
	}

	if percent && n > 50 {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
			p, "Invalid Interval Step",
			fmt.Sprintf("Percentage step must not exceed 50%%, got %d%%.", n),
		)}
	}

	return nil
}

// ReceiversValidator validates the receivers list of a cloud budget resource.
// It enforces that there are no duplicate entries and no empty strings.
var _ validator.List = ReceiversValidator{}

type ReceiversValidator struct{}

func (v ReceiversValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v ReceiversValidator) MarkdownDescription(_ context.Context) string {
	return "Receiver entries must be non-empty and unique."
}

func (v ReceiversValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var receivers []types.String
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &receivers, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	seen := make(map[string]int, len(receivers))
	for i, r := range receivers {
		if r.IsNull() || r.IsUnknown() {
			continue
		}

		val := r.ValueString()
		if val == "" {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtListIndex(i),
				"Empty Receiver Entry",
				"Receiver entries must not be empty.",
			)
			continue
		}

		if prev, ok := seen[val]; ok {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtListIndex(i),
				"Duplicate Receiver",
				fmt.Sprintf("Receiver %q is already listed at index %d.", val, prev),
			)
			continue
		}

		seen[val] = i
	}
}
