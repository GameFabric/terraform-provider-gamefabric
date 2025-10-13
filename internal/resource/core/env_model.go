package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
	kcorev1 "k8s.io/api/core/v1"
)

// EnvVarModel is the environment variable model.
type EnvVarModel struct {
	Name      types.String       `tfsdk:"name"`
	Value     types.String       `tfsdk:"value"`
	ValueFrom *EnvVarSourceModel `tfsdk:"value_from"`
}

// NewEnvVarModel creates a new model from the API object.
func NewEnvVarModel(obj corev1.EnvVar) EnvVarModel {
	model := EnvVarModel{
		Name:  types.StringValue(obj.Name),
		Value: conv.OptionalFunc(obj.Value, types.StringValue, types.StringNull),
	}
	if obj.ValueFrom != nil {
		model.ValueFrom = &EnvVarSourceModel{}
		switch {
		case obj.ValueFrom.FieldRef != nil:
			model.ValueFrom.FieldRef = &ObjectFieldSelectorModel{
				APIVersion: types.StringValue(obj.ValueFrom.FieldRef.APIVersion),
				FieldPath:  types.StringValue(obj.ValueFrom.FieldRef.FieldPath),
			}
		case obj.ValueFrom.ConfigFileKeyRef != nil:
			model.ValueFrom.ConfigFileKeyRef = &ConfigFileKeySelectorModel{
				Name: types.StringValue(obj.ValueFrom.ConfigFileKeyRef.Name),
			}
		}
	}
	return model
}

// ToObject converts the model to an API object.
func (m EnvVarModel) ToObject() corev1.EnvVar {
	obj := corev1.EnvVar{
		Name:  m.Name.ValueString(),
		Value: m.Value.ValueString(),
	}
	if m.ValueFrom != nil {
		obj.ValueFrom = &corev1.EnvVarSource{}
		switch {
		case m.ValueFrom.FieldRef != nil:
			obj.ValueFrom.FieldRef = &kcorev1.ObjectFieldSelector{
				APIVersion: m.ValueFrom.FieldRef.APIVersion.ValueString(),
				FieldPath:  m.ValueFrom.FieldRef.FieldPath.ValueString(),
			}
		case m.ValueFrom.ConfigFileKeyRef != nil:
			obj.ValueFrom.ConfigFileKeyRef = &corev1.ConfigFileKeySelector{
				Name: m.ValueFrom.ConfigFileKeyRef.Name.ValueString(),
			}
		}
	}
	return obj
}

// EnvVarSourceModel represents a source for the value of an EnvVar.
type EnvVarSourceModel struct {
	FieldRef         *ObjectFieldSelectorModel   `tfsdk:"field_ref"`
	ConfigFileKeyRef *ConfigFileKeySelectorModel `tfsdk:"config_file_key_ref"`
}

// ObjectFieldSelectorModel selects a field of an object.
type ObjectFieldSelectorModel struct {
	APIVersion types.String `tfsdk:"api_version"`
	FieldPath  types.String `tfsdk:"field_path"`
}

// ConfigFileKeySelectorModel selects a key of a ConfigFile.
type ConfigFileKeySelectorModel struct {
	Name types.String `tfsdk:"name"`
}
