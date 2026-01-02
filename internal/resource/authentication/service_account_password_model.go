package authentication

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceAccountPasswordResourceModel struct {
	ID             types.String            `tfsdk:"id"`
	ServiceAccount types.String            `tfsdk:"service_account"`
	Labels         map[string]types.String `tfsdk:"labels"`
	PasswordWo     types.String            `tfsdk:"password_wo"`
}
