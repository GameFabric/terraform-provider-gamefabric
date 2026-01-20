package authentication

import (
	"fmt"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	authv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const ecDomain = "ec.nitrado.systems"

type serviceAccountResourceModel struct {
	ID     types.String            `tfsdk:"id"`
	Name   types.String            `tfsdk:"name"`
	Labels map[string]types.String `tfsdk:"labels"`
	Email  types.String            `tfsdk:"email"`
}

func newServiceAccountResourceModel(obj *authv1.ServiceAccount) serviceAccountResourceModel {
	return serviceAccountResourceModel{
		ID:     types.StringValue(obj.Name),
		Name:   types.StringValue(obj.Name),
		Labels: toStringMap(obj.Labels),
		Email:  types.StringValue(obj.Spec.Email),
	}
}

func toStringMap(m map[string]string) map[string]types.String {
	res := make(map[string]types.String, len(m))
	for k, v := range m {
		res[k] = types.StringValue(v)
	}
	return res
}

func (m serviceAccountResourceModel) ToObject() *authv1.ServiceAccount {
	return &authv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:   m.Name.ValueString(),
			Labels: conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
		},
		Spec: authv1.ServiceAccountSpec{
			Username: m.Name.ValueString(),
			Email:    fmt.Sprintf("%s@%s", m.Name.ValueString(), ecDomain),
		},
	}
}
