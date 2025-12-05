package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type secretsModel struct {
	Environment types.String            `tfsdk:"environment"`
	LabelFilter map[string]types.String `tfsdk:"label_filter"`
	Secrets     []secretModel           `tfsdk:"secrets"`
}

func newSecretsModel(items []corev1.Secret) secretsModel {
	return secretsModel{
		Secrets: conv.ForEachSliceItem(items, func(item corev1.Secret) secretModel {
			return newSecretItemModel(&item)
		}),
	}
}

func newSecretItemModel(obj *corev1.Secret) secretModel {
	description := types.StringNull()
	if obj.Description != "" {
		description = types.StringValue(obj.Description)
	}

	return secretModel{
		Name:        types.StringValue(obj.Name),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Environment: types.StringValue(obj.Environment),
		Description: description,
		Data:        dataKeys(obj),
	}
}
