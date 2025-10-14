package core

import (
	"cmp"

	"agones.dev/agones/pkg/apis"
	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type regionModel struct {
	ID          types.String               `tfsdk:"id"`
	Name        types.String               `tfsdk:"name"`
	Environment types.String               `tfsdk:"environment"`
	Labels      map[string]types.String    `tfsdk:"labels"`
	DisplayName types.String               `tfsdk:"display_name"`
	Description types.String               `tfsdk:"description"`
	Types       map[string]regionTypeModel `tfsdk:"types"`
}

func newRegionModel(obj *corev1.Region) regionModel {
	model := regionModel{
		ID:          types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:        types.StringValue(obj.Name),
		Environment: types.StringValue(obj.Environment),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		DisplayName: types.StringValue(obj.Spec.DisplayName),
		Description: conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		Types:       map[string]regionTypeModel{},
	}
	for _, typ := range obj.Spec.Types {
		template := cmp.Or(typ.Template, &corev1.RegionTemplate{})
		model.Types[typ.Name] = regionTypeModel{
			Locations:  conv.ForEachSliceItem(typ.Locations, types.StringValue),
			Env:        conv.ForEachSliceItem(template.Env, NewEnvVarModel),
			Scheduling: conv.OptionalFunc(string(template.Scheduling), types.StringValue, types.StringNull),
		}
	}
	return model
}

func (m regionModel) ToObject() *corev1.Region {
	obj := &corev1.Region{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
			Labels:      conv.ForEachMapItem(m.Labels, func(item types.String) string { return item.ValueString() }),
		},
		Spec: corev1.RegionSpec{
			DisplayName: m.DisplayName.ValueString(),
			Description: m.Description.ValueString(),
		},
	}
	for name, typ := range m.Types {
		regTyp := corev1.RegionType{
			Name:      name,
			Locations: conv.ForEachSliceItem(typ.Locations, func(v types.String) string { return v.ValueString() }),
		}
		if len(typ.Env) > 0 || conv.IsKnown(typ.Scheduling) {
			regTyp.Template = &corev1.RegionTemplate{
				Env:        conv.ForEachSliceItem(typ.Env, func(v EnvVarModel) corev1.EnvVar { return v.ToObject() }),
				Scheduling: apis.SchedulingStrategy(typ.Scheduling.ValueString()),
			}
		}
		obj.Spec.Types = append(obj.Spec.Types, regTyp)
	}
	return obj
}

type regionTypeModel struct {
	Locations  []types.String `tfsdk:"locations"`
	Env        []EnvVarModel  `tfsdk:"env"`
	Scheduling types.String   `tfsdk:"scheduling"`
}
