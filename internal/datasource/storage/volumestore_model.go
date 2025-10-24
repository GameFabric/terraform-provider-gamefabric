package storage

import (
	storagev1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type volumeStoreModel struct {
	Name          types.String `tfsdk:"name"`
	Region        types.String `tfsdk:"region"`
	MaxVolumeSize types.String `tfsdk:"max_volume_size"`
}

func newVolumeStoreModel(obj *storagev1.VolumeStore) volumeStoreModel {
	return volumeStoreModel{
		Name:          conv.OptionalFunc(obj.Name, types.StringValue, types.StringNull),
		Region:        types.StringValue(obj.Spec.Region),
		MaxVolumeSize: conv.OptionalFunc(obj.Spec.MaxVolumeSize.String(), types.StringValue, types.StringNull),
	}
}
