package core

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type locationModel struct {
	Name types.String `tfsdk:"name"`
}
