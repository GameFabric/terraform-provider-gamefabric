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
			return newSecretModel(&item)
		}),
	}
}
