package protection

import (
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type gatewayPolicyModel struct {
	Name             types.String            `tfsdk:"name"`
	Labels           map[string]types.String `tfsdk:"labels"`
	DisplayName      types.String            `tfsdk:"display_name"`
	Description      types.String            `tfsdk:"description"`
	DestinationCIDRs []types.String          `tfsdk:"destination_cidrs"`
}

func newGatewayPolicyModel(obj *protectionv1.GatewayPolicy) gatewayPolicyModel {
	return gatewayPolicyModel{
		Name:             types.StringValue(obj.Name),
		Labels:           conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		DisplayName:      conv.OptionalFunc(obj.Spec.DisplayName, types.StringValue, types.StringNull),
		Description:      conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		DestinationCIDRs: conv.ForEachSliceItem(obj.Spec.DestinationCIDRs, func(item string) types.String { return types.StringValue(item) }),
	}
}
