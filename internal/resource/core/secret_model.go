package core

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type secretModel struct {
	ID          types.String            `tfsdk:"id"`
	Name        types.String            `tfsdk:"name"`
	Environment types.String            `tfsdk:"environment"`
	Labels      map[string]types.String `tfsdk:"labels"`
	Annotations map[string]types.String `tfsdk:"annotations"`
	Description types.String            `tfsdk:"description"`
	Data        map[string]types.String `tfsdk:"data"`
	DataWO      map[string]types.String `tfsdk:"data_wo"`
}

func newSecretModel(obj *corev1.Secret) secretModel {
	return secretModel{
		ID:          types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:        types.StringValue(obj.Name),
		Environment: types.StringValue(obj.Environment),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		Annotations: conv.ForEachMapItem(obj.Annotations, func(item string) types.String { return types.StringValue(item) }),
		Description: conv.OptionalFunc(obj.Description, types.StringValue, types.StringNull),
		Data:        conv.ForEachMapItem(obj.Data, func(v string) types.String { return types.StringValue(v) }),
		DataWO:      nil,
	}
}

func (m secretModel) ToObject() *corev1.Secret {
	o := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
			Annotations: conv.ForEachMapItem(m.Annotations, func(item types.String) string { return item.ValueString() }),
		},
		Description: m.Description.ValueString(),
		Data:        conv.ForEachMapItem(chooseData(m), func(item types.String) string { return item.ValueString() }),
	}

	fmt.Printf("DEBUG: obj = %s", spew.Sdump(o))

	return o
}

func chooseData(m secretModel) map[string]types.String {
	if len(m.DataWO) > 0 {
		return m.DataWO
	}
	return m.Data
}
