package protection

import (
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type gatewayPolicyModel struct {
	ID               types.String            `tfsdk:"id"`
	Name             types.String            `tfsdk:"name"`
	Labels           map[string]types.String `tfsdk:"labels"`
	Annotations      map[string]types.String `tfsdk:"annotations"`
	DisplayName      types.String            `tfsdk:"display_name"`
	Description      types.String            `tfsdk:"description"`
	DestinationCIDRs []types.String          `tfsdk:"destination_cidrs"`
}

func newGatewayPolicyModel(obj *protectionv1.GatewayPolicy) gatewayPolicyModel {
	return gatewayPolicyModel{
		ID:               types.StringValue(obj.Name),
		Name:             types.StringValue(obj.Name),
		Labels:           conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Annotations:      conv.ForEachMapItem(obj.Annotations, func(item string) types.String { return types.StringValue(item) }),
		DisplayName:      types.StringValue(obj.Spec.DisplayName),
		Description:      types.StringValue(obj.Spec.Description),
		DestinationCIDRs: conv.ForEachSliceItem(obj.Spec.DestinationCIDRs, func(item string) types.String { return types.StringValue(item) }),
	}
}

func (m gatewayPolicyModel) ToObject() *protectionv1.GatewayPolicy {
	return &protectionv1.GatewayPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(item types.String) string { return item.ValueString() }),
		},
		Spec: protectionv1.GatewayPolicySpec{
			DisplayName:      m.DisplayName.ValueString(),
			Description:      m.Description.ValueString(),
			DestinationCIDRs: conv.ForEachSliceItem(m.DestinationCIDRs, func(v types.String) string { return v.ValueString() }),
		},
	}
}
