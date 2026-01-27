package mps

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerProvider is an interface for any model that has containers with resources.
type ContainerProvider interface {
	GetContainers() []ContainerModel
}

// PreserveContainerQuantities is a generic helper that preserves quantity values
// from plan containers to state containers when they are semantically equal.
// This works for any resource that implements the ContainerProvider interface.
func PreserveContainerQuantities(ctx context.Context, stateModel, planModel ContainerProvider) diag.Diagnostics {
	if stateModel == nil || planModel == nil {
		return diag.Diagnostics{}
	}

	stateContainers := stateModel.GetContainers()
	planContainers := planModel.GetContainers()

	return PreserveContainerQuantityValues(ctx, stateContainers, planContainers)
}

// PreserveContainerQuantityValues preserves quantity values from plan containers to state containers
// when they are semantically equal. This preserves the user's preferred format (e.g., "1000m" instead of "1").
func PreserveContainerQuantityValues(ctx context.Context, stateContainers, planContainers []ContainerModel) diag.Diagnostics {
	var diags diag.Diagnostics

	for i := range stateContainers {
		if i >= len(planContainers) {
			break
		}

		stateContainer := &stateContainers[i]
		planContainer := &planContainers[i]

		if stateContainer.Resources != nil && planContainer.Resources != nil {
			// Preserve limits
			if stateContainer.Resources.Limits != nil && planContainer.Resources.Limits != nil {
				preserveQuantityString(&stateContainer.Resources.Limits.CPU, &planContainer.Resources.Limits.CPU)
			}

			// Preserve requests
			if stateContainer.Resources.Requests != nil && planContainer.Resources.Requests != nil {
				preserveQuantityString(&stateContainer.Resources.Requests.CPU, &planContainer.Resources.Requests.CPU)
			}
		}
	}

	return diags
}

// preserveQuantityString replaces stateVal with planVal if they are semantically equal quantities.
func preserveQuantityString(stateVal, planVal *types.String) {
	if stateVal == nil || planVal == nil {
		return
	}

	if stateVal.IsNull() || stateVal.IsUnknown() || planVal.IsNull() || planVal.IsUnknown() {
		return
	}

	stateStr := stateVal.ValueString()
	planStr := planVal.ValueString()

	// If exactly equal, no need to check
	if stateStr == planStr {
		return
	}

	// Try to parse as quantities
	stateQty, stateErr := resource.ParseQuantity(stateStr)
	planQty, planErr := resource.ParseQuantity(planStr)

	// If both are valid quantities and semantically equal, use plan value
	if stateErr == nil && planErr == nil && stateQty.Cmp(planQty) == 0 {
		*stateVal = *planVal
	}
}
