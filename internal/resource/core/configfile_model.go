package core

import (
	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type configFileModel struct {
	ID          types.String            `tfsdk:"id"`
	Name        types.String            `tfsdk:"name"`
	Environment types.String            `tfsdk:"environment"`
	Labels      map[string]types.String `tfsdk:"labels"`
	Annotations map[string]types.String `tfsdk:"annotations"`
	Description types.String            `tfsdk:"description"`
	Data        types.String            `tfsdk:"data"`
}

func newConfigModel(obj *corev1.ConfigFile) configFileModel {
	return configFileModel{
		ID:          types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:        types.StringValue(obj.Name),
		Environment: types.StringValue(obj.Environment),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Annotations: conv.ForEachMapItem(obj.Annotations, func(item string) types.String { return types.StringValue(item) }),
		Description: conv.OptionalFunc(obj.Description, types.StringValue, types.StringNull),
		Data:        types.StringValue(obj.Data),
	}
}

func (m configFileModel) ToObject() *corev1.ConfigFile {
	return &corev1.ConfigFile{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(item types.String) string { return item.ValueString() }),
		},
		Description: m.Description.ValueString(),
		Data:        m.Data.ValueString(),
	}
}
