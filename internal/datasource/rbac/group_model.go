package rbac

import (
	rbacv1 "github.com/gamefabric/gf-core/pkg/api/rbac/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type groupModel struct {
	Name types.String `tfsdk:"name"`
}

func newGroupModel(obj *rbacv1.Group) groupModel {
	return groupModel{
		Name: types.StringValue(obj.Name),
	}
}
