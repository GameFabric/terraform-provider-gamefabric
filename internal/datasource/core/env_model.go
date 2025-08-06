package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvVarV1Model struct {
	Name      types.String       `tfsdk:"name"`
	Value     types.String       `tfsdk:"value"`
	ValueFrom *EnvVarSourceModel `tfsdk:"value_from"`
}

func NewEnvVarV1Model(obj corev1.EnvVar) EnvVarV1Model {
	model := EnvVarV1Model{
		Name:  types.StringValue(obj.Name),
		Value: conv.OptionalFunc(obj.Value, types.StringValue, types.StringNull),
	}
	if obj.ValueFrom != nil {
		model.ValueFrom = &EnvVarSourceModel{}
		if obj.ValueFrom.FieldRef != nil {
			model.ValueFrom.FieldRef = &ObjectFieldSelectorModel{
				APIVersion: types.StringValue(obj.ValueFrom.FieldRef.APIVersion),
				FieldPath:  types.StringValue(obj.ValueFrom.FieldRef.FieldPath),
			}
		}
		if obj.ValueFrom.ConfigFileKeyRef != nil {
			model.ValueFrom.ConfigFileKeyRef = &ConfigFileKeySelectorModel{
				Name: types.StringValue(obj.ValueFrom.ConfigFileKeyRef.Name),
			}
		}
	}
	return model
}

type EnvVarSourceModel struct {
	FieldRef         *ObjectFieldSelectorModel   `tfsdk:"field_ref"`
	ConfigFileKeyRef *ConfigFileKeySelectorModel `tfsdk:"config_file_key_ref"`
}

type ObjectFieldSelectorModel struct {
	APIVersion types.String `tfsdk:"api_version"`
	FieldPath  types.String `tfsdk:"field_path"`
}

type ConfigFileKeySelectorModel struct {
	Name types.String `tfsdk:"name"`
}
