package authentication

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceAccountPasswordResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ServiceAccount types.String `tfsdk:"service_account"`
	Password       types.String `tfsdk:"password"`
}
