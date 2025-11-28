package core

import (
	"time"

	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type secretModel struct {
	Name           types.String `tfsdk:"name"`
	Environment    types.String `tfsdk:"environment"`
	Description    types.String `tfsdk:"description"`
	Status         types.String `tfsdk:"status"`
	LastDataChange types.String `tfsdk:"last_data_change"`
}

func newSecretModel(obj *corev1.Secret) secretModel {
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
