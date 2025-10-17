package protection

import (
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type gatewayPoliciesModel struct {
	LabelFilter map[string]types.String `tfsdk:"label_filter"`
	Names       []types.String          `tfsdk:"names"`
}

func newGatewayPoliciesModel(objs []protectionv1.GatewayPolicy) gatewayPoliciesModel {
	return gatewayPoliciesModel{
		Names: conv.ForEachSliceItem(objs, func(item protectionv1.GatewayPolicy) types.String {
			return types.StringValue(item.Name)
		}),
	}
}
