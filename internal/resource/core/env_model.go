package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
	kcorev1 "k8s.io/api/core/v1"
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

func (m EnvVarV1Model) ToObject() corev1.EnvVar {
	obj := corev1.EnvVar{
		Name:  m.Name.ValueString(),
		Value: m.Value.ValueString(),
	}
	if m.ValueFrom != nil {
		obj.ValueFrom = &corev1.EnvVarSource{}
		if m.ValueFrom.FieldRef != nil {
			obj.ValueFrom.FieldRef = &kcorev1.ObjectFieldSelector{
				APIVersion: m.ValueFrom.FieldRef.APIVersion.ValueString(),
				FieldPath:  m.ValueFrom.FieldRef.FieldPath.ValueString(),
			}
		}
		if m.ValueFrom.ConfigFileKeyRef != nil {
			obj.ValueFrom.ConfigFileKeyRef = &corev1.ConfigFileKeySelector{
				Name: m.ValueFrom.ConfigFileKeyRef.Name.ValueString(),
			}
		}
	}
	return obj
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
