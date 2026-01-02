package rbac

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type groupModel struct {
	Name types.String `tfsdk:"name"`
}
