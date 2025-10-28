package authentication

import (
	authv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceAccountsModel struct {
	LabelFilters map[string]types.String `tfsdk:"label_filters"`
	Items        []serviceAccountModel   `tfsdk:"items"`
}

func newServiceAccountsModel(list *authv1.ServiceAccountList, labelFilters map[string]string) serviceAccountsModel {
	var result []serviceAccountModel
	for _, obj := range list.Items {
		if matchesLabels(obj.Labels, labelFilters) {
			result = append(result, newServiceAccountModel(&obj))
		}
	}
	return serviceAccountsModel{
		Items: result,
	}
}

func matchesLabels(labels, filters map[string]string) bool {
	for k, v := range filters {
		if labels[k] != v {
			return false
		}
	}
	return true
}
