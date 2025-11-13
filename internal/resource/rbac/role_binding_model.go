package rbac

import (
	v1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	rbacv1 "github.com/gamefabric/gf-core/pkg/api/rbac/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type roleBindingModel struct {
	ID     types.String   `tfsdk:"id"`
	Users  []types.String `tfsdk:"users"`
	Groups []types.String `tfsdk:"groups"`
	Role   types.String   `tfsdk:"role"`
}

func newRoleBindingModel(obj *rbacv1.RoleBinding) roleBindingModel {
	return roleBindingModel{
		ID:     types.StringValue(obj.Name),
		Users:  conv.ForEachSliceItem(obj.Users, types.StringValue),
		Groups: conv.ForEachSliceItem(obj.Groups, types.StringValue),
		Role:   types.StringValue(obj.Role),
	}
}

func (m roleBindingModel) ToObject() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: m.Role.ValueString(),
		},
		Users:  conv.ForEachSliceItem(m.Users, func(item types.String) string { return item.ValueString() }),
		Groups: conv.ForEachSliceItem(m.Groups, func(item types.String) string { return item.ValueString() }),
		Role:   m.Role.ValueString(),
	}
}
