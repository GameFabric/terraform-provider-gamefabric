package rbac

import (
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	rbacv1 "github.com/gamefabric/gf-core/pkg/api/rbac/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type groupModel struct {
	ID          types.String            `tfsdk:"id"`
	Name        types.String            `tfsdk:"name"`
	Labels      map[string]types.String `tfsdk:"labels"`
	Annotations map[string]types.String `tfsdk:"annotations"`
	Users       []types.String          `tfsdk:"users"`
}

func newGroupModel(obj *rbacv1.Group) groupModel {
	return groupModel{
		ID:          types.StringValue(obj.Name),
		Name:        types.StringValue(obj.Name),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Annotations: conv.ForEachMapItem(obj.Annotations, func(item string) types.String { return types.StringValue(item) }),
		Users:       conv.ForEachSliceItem(obj.Users, func(item string) types.String { return types.StringValue(item) }),
	}
}

func (m groupModel) ToObject() *rbacv1.Group {
	return &rbacv1.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(item types.String) string { return item.ValueString() }),
		},
		Users: conv.ForEachSliceItem(m.Users, func(item types.String) string { return item.ValueString() }),
	}
}
