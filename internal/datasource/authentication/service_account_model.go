package authentication

import (
	authv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceAccountModel struct {
	Name     types.String            `tfsdk:"name"`
	Labels   map[string]types.String `tfsdk:"labels"`
	Username types.String            `tfsdk:"username"`
	Email    types.String            `tfsdk:"email"`
	State    types.String            `tfsdk:"state"`
	Password types.String            `tfsdk:"password"`
}

func newServiceAccountModel(obj *authv1.ServiceAccount) serviceAccountModel {
	return serviceAccountModel{
		Name:     conv.OptionalFunc(obj.Name, types.StringValue, types.StringNull),
		Labels:   conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Username: conv.OptionalFunc(obj.Spec.Username, types.StringValue, types.StringNull),
		Email:    conv.OptionalFunc(obj.Spec.Email, types.StringValue, types.StringNull),
		State:    conv.OptionalFunc(string(obj.Status.State), types.StringValue, types.StringNull),
		Password: conv.OptionalFunc(obj.Status.Password, types.StringValue, types.StringNull),
	}
}
