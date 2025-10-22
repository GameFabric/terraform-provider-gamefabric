package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type configFileModel struct {
	Name        types.String `tfsdk:"name"`
	Environment types.String `tfsdk:"environment"`
	Data        types.String `tfsdk:"data"`
}

func newConfigModel(obj *corev1.ConfigFile) configFileModel {
	return configFileModel{
		Name:        types.StringValue(obj.Name),
		Environment: types.StringValue(obj.Environment),
		Data:        types.StringValue(obj.Data),
	}
}
