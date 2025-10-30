package core

import (
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type environmentModel struct {
	ID          types.String            `tfsdk:"id"`
	Name        types.String            `tfsdk:"name"`
	Labels      map[string]types.String `tfsdk:"labels"`
	Annotations map[string]types.String `tfsdk:"annotations"`
	DisplayName types.String            `tfsdk:"display_name"`
	Description types.String            `tfsdk:"description"`
}

func newEnvironmentModel(obj *corev1.Environment) environmentModel {
	return environmentModel{
		ID:          types.StringValue(obj.Name),
		Name:        types.StringValue(obj.Name),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Annotations: conv.ForEachMapItem(obj.Annotations, func(item string) types.String { return types.StringValue(item) }),
		DisplayName: types.StringValue(obj.Spec.DisplayName),
		Description: types.StringValue(obj.Spec.Description),
	}
}

func (m environmentModel) ToObject() *corev1.Environment {
	return &corev1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(item types.String) string { return item.ValueString() }),
		},
		Spec: corev1.EnvironmentSpec{
			DisplayName: m.DisplayName.ValueString(),
			Description: m.Description.ValueString(),
		},
	}
}
