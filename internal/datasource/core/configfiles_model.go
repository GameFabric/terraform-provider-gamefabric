package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type configFilesModel struct {
	Environment types.String            `tfsdk:"environment"`
	LabelFilter map[string]types.String `tfsdk:"label_filter"`
	ConfigFiles []configFileModel       `tfsdk:"config_files"`
}

func newConfigFilesModel(cfgFiles []corev1.ConfigFile) configFilesModel {
	return configFilesModel{
		ConfigFiles: conv.ForEachSliceItem(cfgFiles, func(item corev1.ConfigFile) configFileModel {
			return newConfigModel(&item)
		}),
	}
}
