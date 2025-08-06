package core

import (
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	kcorev1 "k8s.io/api/core/v1"
)

type EnvVarV1Model struct {
	Name      types.String       `tfsdk:"name"`
	Value     types.String       `tfsdk:"value"`
	ValueFrom *EnvVarSourceModel `tfsdk:"value_from"`
}

func (m EnvVarV1Model) ToObject() corev1.EnvVar {
	envVar := corev1.EnvVar{
		Name:  m.Name.ValueString(),
		Value: m.Value.ValueString(),
	}
	if m.ValueFrom != nil {
		envVar.ValueFrom = &corev1.EnvVarSource{}
		if m.ValueFrom.FieldRef != nil {
			envVar.ValueFrom.FieldRef = &kcorev1.ObjectFieldSelector{
				APIVersion: m.ValueFrom.FieldRef.APIVersion.ValueString(),
				FieldPath:  m.ValueFrom.FieldRef.FieldPath.ValueString(),
			}
		}
		if m.ValueFrom.ConfigFileKeyRef != nil {
			envVar.ValueFrom.ConfigFileKeyRef = &corev1.ConfigFileKeySelector{
				Name: m.ValueFrom.ConfigFileKeyRef.Name.ValueString(),
			}
		}
	}
	return envVar
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
