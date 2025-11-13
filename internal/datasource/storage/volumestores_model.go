package storage

import (
	storagev1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
)

type volumeStoresModel struct {
	VolumeStores []volumeStoreModel `tfsdk:"volumestores"`
}

func newVolumeStoresModel(objs []storagev1.VolumeStore) volumeStoresModel {
	return volumeStoresModel{
		VolumeStores: conv.ForEachSliceItem(objs, func(item storagev1.VolumeStore) volumeStoreModel {
			return newVolumeStoreModel(&item)
		}),
	}
}
