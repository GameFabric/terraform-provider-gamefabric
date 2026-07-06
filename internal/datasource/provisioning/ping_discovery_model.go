package provisioning

import (
	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type pingDiscoveryModel struct {
	Name        types.String   `tfsdk:"name"`
	URL         types.String   `tfsdk:"url"`
	ActiveToken types.String   `tfsdk:"active_token"`
	Tokens      []types.String `tfsdk:"tokens"`
}

func newPingDiscoveryModel(obj *provisioningv1beta1.PingDiscovery) pingDiscoveryModel {
	return pingDiscoveryModel{
		Name:        types.StringValue(obj.Name),
		URL:         conv.OptionalFunc(obj.Status.URL, types.StringValue, types.StringNull),
		ActiveToken: activeToken(obj.Status.Tokens),
		Tokens:      conv.EmptyIfNil(conv.ForEachSliceItem(obj.Status.Tokens, types.StringValue)),
	}
}
