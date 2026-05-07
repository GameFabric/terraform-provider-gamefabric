package planmodifiers

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DerivedScaleDownDefault defaults the scale down value relative to the scale up value.
func DerivedScaleDownDefault(add int32) planmodifier.Int32 {
	return derivedScaleDownDefault{
		add: add,
	}
}

type derivedScaleDownDefault struct {
	add int32
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

	if scaleUpAttr.IsUnknown() || scaleUpAttr.IsNull() {
		return
	}

	val := scaleUpAttr.ValueInt32() + d.add
	switch {
	case val < 1:
		val = 1
	case val > 99:
		val = 99
	}

	resp.PlanValue = types.Int32Value(val)
}

// SharedScaleToZeroDefault defaults region-specific scale to zero to the shared scale to zero value.
//
// The parent argument distinguishes between defaulting autoscaling and autoscaling.scale_to_zero.
// The add argument defaults the scale down value relative to the scale up value when neither the region-specific nor the shared value is set.
func SharedScaleToZeroDefault(parent bool, add int32) planmodifier.Object {
	return sharedScaleToZeroDefault{
		parent: parent,
		add:    add,
	}
}

type sharedScaleToZeroDefault struct {
	parent bool
	add    int32
}

func (d sharedScaleToZeroDefault) Description(_ context.Context) string {
	return "Defaults to the values configured in the shared autoscaling.scale_to_zero block."
}

func (d sharedScaleToZeroDefault) MarkdownDescription(_ context.Context) string {
	return "Defaults to the values configured in the shared `autoscaling.scale_to_zero` block."
}

func (d sharedScaleToZeroDefault) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}

	if !req.ConfigValue.IsNull() {
		return
	}

	var sharedSTZ types.Object
	diags := req.Plan.GetAttribute(ctx, path.Root("autoscaling").AtName("scale_to_zero"), &sharedSTZ)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if sharedSTZ.IsNull() || sharedSTZ.IsUnknown() {
		resp.PlanValue = types.ObjectNull(req.PlanValue.AttributeTypes(ctx))
		return
	}

	sharedSTZ = fillDerivedScaleDown(ctx, sharedSTZ, d.add, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !d.parent {
		// When attached directly to a scale_to_zero attribute, the inner object is already the expected shape.
		resp.PlanValue = sharedSTZ
		return
	}

	// When attached to the region's "autoscaling" attribute, wrap the shared
	// scale_to_zero object so the resulting type matches the region's
	// autoscaling schema: { scale_to_zero = { ... } }.
	wrapped, wrapDiags := types.ObjectValue(
		map[string]attr.Type{
			"scale_to_zero": sharedSTZ.Type(ctx),
		},
		map[string]attr.Value{
			"scale_to_zero": sharedSTZ,
		},
	)
	resp.Diagnostics.Append(wrapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = wrapped
}

func fillDerivedScaleDown(ctx context.Context, obj types.Object, add int32, diags *diag.Diagnostics) types.Object {
	if obj.IsNull() || obj.IsUnknown() {
		return obj
	}

	attrs := obj.Attributes()
	scaleDown, ok1 := attrs["scale_down_utilization"].(types.Int32)
	scaleUp, ok2 := attrs["scale_up_utilization"].(types.Int32)
	if !ok1 || !ok2 {
		return obj
	}

	if !scaleDown.IsNull() && !scaleDown.IsUnknown() {
		return obj
	}
	if scaleUp.IsNull() || scaleUp.IsUnknown() {
		return obj
	}

	val := scaleUp.ValueInt32() + add
	switch {
	case val < 1:
		val = 1
	case val > 99:
		val = 99
	}

	newAttrs := make(map[string]attr.Value, len(attrs)+1)
	maps.Copy(newAttrs, attrs)
	newAttrs["scale_down_utilization"] = types.Int32Value(val)

	updated, d := types.ObjectValue(obj.AttributeTypes(ctx), newAttrs)
	diags.Append(d...)
	if d.HasError() {
		return obj
	}
	return updated
}
