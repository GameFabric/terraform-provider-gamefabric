package core

import (
	"slices"
	"strings"

	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type secretModel struct {
	Name        types.String            `tfsdk:"name"`
	Labels      map[string]types.String `tfsdk:"labels"`
	Environment types.String            `tfsdk:"environment"`
	Description types.String            `tfsdk:"description"`
	Data        []types.String          `tfsdk:"data"`
}

func newSecretModel(obj *corev1.Secret) secretModel {
	return secretModel{
		Name:        types.StringValue(obj.Name),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Environment: types.StringValue(obj.Environment),
		Description: conv.OptionalFunc(obj.Description, types.StringValue, types.StringNull),
		Data:        dataKeys(obj),
	}
}

func dataKeys(obj *corev1.Secret) []types.String {
	keys := make([]types.String, 0, len(obj.Data))
	for k := range obj.Data {
		keys = append(keys, types.StringValue(k))
	}
	slices.SortFunc(keys, func(a, b types.String) int {
		return strings.Compare(a.ValueString(), b.ValueString())
	})
	return keys
}
