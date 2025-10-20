package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type configFileModel struct {
	Name        types.String `tfsdk:"name"`
	Environment types.String `tfsdk:"environment"`
	Data        types.String `tfsdk:"data"`
}

func newConfigModel(obj *corev1.ConfigFile) configFileModel {
	return configFileModel{
		Name:        conv.OptionalFunc(obj.Name, types.StringValue, types.StringNull),
		Environment: types.StringValue(obj.Environment),
		Data:        types.StringValue(obj.Data),
	}
}
