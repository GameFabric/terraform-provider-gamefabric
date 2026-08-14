package provisioning

import (
	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tokenServicesModel struct {
	LabelFilter   map[string]types.String `tfsdk:"label_filter"`
	TokenServices []tokenServiceModel     `tfsdk:"token_services"`
}

func newTokenServicesModel(items []provisioningv1beta1.TokenService) tokenServicesModel {
	return tokenServicesModel{
		TokenServices: conv.EmptyIfNil(conv.ForEachSliceItem(items, func(item provisioningv1beta1.TokenService) tokenServiceModel {
			return newTokenServiceModel(&item)
		})),
	}
}
