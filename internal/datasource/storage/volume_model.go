package storage

import (
	storagev1beta1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type volumeModel struct {
	Name        types.String `tfsdk:"name"`
	Environment types.String `tfsdk:"environment"`
}

func newVolumeModel(vol *storagev1beta1.Volume) volumeModel {
	return volumeModel{
		Name:        types.StringValue(vol.Name),
		Environment: types.StringValue(vol.Environment),
	}
}
