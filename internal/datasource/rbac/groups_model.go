package rbac

import (
	rbacv1 "github.com/gamefabric/gf-core/pkg/api/rbac/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type groupsModel struct {
	LabelFilter map[string]types.String `tfsdk:"label_filter"`
	Names       []types.String          `tfsdk:"names"`
}

func newGroupsModel(objs []rbacv1.Group) groupsModel {
	return groupsModel{
		Names: conv.ForEachSliceItem(objs, func(item rbacv1.Group) types.String {
			return types.StringValue(item.Name)
		}),
	}
}
