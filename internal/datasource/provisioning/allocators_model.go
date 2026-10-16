package provisioning

import (
	provisioningv1beta3 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta3"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type allocatorsModel struct {
	LabelFilter map[string]types.String `tfsdk:"label_filter"`
	Allocators  []allocatorModel        `tfsdk:"allocators"`
}

func newAllocatorsModel(items []provisioningv1beta3.Allocator) allocatorsModel {
	return allocatorsModel{
		Allocators: conv.EmptyIfNil(conv.ForEachSliceItem(items, func(item provisioningv1beta3.Allocator) allocatorModel {
			return newAllocatorModel(&item)
		})),
	}
}
