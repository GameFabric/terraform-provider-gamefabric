package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type regionModel struct {
	Name        types.String               `tfsdk:"name"`
	Environment types.String               `tfsdk:"environment"`
	DisplayName types.String               `tfsdk:"display_name"`
	Types       map[string]RegionTypeModel `tfsdk:"types"`
}

func newRegionModel(obj *corev1.Region) regionModel {
	model := regionModel{
		Name:        conv.OptionalFunc(obj.Name, types.StringValue, types.StringNull),
		Environment: types.StringValue(obj.Environment),
		DisplayName: conv.OptionalFunc(obj.Spec.DisplayName, types.StringValue, types.StringNull),
		Types:       map[string]RegionTypeModel{},
	}
	for _, typ := range obj.Spec.Types {
		var regTypModel RegionTypeModel
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

// RegionTypeModel is the model for a region type.
type RegionTypeModel struct {
	CPU    types.String `tfsdk:"cpu"`
	Memory types.String `tfsdk:"memory"`
}
