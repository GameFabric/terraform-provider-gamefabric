package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type environmentModel struct {
	Name        types.String            `tfsdk:"name"`
	Labels      map[string]types.String `tfsdk:"labels"`
	DisplayName types.String            `tfsdk:"display_name"`
	Description types.String            `tfsdk:"description"`
}

func newEnvironmentModel(obj *corev1.Environment) environmentModel {
	return environmentModel{
		Name:        types.StringValue(obj.Name),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		DisplayName: conv.OptionalFunc(obj.Spec.DisplayName, types.StringValue, types.StringNull),
		Description: conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
	}
}
