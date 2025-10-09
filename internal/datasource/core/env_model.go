package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvVarV1Model is the environment variable model.
type EnvVarV1Model struct {
	Name      types.String       `tfsdk:"name"`
	Value     types.String       `tfsdk:"value"`
	ValueFrom *EnvVarSourceModel `tfsdk:"value_from"`
}

// NewEnvVarV1Model creates a new model from the API object.
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

// EnvVarSourceModel is the environment variable source model.
type EnvVarSourceModel struct {
	FieldRef         *ObjectFieldSelectorModel   `tfsdk:"field_ref"`
	ConfigFileKeyRef *ConfigFileKeySelectorModel `tfsdk:"config_file_key_ref"`
}

// ObjectFieldSelectorModel is the object field selector model.
type ObjectFieldSelectorModel struct {
	APIVersion types.String `tfsdk:"api_version"`
	FieldPath  types.String `tfsdk:"field_path"`
}

// ConfigFileKeySelectorModel is the config file key selector model.
type ConfigFileKeySelectorModel struct {
	Name types.String `tfsdk:"name"`
}
