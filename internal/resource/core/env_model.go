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
			model.ValueFrom.FieldPath = types.StringValue(obj.ValueFrom.FieldRef.FieldPath)
		case obj.ValueFrom.ConfigFileKeyRef != nil:
			model.ValueFrom.ConfigFile = types.StringValue(obj.ValueFrom.ConfigFileKeyRef.Name)
		case obj.ValueFrom.SecretKeyRef != nil:
			model.ValueFrom.Secret = &SecretType{
				Key:  types.StringValue(obj.ValueFrom.SecretKeyRef.Key),
				Name: types.StringValue(obj.ValueFrom.SecretKeyRef.Name),
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
		case !m.ValueFrom.FieldPath.IsNull():
			obj.ValueFrom.FieldRef = &kcorev1.ObjectFieldSelector{
				FieldPath: m.ValueFrom.FieldPath.ValueString(),
			}
		case !m.ValueFrom.ConfigFile.IsNull():
			obj.ValueFrom.ConfigFileKeyRef = &corev1.ConfigFileKeySelector{
				Name: m.ValueFrom.ConfigFile.ValueString(),
			}
		case m.ValueFrom.Secret != nil && !m.ValueFrom.Secret.Key.IsNull() && !m.ValueFrom.Secret.Name.IsNull():
			obj.ValueFrom.SecretKeyRef = &corev1.SecretKeySelector{
				Key:  m.ValueFrom.Secret.Key.ValueString(),
				Name: m.ValueFrom.Secret.Name.ValueString(),
			}
		}
	}
	return obj
}

// EnvVarSourceModel represents a source for the value of an EnvVar.
type EnvVarSourceModel struct {
	FieldPath  types.String `tfsdk:"field_path"`
	ConfigFile types.String `tfsdk:"config_file"`
	Secret     *SecretType  `tfsdk:"secret"`
}

// SecretType represents a reference to a secret key.
type SecretType struct {
	Key  types.String `tfsdk:"key"`
	Name types.String `tfsdk:"name"`
}
