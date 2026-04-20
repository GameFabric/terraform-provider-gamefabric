package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type derivedScaleDownDefault struct{}

func (m derivedScaleDownDefault) Description(_ context.Context) string {
	return "Defaults to other_attr minus 5."
}

func (m derivedScaleDownDefault) MarkdownDescription(_ context.Context) string {
	return "Defaults to `other_attr - 5`."
}

func (m derivedScaleDownDefault) PlanModifyInt32(ctx context.Context, req planmodifier.Int32Request, resp *planmodifier.Int32Response) {
	if !req.ConfigValue.IsNull() {
		return
	}

	var otherAttr types.Int32
	diags := req.Plan.GetAttribute(ctx, req.Path.ParentPath().AtName("scale_up_utilization"), &otherAttr)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Can't compute if the source is unknown (not yet known at plan time).
	if otherAttr.IsUnknown() || otherAttr.IsNull() {
		return
	}

	resp.PlanValue = types.Int32Value(otherAttr.ValueInt32() - 5)
}

func DerivedScaleDownDefault() planmodifier.Int32 {
	return derivedScaleDownDefault{}
}
