package core

import (
	"time"

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
	lastDataChange := types.StringNull()
	if obj.Status.LastDataChange != nil {
		lastDataChange = types.StringValue(obj.Status.LastDataChange.Format(time.RFC3339))
	}
	description := types.StringNull()
	if obj.Description != "" {
		description = types.StringValue(obj.Description)
	}

	return secretModel{
		Name:           types.StringValue(obj.Name),
		Environment:    types.StringValue(obj.Environment),
		Description:    description,
		Status:         types.StringValue(string(obj.Status.State)),
		LastDataChange: lastDataChange,
	}
}
