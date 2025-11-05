package storage

import (
	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1beta1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type volumeModel struct {
	ID          types.String            `tfsdk:"id"`
	Name        types.String            `tfsdk:"name"`
	Environment types.String            `tfsdk:"environment"`
	Labels      map[string]types.String `tfsdk:"labels"`
	Annotations map[string]types.String `tfsdk:"annotations"`
	VolumeStore types.String            `tfsdk:"volume_store"`
	Capacity    types.String            `tfsdk:"capacity"`
}

func newVolumeModel(obj *storagev1beta1.Volume) volumeModel {
	return volumeModel{
		ID:          types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:        types.StringValue(obj.Name),
		Environment: types.StringValue(obj.Environment),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Annotations: conv.ForEachMapItem(obj.Annotations, func(item string) types.String { return types.StringValue(item) }),
		VolumeStore: types.StringValue(obj.Spec.VolumeStoreName),
		Capacity:    types.StringValue(obj.Spec.Capacity.String()),
	}
}

func (m volumeModel) ToObject() *storagev1beta1.Volume {
	return &storagev1beta1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(item types.String) string { return item.ValueString() }),
		},
		Spec: storagev1beta1.VolumeSpec{
			VolumeStoreName: m.VolumeStore.ValueString(),
			Capacity:        *conv.Quantity(m.Capacity),
		},
	}
}
