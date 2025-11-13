package authentication

import (
	authv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceAccountsModel struct {
	LabelFilter     map[string]types.String `tfsdk:"label_filter"`
	ServiceAccounts []serviceAccountModel   `tfsdk:"service_accounts"`
}

func newServiceAccountsModel(items []authv1.ServiceAccount) serviceAccountsModel {
	return serviceAccountsModel{
		ServiceAccounts: conv.ForEachSliceItem(items, func(item authv1.ServiceAccount) serviceAccountModel {
			return newServiceAccountModel(&item)
		}),
	}
}
