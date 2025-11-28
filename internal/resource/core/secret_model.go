package core

import (
	"time"

	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type secretModel struct {
	ID             types.String            `tfsdk:"id"`
	Name           types.String            `tfsdk:"name"`
	Environment    types.String            `tfsdk:"environment"`
	Labels         map[string]types.String `tfsdk:"labels"`
	Annotations    map[string]types.String `tfsdk:"annotations"`
	Description    types.String            `tfsdk:"description"`
	Data           map[string]types.String `tfsdk:"data"`
	State          types.String            `tfsdk:"state"`
	LastDataChange types.String            `tfsdk:"last_data_change"`
}

func newSecretModel(obj *corev1.Secret) secretModel {
	dataMap := make(map[string]types.String)
	for k, v := range obj.Data {
		dataMap[k] = types.StringValue(v)
	}

	var lastDataChange types.String
	if obj.Status.LastDataChange != nil {
		lastDataChange = types.StringValue(obj.Status.LastDataChange.Format(time.RFC3339))
	} else {
		lastDataChange = types.StringNull()
	}

	return secretModel{
		ID:             types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:           types.StringValue(obj.Name),
		Environment:    types.StringValue(obj.Environment),
		Labels:         conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Annotations:    conv.ForEachMapItem(obj.Annotations, func(item string) types.String { return types.StringValue(item) }),
		Description:    conv.OptionalFunc(obj.Description, types.StringValue, types.StringNull),
		Data:           dataMap,
		State:          types.StringValue(string(obj.Status.State)),
		LastDataChange: lastDataChange,
	}
}

func (m secretModel) ToObject() *corev1.Secret {
	dataMap := make(map[string]string)
	for k, v := range m.Data {
		dataMap[k] = v.ValueString()
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(item types.String) string { return item.ValueString() }),
		},
		Description: m.Description.ValueString(),
		Data:        dataMap,
	}
}
