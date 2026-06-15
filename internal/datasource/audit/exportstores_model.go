package audit

import (
	auditv1alpha1 "github.com/gamefabric/gf-core/pkg/api/audit/v1alpha1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
)

type exportStoresModel struct {
	ExportStores []exportStoreModel `tfsdk:"exportstores"`
}

func newExportStoresModel(objs []auditv1alpha1.ExportStore) exportStoresModel {
	return exportStoresModel{
		ExportStores: conv.ForEachSliceItem(objs, func(item auditv1alpha1.ExportStore) exportStoreModel {
			return newExportStoreModel(&item)
		}),
	}
}
