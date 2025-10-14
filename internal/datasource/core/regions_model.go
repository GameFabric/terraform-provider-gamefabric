package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type regionsModel struct {
	Environment types.String            `tfsdk:"environment"`
	LabelFilter map[string]types.String `tfsdk:"label_filter"`
	Regions     []regionModel           `tfsdk:"regions"`
}

func newRegionsModel(regions []corev1.Region) regionsModel {
	return regionsModel{
		Regions: conv.ForEachSliceItem(regions, func(item corev1.Region) regionModel {
			return newRegionModel(&item)
		}),
	}
}
