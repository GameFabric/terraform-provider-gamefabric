package planmodifiers

import (
	"context"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/dynamicbuffer"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	attrUtil = "max_buffer_utilization"
	attrMin  = "dynamic_min_buffer_threshold"
	attrMax  = "dynamic_max_buffer_threshold"
)

type dynamicBuffer struct {
	attr string
}

// NewDynamicMinBufferThreshold returns a plan modifier that calculates the dynamic minimum buffer threshold based on the maximum buffer utilization.
func NewDynamicMinBufferThreshold() planmodifier.Int32 {
	return &dynamicBuffer{
		attr: attrMin,
	}
}

// NewDynamicMaxBufferThreshold returns a plan modifier that calculates the dynamic maximum buffer threshold based on the maximum buffer utilization.
func NewDynamicMaxBufferThreshold() planmodifier.Int32 {
	return &dynamicBuffer{
		attr: attrMax,
	}
}

// PlanModifyInt32 implements the plan modification logic for the dynamic buffer thresholds.
func (d *dynamicBuffer) PlanModifyInt32(ctx context.Context, req planmodifier.Int32Request, resp *planmodifier.Int32Response) {
	if conv.IsKnown(req.ConfigValue) {
		return
	}

	// Retrieve maximum buffer utilization.
	siblingPath := req.Path.ParentPath().AtName(attrUtil)

	var maxBufferUtilization types.Int32
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, siblingPath, &maxBufferUtilization)...)
	if resp.Diagnostics.HasError() || maxBufferUtilization.IsNull() || maxBufferUtilization.IsUnknown() {
		return
	}

	// Set plan value.
	switch d.attr {
	case attrMin:
		resp.PlanValue = types.Int32Value(dynamicbuffer.DynamicMinBufferThreshold(maxBufferUtilization.ValueInt32()))

	case attrMax:
		resp.PlanValue = types.Int32Value(dynamicbuffer.DynamicMaxBufferThreshold(maxBufferUtilization.ValueInt32()))
	}
}

// Description returns a human-readable description of the plan modifier.
func (d *dynamicBuffer) Description(context.Context) string {
	return "Dynamic buffer plan modifier"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (d *dynamicBuffer) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}
