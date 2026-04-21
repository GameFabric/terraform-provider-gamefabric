package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type derivedScaleDownDefault struct {
	Add int32
}

func (d derivedScaleDownDefault) Description(_ context.Context) string {
	return "Defaults to a value relative to the scale up utilization."
}

func (d derivedScaleDownDefault) MarkdownDescription(_ context.Context) string {
	return "Defaults to a value relative to the scale up utilization."
}

func (d derivedScaleDownDefault) PlanModifyInt32(ctx context.Context, req planmodifier.Int32Request, resp *planmodifier.Int32Response) {
	if !req.ConfigValue.IsNull() {
		return
	}

	var scaleUpAttr types.Int32
	diags := req.Plan.GetAttribute(ctx, req.Path.ParentPath().AtName("scale_up_utilization"), &scaleUpAttr)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Can't compute if the source is unknown (not yet known at plan time).
	if scaleUpAttr.IsUnknown() || scaleUpAttr.IsNull() {
		return
	}

	val := scaleUpAttr.ValueInt32() + d.Add
	switch {
	case val < 1:
		val = 1
	case val > 99:
		val = 99
	}

	resp.PlanValue = types.Int32Value(val)
}

func DerivedScaleDownDefault(add int32) planmodifier.Int32 {
	return derivedScaleDownDefault{
		Add: add,
	}
}
