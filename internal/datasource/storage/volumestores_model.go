package storage

import (
	storagev1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
)

type volumeStoresModel struct {
	VolumeStores []volumeStoreModel `tfsdk:"volumestores"`
}

func newVolumeStoresModel(objs []storagev1.VolumeStore) volumeStoresModel {
	model := volumeStoresModel{
		VolumeStores: make([]volumeStoreModel, 0, len(objs)),
	}
	for _, item := range objs {
		m := newVolumeStoreModel(&item)
		model.VolumeStores = append(model.VolumeStores, m)
	}
	return model
}
