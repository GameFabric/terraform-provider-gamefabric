package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type regionModel struct {
	Name        types.String               `tfsdk:"name"`
	Environment types.String               `tfsdk:"environment"`
	Labels      map[string]types.String    `tfsdk:"labels"`
	DisplayName types.String               `tfsdk:"display_name"`
	Description types.String               `tfsdk:"description"`
	Types       map[string]RegionTypeModel `tfsdk:"types"`
}

func newRegionModel(obj *corev1.Region) regionModel {
	model := regionModel{
		Name:        types.StringValue(obj.Name),
		Environment: types.StringValue(obj.Environment),
		Labels:      conv.ForEachMapItem(obj.Labels, func(item string) types.String { return types.StringValue(item) }),
		DisplayName: conv.OptionalFunc(obj.Spec.DisplayName, types.StringValue, types.StringNull),
		Description: conv.OptionalFunc(obj.Spec.Description, types.StringValue, types.StringNull),
		Types:       map[string]RegionTypeModel{},
	}
	for _, typ := range obj.Spec.Types {
		regTypModel := RegionTypeModel{
			Locations:  conv.ForEachSliceItem(typ.Locations, types.StringValue),
			Env:        conv.ForEachSliceItem(typ.Template.Env, NewEnvVarV1Model),
			Scheduling: conv.OptionalFunc(string(typ.Template.Scheduling), types.StringValue, types.StringNull),
		}
		for _, typStatus := range obj.Status.Types {
			if typStatus.Name != typ.Name {
				continue
			}
			regTypModel.CPU = types.StringValue(typStatus.Limit.CPU.String())
			regTypModel.Memory = types.StringValue(typStatus.Limit.Memory.String())
			break
		}
		model.Types[typ.Name] = regTypModel
	}
	return model
}

type RegionTypeModel struct {
	Locations  []types.String  `tfsdk:"locations"`
	Env        []EnvVarV1Model `tfsdk:"env"`
	Scheduling types.String    `tfsdk:"scheduling"`
	CPU        types.String    `tfsdk:"cpu"`
	Memory     types.String    `tfsdk:"memory"`
}
