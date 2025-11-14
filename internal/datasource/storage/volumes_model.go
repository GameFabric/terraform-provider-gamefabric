package storage

import (
	storagev1beta1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type volumesModel struct {
	Environment types.String            `tfsdk:"environment"`
	LabelFilter map[string]types.String `tfsdk:"label_filter"`
	Volumes     []volumeModel           `tfsdk:"volumes"`
}

func newVolumesModel(vols []storagev1beta1.Volume) volumesModel {
	return volumesModel{
		Volumes: conv.ForEachSliceItem(vols, func(item storagev1beta1.Volume) volumeModel {
			return newVolumeModel(&item)
		}),
	}
}
